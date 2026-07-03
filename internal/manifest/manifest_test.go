package manifest

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"testing"
	"time"

	"github.com/olivermorgan2/llm-wiki-kit/internal/profile"
	"github.com/olivermorgan2/llm-wiki-kit/internal/scaffold"
	"github.com/olivermorgan2/llm-wiki-kit/internal/txn"
)

// fixedClock is pinned so scaffold.Plan's date-stamped templates are stable
// within a test run. The scaffold embeds the date into two pages, so golden
// hash literals would rot across days — every assertion re-derives hashes from
// the same in-memory change slice instead.
var fixedClock = time.Date(2026, time.July, 3, 12, 0, 0, 0, time.UTC)

// coreVersions is a representative Versions input built from the shipped core
// profile, mirroring what runInstall assembles.
func coreVersions(t *testing.T) Versions {
	t.Helper()
	p, err := profile.Resolve(profile.CoreID)
	if err != nil {
		t.Fatalf("resolve core profile: %v", err)
	}
	return Versions{
		Plugin:  "0.1.0-dev",
		CLI:     "0.1.0-dev",
		OKF:     scaffold.OKFVersion,
		Profile: ProfileRecord{ID: p.ID, Version: p.Version},
	}
}

func planChanges(t *testing.T) []txn.FileChange {
	t.Helper()
	p, err := profile.Resolve(profile.CoreID)
	if err != nil {
		t.Fatalf("resolve core profile: %v", err)
	}
	return scaffold.Plan(p, fixedClock)
}

func TestBuildRecordsEveryChangeAsPluginOwned(t *testing.T) {
	changes := planChanges(t)
	m := Build(coreVersions(t), changes)

	if len(m.Assets) != len(changes) {
		t.Fatalf("assets = %d, want %d", len(m.Assets), len(changes))
	}
	for _, a := range m.Assets {
		if a.Class != ClassPluginOwned {
			t.Errorf("asset %q class = %q, want %q", a.Path, a.Class, ClassPluginOwned)
		}
		if a.Path == Path {
			t.Errorf("manifest must not list itself as an asset: %q", a.Path)
		}
	}
	if !sort.SliceIsSorted(m.Assets, func(i, j int) bool { return m.Assets[i].Path < m.Assets[j].Path }) {
		t.Errorf("assets are not sorted by path: %+v", m.Assets)
	}
}

func TestBuildHashEqualsLastInstalledHashEqualsSHA256(t *testing.T) {
	changes := planChanges(t)
	m := Build(coreVersions(t), changes)

	byPath := map[string]txn.FileChange{}
	for _, c := range changes {
		byPath[c.Target] = c
	}
	for _, a := range m.Assets {
		sum := sha256.Sum256(byPath[a.Path].Data)
		want := hex.EncodeToString(sum[:])
		if a.Hash != want {
			t.Errorf("asset %q hash = %q, want %q", a.Path, a.Hash, want)
		}
		if a.LastInstalledHash != want {
			t.Errorf("asset %q lastInstalledHash = %q, want %q", a.Path, a.LastInstalledHash, want)
		}
	}
}

func TestBuildRecordsVersions(t *testing.T) {
	v := coreVersions(t)
	m := Build(v, planChanges(t))

	if m.SchemaVersion != SchemaVersion {
		t.Errorf("schemaVersion = %q, want %q", m.SchemaVersion, SchemaVersion)
	}
	if m.Plugin != v.Plugin {
		t.Errorf("plugin = %q, want %q", m.Plugin, v.Plugin)
	}
	if m.CLI != v.CLI {
		t.Errorf("cli = %q, want %q", m.CLI, v.CLI)
	}
	if m.OKF != v.OKF {
		t.Errorf("okf = %q, want %q", m.OKF, v.OKF)
	}
	if m.Profile != v.Profile {
		t.Errorf("profile = %+v, want %+v", m.Profile, v.Profile)
	}
}

func TestMarshalIsDeterministic(t *testing.T) {
	m := Build(coreVersions(t), planChanges(t))
	a, err := m.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	b, err := m.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !bytes.Equal(a, b) {
		t.Errorf("marshal is not deterministic:\n%s\nvs\n%s", a, b)
	}
	if len(a) == 0 || a[len(a)-1] != '\n' {
		t.Errorf("marshal must end in a trailing newline")
	}
}

func TestParseRoundTrip(t *testing.T) {
	m := Build(coreVersions(t), planChanges(t))
	data, err := m.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got, err := Parse(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	round, err := got.Marshal()
	if err != nil {
		t.Fatalf("re-marshal: %v", err)
	}
	if !bytes.Equal(data, round) {
		t.Errorf("round-trip differs:\n%s\nvs\n%s", data, round)
	}
}

func TestChangeTargetsManifestPath(t *testing.T) {
	m := Build(coreVersions(t), planChanges(t))
	c, err := m.Change()
	if err != nil {
		t.Fatalf("change: %v", err)
	}
	if c.Target != Path {
		t.Errorf("change target = %q, want %q", c.Target, Path)
	}
	if c.Mode.Perm() != 0o644 {
		t.Errorf("change mode = %v, want 0644", c.Mode)
	}
	data, err := m.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !bytes.Equal(c.Data, data) {
		t.Errorf("change data does not equal marshaled manifest")
	}
}

func TestBuildWithFixedClockIsReproducible(t *testing.T) {
	a := Build(coreVersions(t), planChanges(t))
	b := Build(coreVersions(t), planChanges(t))
	am, err := a.Marshal()
	if err != nil {
		t.Fatalf("marshal a: %v", err)
	}
	bm, err := b.Marshal()
	if err != nil {
		t.Fatalf("marshal b: %v", err)
	}
	if !bytes.Equal(am, bm) {
		t.Errorf("build under a fixed clock is not reproducible")
	}
}
