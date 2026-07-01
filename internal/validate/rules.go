package validate

import (
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

	return okfRules(path, m)
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
