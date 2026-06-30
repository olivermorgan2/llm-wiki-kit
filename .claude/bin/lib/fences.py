#!/usr/bin/env python3
"""fences — deterministic marker-fenced zone reading and splicing (ADR-054).

Pure stdlib — no third-party dependency, matching the kit's "no new runtime
dependency" rule. This module is the single shared implementation of the
marker-fence mechanics that several skills used to describe inline in prose
(pause, resume, prepare-issue, claude-issue-executor, workflow-docs). It is
imported directly by Python consumers (e.g. docs-render, the fences-eval
runner) and driven as a CLI by `bin/fence`, which markdown skills shell out to.

The kit uses several marker dialects that differ only in token order:

    state / summary   <!-- state:in-flight:start -->        (prefix:zone:kw)
    workflow-docs     <!-- workflow-docs:start:overview -->  (prefix:kw:zone)
    adr-index         <!-- adr-index:start -->               (prefix:kw, no zone)

A `FenceDialect` captures the prefix and token order so one implementation
serves all of them.

Byte-preservation guarantee
---------------------------
All operations work on `text.split("\\n")` and reassemble with `"\\n".join(...)`,
which is an exact round-trip for any input. A replace/omit only ever touches the
lines between (and, for omit, including) the markers; every byte outside that
span is the original line, never reformatted. This is the load-bearing safety
property — fence corruption is silent and destructive, so it is pinned by the
golden round-trip fixtures under bin/lib/fences-fixtures/ via bin/fences-eval.
"""

import json
import os
import re
import sys
import tempfile
from collections import namedtuple

# --------------------------------------------------------------------------
# Dialects
# --------------------------------------------------------------------------

ZONE_CHARS = r"[A-Za-z0-9_-]+"

ZoneSpan = namedtuple("ZoneSpan", ["zone", "start_idx", "end_idx"])


class FenceError(Exception):
    """Base class for fence problems. .exit_code maps to the CLI convention."""
    exit_code = 1


class ZoneNotFound(FenceError):
    exit_code = 1


class MalformedFence(FenceError):
    exit_code = 1

    def __init__(self, reason, lineno=None):
        self.reason = reason
        self.lineno = lineno
        super().__init__(reason if lineno is None else "%s (line %d)" % (reason, lineno))


class DuplicateZone(MalformedFence):
    pass


class FenceDialect:
    """A marker syntax. `order` is one of:
        'zkw'    prefix:zone:keyword   (state, summary)
        'kwz'    prefix:keyword:zone   (workflow-docs)
        'kwonly' prefix:keyword        (adr-index; single anonymous zone "")
    """

    def __init__(self, prefix, order="zkw"):
        self.prefix = prefix
        self.order = order

    def _inner(self, zone, kw):
        if self.order == "zkw":
            return "%s:%s:%s" % (self.prefix, zone, kw)
        if self.order == "kwz":
            return "%s:%s:%s" % (self.prefix, kw, zone)
        return "%s:%s" % (self.prefix, kw)  # kwonly

    def marker(self, zone, kw):
        """The canonical marker line, written flush-left with single spaces."""
        return "<!-- %s -->" % self._inner(zone, kw)

    def _line_re(self, kw):
        """A regex matching a marker line for `kw`, capturing the zone name.

        Tolerant of leading/trailing whitespace and flexible internal spacing
        on read; writes always use the canonical `marker()` form.
        """
        if self.order == "zkw":
            inner = r"%s:(?P<zone>%s):%s" % (re.escape(self.prefix), ZONE_CHARS, kw)
        elif self.order == "kwz":
            inner = r"%s:%s:(?P<zone>%s)" % (re.escape(self.prefix), kw, ZONE_CHARS)
        else:  # kwonly
            inner = r"%s:%s" % (re.escape(self.prefix), kw)
        return re.compile(r"^\s*<!--\s*%s\s*-->\s*$" % inner)

    def start_re(self):
        return self._line_re("start")

    def end_re(self):
        return self._line_re("end")


STATE = FenceDialect("state", "zkw")
SUMMARY = FenceDialect("summary", "zkw")
WDOCS = FenceDialect("workflow-docs", "kwz")
ADR_INDEX = FenceDialect("adr-index", "kwonly")

DIALECTS = {
    "state": STATE,
    "summary": SUMMARY,
    "workflow-docs": WDOCS,
    "adr-index": ADR_INDEX,
}


# --------------------------------------------------------------------------
# Core operations
# --------------------------------------------------------------------------

def _zone_of(match, dialect):
    if dialect.order == "kwonly":
        return ""
    return match.group("zone")


