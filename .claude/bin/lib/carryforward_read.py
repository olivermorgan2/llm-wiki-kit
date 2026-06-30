#!/usr/bin/env python3
"""carryforward_read — parse design-questions carry-forward blocks (ADR-054).

The cross-skill carry-forward unit (ADR-040) is a `### design-questions`
heading followed by a fenced YAML block, under `## Follow-ups` of
notes/eval-issue-NNN.md. Two scripts read it: bin/validate-carry-forward
(strict conformance against schemas/design-questions.v1.yaml) and bin/pr-context
(surface the entries to render `## Notes for #M` sections). They share this one
dependency-free parser so the grammar is implemented once.

  find_blocks(text)        -> [(start_line_1based, block_lines)]
  parse_block(block_lines) -> (entries, parse_errors)
      Each entry is a dict of field->value plus an internal "__line__".

Both functions are preserved verbatim from validate-carry-forward's previous
inline definitions, so extracting them changes no behavior.
"""

import re


# A block is the body of a fenced ```yaml ... ``` that immediately follows
# a `### design-questions` heading (blank lines allowed between).
def find_blocks(text):
    lines = text.split("\n")
    blocks = []
    i, n = 0, len(lines)
    while i < n:
        if re.match(r"^###\s+design-questions\s*$", lines[i]):
            j = i + 1
            while j < n and lines[j].strip() == "":
                j += 1
            if j < n and re.match(r"^```", lines[j]):
                start = j + 1
                k = start
                while k < n and not re.match(r"^```\s*$", lines[k]):
                    k += 1
                blocks.append((start + 1, lines[start:k]))  # 1-based start line
                i = k + 1
                continue
        i += 1
    return blocks


# ---- parse the constrained block grammar (no YAML dependency) ----------
# Grammar (docs/workflow-guide.md §6):
#   - title: <scalar>
#     target-issue: "#N"
#     context: |
#       <one paragraph, possibly multi-line>
# Returns (entries, parse_errors). Each entry is a dict of field->value.
def parse_block(block_lines):
    entries, errs = [], []
    # base indent = indent of the first "- " item line
    base = None
    items = []  # list of (start_idx, lines_for_item)
    cur = None
    for idx, raw in enumerate(block_lines):
        if "\t" in raw:
            errs.append("line %d: tab character (use spaces)" % (idx + 1))
        if raw.strip() == "":
            if cur is not None:
                cur[1].append(raw)
            continue
        indent = len(raw) - len(raw.lstrip(" "))
        stripped = raw.lstrip(" ")
        if stripped.startswith("- "):
            if base is None:
                base = indent
            if indent == base:
                if cur is not None:
                    items.append(cur)
                cur = (idx + 1, [raw])
                continue
        if cur is not None:
            cur[1].append(raw)
        else:
            errs.append("line %d: content before first list item" % (idx + 1))
    if cur is not None:
        items.append(cur)

    for start_line, ilines in items:
        entry = {}
        # normalise: drop the "- " from the first line, then every key
        # sits at item-indent+2.
        norm = []
        for n_i, raw in enumerate(ilines):
            if n_i == 0:
                pre = raw[:base] + "  "
                norm.append(pre + raw.lstrip(" ")[2:])
            else:
                norm.append(raw)
        key_indent = base + 2
        p = 0
        m = len(norm)
        while p < m:
            raw = norm[p]
            if raw.strip() == "":
                p += 1; continue
            indent = len(raw) - len(raw.lstrip(" "))
            stripped = raw.lstrip(" ")
            km = re.match(r"^([A-Za-z][\w-]*):\s*(.*)$", stripped)
            if indent == key_indent and km:
                key, val = km.group(1), km.group(2)
                if val in ("|", "|-", ">", ">-"):
                    # literal/folded block: gather deeper-indented lines
                    body = []
                    p += 1
                    while p < m:
                        bl = norm[p]
                        if bl.strip() == "":
                            body.append("")
                            p += 1
                            continue
                        bindent = len(bl) - len(bl.lstrip(" "))
                        if bindent > key_indent:
                            body.append(bl[key_indent + 2:] if len(bl) > key_indent + 2 else bl.strip())
                            p += 1
                        else:
                            break
                    # trim trailing blanks
                    while body and body[-1] == "":
                        body.pop()
                    sep = " " if val in (">", ">-") else "\n"
                    entry[key] = sep.join(body).strip()
                else:
                    entry[key] = val.strip().strip("\"\x27")
                    p += 1
            else:
                errs.append("line %d: unexpected line in entry (not a key at the right indent): %r"
                            % (start_line + p, stripped[:50]))
                p += 1
        entry["__line__"] = start_line
        entries.append(entry)
    return entries, errs
