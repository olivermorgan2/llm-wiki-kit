package validate

import (
	"path"
	"regexp"
	"strings"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
	"github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"
)

// Rule codes. These are the stable identifiers carried by each finding and used
// to build its fingerprint; the concrete set is fixed for this issue (ADR-004).
const (
	codeOKFYAMLParse    = "okf-yaml-parse"
	codeOKFTypePresent  = "okf-type-present"
	codeCoreReqTitle    = "core-required-title"
	codeCoreReqDesc     = "core-required-description"
	codeCoreFieldType   = "core-field-type"
	codeCoreRecommended = "core-recommended-missing"
	codeCoreKebab       = "core-kebab-filename"
)

// evaluatePage runs the OKF and core-profile rules over one page and returns
// findings at their core-default severity (the first layer of the ADR-004
// precedence). Parsing is a hard precondition: if the frontmatter block is
// missing/unterminated or the YAML fails to parse, the only finding is the
// never-suppressible okf-yaml-parse error and no other rule runs for the page
// (a finding cannot be computed over unparsed content).
func evaluatePage(yaml yamladapter.Adapter, path string, content []byte) []contract.Finding {
	fm, _, err := splitFrontmatter(content)
	if err != nil {
		return []contract.Finding{parseFailure(path, err)}
	}
	var m map[string]any
	if err := yaml.Unmarshal(fm, &m); err != nil {
		return []contract.Finding{parseFailure(path, err)}
	}
	if m == nil {
		m = map[string]any{}
	}

	var out []contract.Finding
	out = append(out, okfRules(path, m)...)
	out = append(out, profileRules(path, m)...)
	return out
}

// parseFailure builds the okf-yaml-parse finding for a page whose frontmatter is
// missing, unterminated, or unparseable.
func parseFailure(path string, err error) contract.Finding {
	return contract.Finding{
		Ruleset:  contract.RulesetOKF,
		Severity: contract.SeverityError,
		Code:     codeOKFYAMLParse,
		Message:  "frontmatter YAML does not parse: " + err.Error(),
		Path:     path,
	}
}

// okfRules evaluates base OKF conformance over parsed frontmatter. The only OKF
// content rule in this issue is okf-type-present; unknown type *values* are
// accepted (per-type rules are a later slice).
func okfRules(path string, m map[string]any) []contract.Finding {
	var out []contract.Finding
	if !hasNonEmptyString(m, "type") {
		out = append(out, contract.Finding{
			Ruleset:  contract.RulesetOKF,
			Severity: contract.SeverityError,
			Code:     codeOKFTypePresent,
			Message:  "every concept document must have a non-empty `type`",
			Path:     path,
		})
	}
	return out
}

// recommendedFields are the core-profile fields whose absence is advisory
// (addendum 003). Order is fixed for deterministic messages.
var recommendedFields = []string{"timestamp", "tags", "aliases", "resource"}

// kebabName matches a kebab-case markdown filename: lowercase alphanumeric words
// joined by single hyphens, ending in `.md`.
var kebabName = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*\.md$`)

// profileRules evaluates core-profile conformance over parsed frontmatter and
// the page filename. Every finding is tagged RulesetProfile so it renders
// separately from OKF findings (criterion 5). Findings that could recur per
// field (wrong types, missing recommended fields) are aggregated into a single
// finding per rule so each has a unique {ruleset, code, path} fingerprint.
func profileRules(pagePath string, m map[string]any) []contract.Finding {
	var out []contract.Finding

	if missingRequiredString(m, "title") {
		out = append(out, profileFinding(contract.SeverityError, codeCoreReqTitle,
			"core profile requires a non-empty `title`", pagePath))
	}
	if missingRequiredString(m, "description") {
		out = append(out, profileFinding(contract.SeverityError, codeCoreReqDesc,
			"core profile requires a non-empty `description`", pagePath))
	}

	if mistyped := mistypedFields(m); len(mistyped) > 0 {
		out = append(out, profileFinding(contract.SeverityError, codeCoreFieldType,
			"wrong YAML type for field(s): "+strings.Join(mistyped, ", "), pagePath))
	}

	if missing := missingRecommended(m); len(missing) > 0 {
		out = append(out, profileFinding(contract.SeveritySuggestion, codeCoreRecommended,
			"recommended field(s) absent: "+strings.Join(missing, ", "), pagePath))
	}

	if !kebabName.MatchString(path.Base(pagePath)) {
		out = append(out, profileFinding(contract.SeverityWarning, codeCoreKebab,
			"filename should be kebab-case ([a-z0-9] words joined by hyphens)", pagePath))
	}

	return out
}

// profileFinding builds a profile-ruleset finding.
func profileFinding(sev contract.Severity, code, msg, path string) contract.Finding {
	return contract.Finding{
		Ruleset:  contract.RulesetProfile,
		Severity: sev,
		Code:     code,
		Message:  msg,
		Path:     path,
	}
}

// missingRequiredString reports whether a required string field is absent or an
// empty/whitespace string. A present-but-wrong-type value returns false: that is
// the field-type rule's responsibility, so one problem yields one finding.
func missingRequiredString(m map[string]any, key string) bool {
	v, ok := m[key]
	if !ok {
		return true
	}
	s, isStr := v.(string)
	if !isStr {
		return false
	}
	return strings.TrimSpace(s) == ""
}

// mistypedFields returns, in a fixed order, the modeled fields whose present
// value has the wrong YAML type: title/description/type must be scalar strings;
// tags/aliases must be sequences. Absent fields are skipped.
func mistypedFields(m map[string]any) []string {
	var bad []string
	for _, k := range []string{"type", "title", "description"} {
		if v, ok := m[k]; ok {
			if _, isStr := v.(string); !isStr {
				bad = append(bad, k)
			}
		}
	}
	for _, k := range []string{"tags", "aliases"} {
		if v, ok := m[k]; ok {
			if _, isSeq := v.([]any); !isSeq {
				bad = append(bad, k)
			}
		}
	}
	return bad
}

// missingRecommended returns the recommended fields absent from the frontmatter,
// in fixed order.
func missingRecommended(m map[string]any) []string {
	var missing []string
	for _, k := range recommendedFields {
		if _, ok := m[k]; !ok {
			missing = append(missing, k)
		}
	}
	return missing
}

// hasNonEmptyString reports whether key holds a non-empty, non-whitespace string
// value. A missing key, a null, an empty/whitespace string, or a non-string
// value all return false. Wrong-type values are left to the field-type rule.
func hasNonEmptyString(m map[string]any, key string) bool {
	v, ok := m[key]
	if !ok {
		return false
	}
	s, isStr := v.(string)
	if !isStr {
		return false
	}
	return strings.TrimSpace(s) != ""
}