def list_zones(text, dialect):
    """Return the [ZoneSpan] in document order, validating fence integrity.

    Raises MalformedFence on an unclosed start, an end without a matching open
    start, or crossed/nested markers; DuplicateZone if a zone name is opened
    twice.
    """
    lines = text.split("\n")
    start_re = dialect.start_re()
    end_re = dialect.end_re()
    spans = []
    open_zone = None
    open_idx = None
    seen = set()
    for idx, line in enumerate(lines):
        ms = start_re.match(line)
        me = end_re.match(line)
        if ms:
            if open_zone is not None:
                raise MalformedFence(
                    "start marker for %r before %r was closed"
                    % (_zone_of(ms, dialect), open_zone), idx + 1)
            open_zone = _zone_of(ms, dialect)
            open_idx = idx
        elif me:
            ezone = _zone_of(me, dialect)
            if open_zone is None:
                raise MalformedFence("end marker for %r without a start" % ezone, idx + 1)
            if ezone != open_zone:
                raise MalformedFence(
                    "end marker %r does not match open start %r" % (ezone, open_zone), idx + 1)
            if open_zone in seen:
                raise DuplicateZone("zone %r appears more than once" % open_zone, idx + 1)
            seen.add(open_zone)
            spans.append(ZoneSpan(open_zone, open_idx, idx))
            open_zone = None
            open_idx = None
    if open_zone is not None:
        raise MalformedFence("start marker for %r is never closed" % open_zone, open_idx + 1)
    return spans


def _find(text, dialect, zone):
    for span in list_zones(text, dialect):
        if span.zone == zone:
            return span
    raise ZoneNotFound("zone %r not found" % zone)


def read_zone(text, dialect, zone):
    """Return the body between (exclusive of) the markers for `zone`."""
    lines = text.split("\n")
    span = _find(text, dialect, zone)
    return "\n".join(lines[span.start_idx + 1:span.end_idx])


def replace_zone(text, dialect, zone, new_body):
    """Replace only the body between the markers for `zone`.

    Markers and all bytes outside the body are preserved exactly. Idempotent:
    replace_zone(t, d, z, read_zone(t, d, z)) == t for any non-degenerate body.
    An empty `new_body` produces adjacent markers (no body line).
    """
    lines = text.split("\n")
    span = _find(text, dialect, zone)
    body_lines = [] if new_body == "" else new_body.split("\n")
    result = lines[:span.start_idx + 1] + body_lines + lines[span.end_idx:]
    return "\n".join(result)


def upsert_zone(text, dialect, zone, new_body, after_zone=None):
    """Replace `zone` if present, else insert a fresh fenced block.

    Insertion goes immediately after `after_zone`'s end marker when given,
    otherwise at end of file. Used by docs-render to add a section that does
    not yet exist in a previously-rendered file.
    """
    try:
        return replace_zone(text, dialect, zone, new_body)
    except ZoneNotFound:
        pass
    block = [dialect.marker(zone, "start")]
    if new_body != "":
        block += new_body.split("\n")
    block += [dialect.marker(zone, "end")]
    lines = text.split("\n")
    if after_zone is not None:
        anchor = _find(text, dialect, after_zone)
        insert_at = anchor.end_idx + 1
    else:
        insert_at = len(lines)
    result = lines[:insert_at] + block + lines[insert_at:]
    return "\n".join(result)


def omit_section(text, dialect, zone):
    """Remove the whole fenced span for `zone`, markers included.

    For the workflow-docs dialect the start marker precedes the section
    heading, so removing [start_idx .. end_idx] inclusive drops the heading,
    body, and both markers in one move. Callers should follow with
    collapse_blank_lines to avoid a stray gap.
    """
    lines = text.split("\n")
    span = _find(text, dialect, zone)
    result = lines[:span.start_idx] + lines[span.end_idx + 1:]
    return "\n".join(result)


def collapse_blank_lines(text):
    """Collapse any run of more than one blank line to exactly one.

    A line is blank when it is empty or whitespace-only. The file's
    final-newline state is preserved.
    """
    lines = text.split("\n")
    out = []
    prev_blank = False
    for line in lines:
        blank = line.strip() == ""
        if blank and prev_blank:
            continue
        out.append(line)
        prev_blank = blank
    return "\n".join(out)


def marker_health(text, dialect):
    """Return ('ok', []) or ('malformed', [reason]) without raising."""
    try:
        spans = list_zones(text, dialect)
        return "ok", [s.zone for s in spans]
    except FenceError as e:
        return "malformed", [str(e)]


# --------------------------------------------------------------------------
# CLI (driven by bin/fence)
# --------------------------------------------------------------------------

def _atomic_write(path, content):
    d = os.path.dirname(os.path.abspath(path))
    fd, tmp = tempfile.mkstemp(dir=d, prefix=".fence.")
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as fh:
            fh.write(content)
        os.replace(tmp, path)
    except BaseException:
        try:
            os.unlink(tmp)
        except OSError:
            pass
        raise


def _read_file(path):
    with open(path, encoding="utf-8") as fh:
        return fh.read()


def _result(status, outputs, text, errors=None):
    return {"status": status, "outputs": outputs, "errors": errors or [], "text": text}


def _err(code, message):
    return _result("error", {}, message, [{"code": code, "message": message}])


