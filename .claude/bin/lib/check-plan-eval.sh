#!/usr/bin/env bash
# check-plan-eval.sh — shared eval functions for bin/check-plan and the
# /check-plan slash-command surface (skills/check-plan/SKILL.md).
#
# Sourced by bin/check-plan. Each eval_*_criteria function takes a
# file path and emits one TSV record per criterion to stdout, with
# fields:
#
#   <id>\t<severity>\t<status>\t<message>\t<remediation>
#
# - severity is "deterministic" or "warning"
# - status   is "pass" | "fail" | "warn"
# - message  and remediation are free-text (avoid embedded TABs)
#
# The criteria list is canonical in skills/check-plan/criteria.md;
# this file is the machine-executable encoding of those rules.
# When you add a new criterion, update both files in lockstep.
#
# Per ADR-043 (the design decision) and ADR-034 (the upstream
# /check-plan skill).

# ---- helpers -----------------------------------------------------------

# emit one TSV record. Args: id severity status message remediation
_cp_emit() {
  printf '%s\t%s\t%s\t%s\t%s\n' "$1" "$2" "$3" "$4" "$5"
}

# Find the line number of the first line matching the given pattern.
# Pattern is a basic regex (anchored with ^ and $ explicitly by the caller).
# Returns empty string if no match.
_cp_lineno() {
  local file="$1" pattern="$2"
  grep -n -E "$pattern" "$file" 2>/dev/null | head -1 | cut -d: -f1
}

# Print lines in [start, end) of the file. start is 1-indexed inclusive,
# end is 1-indexed exclusive. If end is empty, print to EOF.
_cp_slice() {
  local file="$1" start="$2" end="$3"
  if [ -z "$start" ]; then
    return 0
  fi
  if [ -z "$end" ] || [ "$end" -le "$start" ]; then
    sed -n "${start},\$p" "$file"
  else
    local last=$((end - 1))
    sed -n "${start},${last}p" "$file"
  fi
}

# ---- ADR criteria ------------------------------------------------------
# Validates ADRs against templates/adr-template.md. See
# skills/check-plan/criteria.md §"ADR criteria" for the canonical list.

