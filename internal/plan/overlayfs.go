package plan

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

// overlayFS presents os.DirFS(root) with one proposed page layered over it at a
// bundle-relative target path. It lets the shared validate engine run over a
// draft that is not on disk — the content-aware `page inspect` path (issue #38):
// the draft is served at its target, injected into its parent directory's
// listing, and any of its ancestor directories that do not yet exist on disk are
// synthesized. Everything else reads through to the real bundle unchanged, so the
// engine sees the draft as a member of the full bundle file set (criterion 15:
// broken-link and citation membership see the draft) while the live tree is never
// mutated.
//
// It implements fs.FS plus fs.StatFS, fs.ReadDirFS, and fs.ReadFileFS so the
// engine's fs.WalkDir / fs.ReadFile calls dispatch to the overlay directly rather
// than falling back to Open.
type overlayFS struct {
	base    fs.FS  // os.DirFS(root)
	target  string // slash-form bundle-relative path of the proposed page
	content []byte // proposed page bytes served at target
	// synth is the set of the target's ancestor directories (slash form, excluding
	// ".") that do not exist on base and must be synthesized so the walk descends
	// to the draft.
	synth map[string]bool
}

// newOverlayFS builds an overlay of os.DirFS(root) serving content at the
// slash-form bundle-relative target. It probes the target's ancestor directories
// on disk and records the ones missing so ReadDir/Stat can synthesize them.
func newOverlayFS(root, target string, content []byte) *overlayFS {
	o := &overlayFS{
		base:    os.DirFS(root),
		target:  target,
		content: content,
		synth:   map[string]bool{},
	}
	for dir := path.Dir(target); dir != "." && dir != "/" && dir != ""; dir = path.Dir(dir) {
		if _, err := fs.Stat(o.base, dir); err != nil {
			o.synth[dir] = true
		}
	}
	return o
}

// childUnder returns the next path element of target directly under directory
// name (name == "." for the bundle root), and whether that element is the target
// file itself (isFile) rather than an intermediate directory. ok is false when
// name is not an ancestor of target.
func (o *overlayFS) childUnder(name string) (child string, isFile, ok bool) {
	var rest string
	if name == "." {
		rest = o.target
	} else {
		prefix := name + "/"
		if !strings.HasPrefix(o.target, prefix) {
			return "", false, false
		}
		rest = strings.TrimPrefix(o.target, prefix)
	}
	if rest == "" {
		return "", false, false
	}
	if i := strings.IndexByte(rest, '/'); i >= 0 {
		return rest[:i], false, true
	}
	return rest, true, true
}

// Stat resolves the overlay's synthesized entries (the target file, the missing
// ancestor directories) and reads through to the base for everything else.
func (o *overlayFS) Stat(name string) (fs.FileInfo, error) {
	if name == o.target {
		return overlayInfo{name: path.Base(name), size: int64(len(o.content))}, nil
	}
	if o.synth[name] {
		return overlayInfo{name: path.Base(name), dir: true}, nil
	}
	return fs.Stat(o.base, name)
}

// ReadDir lists directory name with the overlay applied: a synthesized directory
// yields only its single overlay child; a real directory yields its on-disk
// entries plus the overlay child when this directory is the target's parent and
// the child is not already present. Entries are returned in lexical order.
func (o *overlayFS) ReadDir(name string) ([]fs.DirEntry, error) {
	var entries []fs.DirEntry
	if o.synth[name] {
		// A synthesized directory has no on-disk backing; its only content is the
		// overlay path descending through it.
		if child, isFile, ok := o.childUnder(name); ok {
			entries = append(entries, overlayEntry{name: child, dir: !isFile})
		}
		return entries, nil
	}

	base, err := fs.ReadDir(o.base, name)
	if err != nil {
		return nil, err
	}
	entries = append(entries, base...)

	if child, isFile, ok := o.childUnder(name); ok {
		present := false
		for _, e := range base {
			if e.Name() == child {
				present = true
				break
			}
		}
		// Inject only a genuinely new child: a draft over an existing page (the
		// file already listed) needs no injection — ReadFile serves its content.
		if !present {
			entries = append(entries, overlayEntry{name: child, dir: !isFile})
		}
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	return entries, nil
}

// ReadFile serves the proposed bytes at the target and reads through otherwise.
func (o *overlayFS) ReadFile(name string) ([]byte, error) {
	if name == o.target {
		return append([]byte(nil), o.content...), nil
	}
	return fs.ReadFile(o.base, name)
}

// Open backs the fs.FS contract. The engine reaches the overlay through the
// Stat/ReadDir/ReadFile fast paths, so Open only needs to serve the target file
// (as an in-memory file) and read through for everything else; synthesized
// directories are never Opened directly.
func (o *overlayFS) Open(name string) (fs.File, error) {
	if name == o.target {
		return &overlayFile{
			info:   overlayInfo{name: path.Base(name), size: int64(len(o.content))},
			reader: bytes.NewReader(o.content),
		}, nil
	}
	return o.base.Open(name)
}

// overlayEntry is a synthesized fs.DirEntry for an injected file or directory.
type overlayEntry struct {
	name string
	dir  bool
}

func (e overlayEntry) Name() string { return e.name }
func (e overlayEntry) IsDir() bool  { return e.dir }
func (e overlayEntry) Type() fs.FileMode {
	if e.dir {
		return fs.ModeDir
	}
	return 0
}
func (e overlayEntry) Info() (fs.FileInfo, error) {
	return overlayInfo{name: e.name, dir: e.dir}, nil
}

// overlayInfo is a synthesized fs.FileInfo for the target file or a synthesized
// directory. Size is meaningful only for the file; ModTime is the zero time,
// which the validate engine never inspects.
type overlayInfo struct {
	name string
	size int64
	dir  bool
}

func (i overlayInfo) Name() string { return i.name }
func (i overlayInfo) Size() int64  { return i.size }
func (i overlayInfo) Mode() fs.FileMode {
	if i.dir {
		return fs.ModeDir | 0o755
	}
	return 0o644
}
func (i overlayInfo) ModTime() time.Time { return time.Time{} }
func (i overlayInfo) IsDir() bool        { return i.dir }
func (i overlayInfo) Sys() any           { return nil }

// overlayFile is the in-memory fs.File Open returns for the target page.
type overlayFile struct {
	info   overlayInfo
	reader *bytes.Reader
}

func (f *overlayFile) Stat() (fs.FileInfo, error) { return f.info, nil }
func (f *overlayFile) Read(p []byte) (int, error) { return f.reader.Read(p) }
func (f *overlayFile) Close() error               { return nil }

var _ fs.FS = (*overlayFS)(nil)
var _ fs.StatFS = (*overlayFS)(nil)
var _ fs.ReadDirFS = (*overlayFS)(nil)
var _ fs.ReadFileFS = (*overlayFS)(nil)
var _ io.Reader = (*overlayFile)(nil)