def main(argv):
    """Parse argv, perform the op, print a result JSON object, return exit code.

    Hand-rolled parsing (not argparse) so every path emits the structured
    result JSON that bin/fence renders into the standard envelope.
    """
    if not argv:
        print(json.dumps(_err("usage", "fence: a subcommand is required"))); return 2
    op = argv[0]
    flags = {"file": None, "dialect": None, "zone": None,
             "body": None, "body_file": None, "dry_run": False}
    i = 1
    while i < len(argv):
        a = argv[i]
        if a == "--file":
            flags["file"] = argv[i + 1]; i += 2
        elif a == "--dialect":
            flags["dialect"] = argv[i + 1]; i += 2
        elif a == "--zone":
            flags["zone"] = argv[i + 1]; i += 2
        elif a == "--body":
            flags["body"] = argv[i + 1]; i += 2
        elif a == "--body-file":
            flags["body_file"] = argv[i + 1]; i += 2
        elif a == "--dry-run":
            flags["dry_run"] = True; i += 1
        else:
            print(json.dumps(_err("usage", "fence: unknown argument: %s" % a))); return 2

    if op not in ("read", "replace", "omit", "list", "collapse"):
        print(json.dumps(_err("usage", "fence: unknown subcommand: %s" % op))); return 2
    if not flags["file"]:
        print(json.dumps(_err("usage", "fence: --file is required"))); return 2
    if not os.path.isfile(flags["file"]):
        print(json.dumps(_err("usage", "fence: file not found: %s" % flags["file"]))); return 2

    dialect = None
    if op != "collapse":
        if not flags["dialect"]:
            print(json.dumps(_err("usage", "fence: --dialect is required"))); return 2
        dialect = DIALECTS.get(flags["dialect"])
        if dialect is None:
            print(json.dumps(_err("usage", "fence: --dialect must be one of %s"
                                   % ", ".join(sorted(DIALECTS))))); return 2
    if op in ("read", "replace", "omit") and flags["zone"] is None:
        print(json.dumps(_err("usage", "fence: --zone is required for %s" % op))); return 2

    text = _read_file(flags["file"])

    try:
        if op == "list":
            spans = list_zones(text, dialect)
            zones = [{"name": s.zone, "startLine": s.start_idx + 1,
                      "endLine": s.end_idx + 1} for s in spans]
            names = ", ".join(s.zone for s in spans) or "(none)"
            print(json.dumps(_result("ok", {"file": flags["file"], "zones": zones},
                                     "fence: zones in %s: %s" % (flags["file"], names))))
            return 0

        if op == "read":
            body = read_zone(text, dialect, flags["zone"])
            print(json.dumps(_result("ok", {"file": flags["file"], "zone": flags["zone"],
                                            "body": body}, body)))
            return 0

        if op == "replace":
            if flags["body_file"] is not None:
                new_body = _read_file(flags["body_file"]).rstrip("\n")
            elif flags["body"] == "-":
                new_body = sys.stdin.read().rstrip("\n")
            elif flags["body"] is not None:
                new_body = flags["body"]
            else:
                print(json.dumps(_err("usage",
                      "fence: replace needs --body, --body -, or --body-file"))); return 2
            new_text = replace_zone(text, dialect, flags["zone"], new_body)
            changed = new_text != text
            if not flags["dry_run"] and changed:
                _atomic_write(flags["file"], new_text)
            out = {"file": flags["file"], "zone": flags["zone"], "changed": changed,
                   "dryRun": flags["dry_run"]}
            txt = new_text if flags["dry_run"] else ("fence: replaced zone %r in %s"
                                                     % (flags["zone"], flags["file"]))
            print(json.dumps(_result("ok", out, txt)))
            return 0

        if op == "omit":
            new_text = collapse_blank_lines(omit_section(text, dialect, flags["zone"]))
            changed = new_text != text
            if not flags["dry_run"] and changed:
                _atomic_write(flags["file"], new_text)
            out = {"file": flags["file"], "zone": flags["zone"], "changed": changed,
                   "dryRun": flags["dry_run"]}
            txt = new_text if flags["dry_run"] else ("fence: omitted zone %r from %s"
                                                     % (flags["zone"], flags["file"]))
            print(json.dumps(_result("ok", out, txt)))
            return 0

        if op == "collapse":
            new_text = collapse_blank_lines(text)
            changed = new_text != text
            if not flags["dry_run"] and changed:
                _atomic_write(flags["file"], new_text)
            out = {"file": flags["file"], "changed": changed, "dryRun": flags["dry_run"]}
            txt = new_text if flags["dry_run"] else "fence: collapsed blank lines in %s" % flags["file"]
            print(json.dumps(_result("ok", out, txt)))
            return 0
    except FenceError as e:
        print(json.dumps(_err("malformed-fence", "fence: %s" % e)))
        return e.exit_code

    print(json.dumps(_err("usage", "fence: unhandled op"))); return 2


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