eval_adr_criteria() {
  local file="$1"

  # ADR-C1: ## Context, ## Decision, ## Consequences each present and non-empty.
  local ctx_line dec_line con_line missing
  ctx_line=$(_cp_lineno "$file" '^## Context$')
  dec_line=$(_cp_lineno "$file" '^## Decision$')
  con_line=$(_cp_lineno "$file" '^## Consequences$')
  missing=""
  [ -z "$ctx_line" ] && missing="$missing ## Context"
  [ -z "$dec_line" ] && missing="$missing ## Decision"
  [ -z "$con_line" ] && missing="$missing ## Consequences"
  if [ -n "$missing" ]; then
    _cp_emit "ADR-C1" "deterministic" "fail" \
      "Missing section(s):${missing}" \
      "Add the missing section(s); template orders them Context, Options considered, Decision, Consequences."
  else
    # Verify each has at least one non-blank, non-heading line of body.
    local empty_sections="" sec_name sec_start sec_end
    for sec_name in Context Decision Consequences; do
      sec_start=$(_cp_lineno "$file" "^## ${sec_name}\$")
      sec_end=$(awk -v start="$sec_start" 'NR>start && /^## /{print NR; exit}' "$file")
      local body
      body=$(_cp_slice "$file" "$((sec_start + 1))" "$sec_end" | grep -v -E '^\s*$' || true)
      if [ -z "$body" ]; then
        empty_sections="$empty_sections ## ${sec_name}"
      fi
    done
    if [ -n "$empty_sections" ]; then
      _cp_emit "ADR-C1" "deterministic" "fail" \
        "Section(s) present but empty:${empty_sections}" \
        "Each section needs at least one paragraph of body."
    else
      _cp_emit "ADR-C1" "deterministic" "pass" \
        "Context, Decision, Consequences sections present and non-empty" ""
    fi
  fi

  # ADR-C2: ## Options considered has at least 2 ### Option subheadings.
  local opt_line opt_end opt_count
  opt_line=$(_cp_lineno "$file" '^## Options considered$')
  if [ -z "$opt_line" ]; then
    _cp_emit "ADR-C2" "deterministic" "fail" \
      "Options considered section is missing" \
      "Add ## Options considered with at least two ### Option blocks, or remove if the decision is genuinely uncontested."
  else
    opt_end=$(awk -v start="$opt_line" 'NR>start && /^## /{print NR; exit}' "$file")
    opt_count=$(_cp_slice "$file" "$((opt_line + 1))" "$opt_end" | grep -c -E '^### Option ' || true)
    if [ "$opt_count" -ge 2 ]; then
      _cp_emit "ADR-C2" "deterministic" "pass" \
        "Options considered has $opt_count Option blocks" ""
    else
      _cp_emit "ADR-C2" "deterministic" "fail" \
        "Options considered has only $opt_count Option block(s); need at least 2" \
        "Add a second option, or remove the section if the decision is genuinely uncontested."
    fi
  fi

  # ADR-C3: each ### Option block has Pros: and Cons: lines.
  if [ -n "$opt_line" ]; then
    local opt_starts opt_block_end pros_count cons_count missing_pc
    missing_pc=""
    # Each option's start-line; option_end is the next ### Option or ## (section change) or EOF.
    opt_starts=$(awk -v from="$opt_line" 'NR>from && /^### Option /{print NR}' "$file")
    if [ -n "$opt_starts" ]; then
      while IFS= read -r start; do
        opt_block_end=$(awk -v from="$start" 'NR>from && (/^### /||/^## /){print NR; exit}' "$file")
        local block
        block=$(_cp_slice "$file" "$((start + 1))" "$opt_block_end")
        pros_count=$(printf '%s\n' "$block" | grep -c -E '(^|^- )Pros:' || true)
        cons_count=$(printf '%s\n' "$block" | grep -c -E '(^|^- )Cons:' || true)
        if [ "$pros_count" -eq 0 ] || [ "$cons_count" -eq 0 ]; then
          local hdr
          hdr=$(sed -n "${start}p" "$file")
          missing_pc="$missing_pc | ${hdr}"
        fi
      done <<<"$opt_starts"
    fi
    if [ -z "$missing_pc" ]; then
      _cp_emit "ADR-C3" "deterministic" "pass" \
        "Every Option block has Pros: and Cons: lines" ""
    else
      _cp_emit "ADR-C3" "deterministic" "fail" \
        "Option block(s) missing Pros: or Cons: line: ${missing_pc#" | "}" \
        "Add the missing line and at least one bullet for each affected option."
    fi
  fi

  # ADR-C4: ## Decision body names one of the option labels declared in
  # ## Options considered (e.g. "Adopt **Option B**").
  if [ -n "$dec_line" ] && [ -n "$opt_line" ]; then
    local dec_end dec_body labels found_label label
    dec_end=$(awk -v start="$dec_line" 'NR>start && /^## /{print NR; exit}' "$file")
    dec_body=$(_cp_slice "$file" "$((dec_line + 1))" "$dec_end")
    # Extract option labels: take the part after "### Option " up to ":" or end-of-line.
    labels=$(awk -v from="$opt_line" -v to="$opt_end" '
      NR>from && (to=="" || NR<to) && /^### Option / {
        s=$0
        sub(/^### Option /, "", s)
        sub(/:.*/, "", s)
        sub(/[ \t]+$/, "", s)
        print s
      }' "$file")
    found_label=""
    if [ -n "$labels" ]; then
      while IFS= read -r label; do
        [ -z "$label" ] && continue
        if printf '%s' "$dec_body" | grep -q -E "Option[[:space:]*]+${label}\b"; then
          found_label="$label"
          break
        fi
      done <<<"$labels"
    fi
    if [ -n "$found_label" ]; then
      _cp_emit "ADR-C4" "deterministic" "pass" \
        "Decision names Option ${found_label}" ""
    else
      _cp_emit "ADR-C4" "deterministic" "fail" \
        "Decision body does not name any of the listed option labels" \
        "Either name the chosen option explicitly (e.g. \"Adopt Option B\") or rename it to match the Options-considered headings."
    fi
  fi

  # ADR-C5 (warning): every ADR-NNN token resolves to a file in design/adr/.
  local tokens unresolved=""
  tokens=$(grep -o -E '\bADR-[0-9]+' "$file" 2>/dev/null | sort -u || true)
  if [ -n "$tokens" ]; then
    local tok num matches
    while IFS= read -r tok; do
      num="${tok#ADR-}"
      # Normalise to 3-digit zero-padded for the glob, but also try the
      # raw form in case the file uses a different width.
      local padded
      padded=$(printf '%03d' "$((10#$num))" 2>/dev/null || echo "$num")
      matches=$(ls design/adr/adr-"${padded}"-*.md 2>/dev/null | head -1 || true)
      if [ -z "$matches" ]; then
        matches=$(ls design/adr/adr-"${num}"-*.md 2>/dev/null | head -1 || true)
      fi
      if [ -z "$matches" ]; then
        unresolved="$unresolved $tok"
      fi
    done <<<"$tokens"
  fi
  if [ -z "$unresolved" ]; then
    _cp_emit "ADR-C5" "warning" "pass" \
      "All ADR-NNN tokens resolve to files in design/adr/" ""
  else
    _cp_emit "ADR-C5" "warning" "warn" \
      "Unresolved ADR token(s):${unresolved}" \
      "Either fix the token (typo) or create the missing ADR. Warning only — sometimes ADRs reference ones not yet drafted."
  fi

  # ADR-C6 (warning): semantic-conflict check is best-effort substring per
  # ADR-034 and not deterministically encodable. Report as a deferred
  # warning so the row is operator-visible.
  _cp_emit "ADR-C6" "warning" "warn" \
    "Semantic-conflict check deferred (best-effort substring per ADR-034 not implemented in v1)" \
    "Reviewer should manually check the Decision against accepted ADRs in the same area; revisit when a structural rule lands."
}

# ---- Prompt criteria ---------------------------------------------------
# Validates issue prompts against prompts/_template.md. See
# skills/check-plan/criteria.md §"Prompt criteria".

eval_prompt_criteria() {
  local file="$1"

  # PROMPT-C6 first (which sections exist) so PROMPT-C1 can rely on it.
  # The ten required sections per the template (section heading text is
  # bare, not "## "-prefixed — this is the prompt format, not markdown
  # ADR format).
  local req_sections=(
    "Context:"
    "ADR:"
    "GitHub Issue:"
    "Goal"
    "Why it matters"
    "Requirements"
    "Acceptance criteria"
    "Scope and constraints"
    "Evaluation & testing requirements"
    "Instructions for you"
  )
  local missing_sections="" sec
  for sec in "${req_sections[@]}"; do
    if ! grep -q -F -x "$sec" "$file"; then
      missing_sections="$missing_sections | ${sec}"
    fi
  done
  if [ -z "$missing_sections" ]; then
    _cp_emit "PROMPT-C6" "deterministic" "pass" \
      "All 10 required sections present" ""
  else
    _cp_emit "PROMPT-C6" "deterministic" "fail" \
      "Missing required section(s): ${missing_sections#" | "}" \
      "Add the missing section(s); re-run /prepare-issue if the prompt was hand-authored and drifted from the template."
  fi

  # PROMPT-C1: Acceptance criteria section present + at least one bullet.
  local ac_line ac_end ac_body bullet_count
  ac_line=$(_cp_lineno "$file" '^Acceptance criteria$')
  if [ -z "$ac_line" ]; then
    _cp_emit "PROMPT-C1" "deterministic" "fail" \
      "Acceptance criteria section missing" \
      "Add an Acceptance criteria section listing observable end-state outcomes (not implementation tasks)."
  else
    ac_end=$(awk -v start="$ac_line" '
      NR>start && (/^Scope and constraints$/ || /^Evaluation & testing requirements$/ \
                   || /^Instructions for you$/ || /^Why it matters$/) {print NR; exit}' "$file")
    ac_body=$(_cp_slice "$file" "$((ac_line + 1))" "$ac_end")
    bullet_count=$(printf '%s\n' "$ac_body" | grep -c -E '^- ' || true)
    if [ "$bullet_count" -ge 1 ]; then
      _cp_emit "PROMPT-C1" "deterministic" "pass" \
        "Acceptance criteria section has $bullet_count bullet(s)" ""
    else
      _cp_emit "PROMPT-C1" "deterministic" "fail" \
        "Acceptance criteria section has no bullets" \
        "List at least one observable end-state outcome as a '- ' bullet."
    fi
  fi

  # PROMPT-C2: ADR section present and either resolves to a file or says
  # "ADR: none — <reason>".
  local adr_line adr_end adr_body adr_file_line tok num matches adr_status_message
  adr_line=$(_cp_lineno "$file" '^ADR:$')
  if [ -z "$adr_line" ]; then
    _cp_emit "PROMPT-C2" "deterministic" "fail" \
      "ADR: section missing" \
      "Either link the governing ADR (File: line) or replace the section with 'ADR: none — <reason>'."
  else
    adr_end=$(awk -v start="$adr_line" '
      NR>start && (/^GitHub Issue:$/ || /^Goal$/) {print NR; exit}' "$file")
    adr_body=$(_cp_slice "$file" "$((adr_line + 1))" "$adr_end")
    adr_status_message=""
    # Case A: explicit "ADR: none — <reason>" or "none — <reason>" within the section
    if printf '%s\n' "$adr_body" | grep -q -E '^- *(File:)?\s*none\s*[—-]\s*.+'; then
      adr_status_message="explicit none-with-reason"
    elif printf '%s\n' "$adr_body" | grep -q -E '^none\s*[—-]\s*.+'; then
      adr_status_message="explicit none-with-reason"
    elif grep -q -E '^ADR: none\s*[—-]\s*.+' "$file"; then
      adr_status_message="explicit none-with-reason"
    else
      # Case B: at least one File: line that resolves
      adr_file_line=$(printf '%s\n' "$adr_body" | grep -E '^- File: ' | head -1 || true)
      if [ -n "$adr_file_line" ]; then
        tok=$(printf '%s' "$adr_file_line" | grep -o -E 'adr-[0-9]+-[A-Za-z0-9_-]*\.md' | head -1 || true)
        if [ -n "$tok" ]; then
          if [ -f "design/adr/$tok" ]; then
            adr_status_message="resolves to design/adr/$tok"
          fi
        fi
      fi
    fi
    if [ -n "$adr_status_message" ]; then
      _cp_emit "PROMPT-C2" "deterministic" "pass" \
        "ADR section ${adr_status_message}" ""
    else
      _cp_emit "PROMPT-C2" "deterministic" "fail" \
        "ADR section does not resolve to a file in design/adr/ and does not explicitly say 'ADR: none — <reason>'" \
        "Either fix the File: line, or replace the section with 'ADR: none — <reason>'."
    fi
  fi

  # PROMPT-C3: no remaining {{...}} placeholders or unresolved TODO markers.
  local placeholders todos
  placeholders=$(grep -c -E '\{\{[^}]+\}\}' "$file" 2>/dev/null || true)
  todos=$(grep -c -E '<!--[[:space:]]*TODO: fill in[[:space:]]*-->' "$file" 2>/dev/null || true)
  if [ "$placeholders" -eq 0 ] && [ "$todos" -eq 0 ]; then
    _cp_emit "PROMPT-C3" "deterministic" "pass" \
      "No unfilled placeholders or TODO markers" ""
  else
    local detail=""
    [ "$placeholders" -gt 0 ] && detail="${detail}${placeholders} {{...}} placeholder(s)"
    [ "$todos" -gt 0 ] && {
      [ -n "$detail" ] && detail="${detail}, "
      detail="${detail}${todos} TODO marker(s)"
    }
    _cp_emit "PROMPT-C3" "deterministic" "fail" \
      "${detail} remain" \
      "List the unfilled slots with line numbers; fill or explicitly delete the line if the slot is irrelevant."
  fi

  # PROMPT-C4 (warning): scope fits a phase. Skip silently when no
  # build-out-plan or single-phase. The check is a coarse heuristic;
  # warn but never fail.
  local plan="design/build-out-plan.md" phase_count
  if [ -f "$plan" ]; then
    phase_count=$(grep -c -E '^## Phase ' "$plan" || true)
    if [ "$phase_count" -ge 2 ]; then
      _cp_emit "PROMPT-C4" "warning" "warn" \
        "Phase-fit check is heuristic and out of scope for this v1 implementation" \
        "Reviewer should sanity-check that the prompt's scope fits within one named phase; revisit when a structural rule lands."
    fi
    # else: single-phase or zero — skip silently per criteria.md
  fi
  # else: no plan — skip silently per criteria.md

  # PROMPT-C5 (warning): single-PR scope soft cap (Requirements + Scope ≤ ~25
  # bullets total, ≤ ~80 lines).
  local req_line req_end req_body scope_line scope_end scope_body total_bullets total_lines
  req_line=$(_cp_lineno "$file" '^Requirements$')
  scope_line=$(_cp_lineno "$file" '^Scope and constraints$')
  if [ -n "$req_line" ] && [ -n "$scope_line" ]; then
    req_end=$(awk -v start="$req_line" '
      NR>start && (/^Acceptance criteria$/ || /^Scope and constraints$/ \
                   || /^Evaluation & testing requirements$/) {print NR; exit}' "$file")
    req_body=$(_cp_slice "$file" "$((req_line + 1))" "$req_end")
    scope_end=$(awk -v start="$scope_line" '
      NR>start && (/^Evaluation & testing requirements$/ \
                   || /^Instructions for you$/) {print NR; exit}' "$file")
    scope_body=$(_cp_slice "$file" "$((scope_line + 1))" "$scope_end")
    total_bullets=$(( $(printf '%s\n' "$req_body" | grep -c -E '^- ' || true) \
                    + $(printf '%s\n' "$scope_body" | grep -c -E '^- ' || true) ))
    total_lines=$(( $(printf '%s' "$req_body" | wc -l) + $(printf '%s' "$scope_body" | wc -l) ))
    if [ "$total_bullets" -le 25 ] && [ "$total_lines" -le 80 ]; then
      _cp_emit "PROMPT-C5" "warning" "pass" \
        "Scope cap respected ($total_bullets bullets, $total_lines lines)" ""
    else
      _cp_emit "PROMPT-C5" "warning" "warn" \
        "Scope exceeds soft cap ($total_bullets bullets, $total_lines lines; cap is 25/80)" \
        "Either split the issue, or accept the warning if the work is genuinely cohesive."
    fi
  else
    _cp_emit "PROMPT-C5" "warning" "warn" \
      "Cannot apply scope cap — Requirements or Scope section missing" \
      "Add the missing section(s) per PROMPT-C6's remediation."
  fi
}
