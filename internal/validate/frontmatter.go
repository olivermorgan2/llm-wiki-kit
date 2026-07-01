package validate

import (
	"fmt"
	"strings"
)

// frontmatterDelim is the fenced-block delimiter that opens and closes the YAML
// frontmatter of an OKF page.
const frontmatterDelim = "---"

// splitFrontmatter separates a leading `---`-fenced YAML frontmatter block from
// the markdown body. It returns the raw YAML bytes (delimiters removed) and the
// body. A file that does not open with a `---` line, or whose frontmatter block
// is never closed, is a structural failure and returns an error — the caller
// reports it as the never-suppressible okf-yaml-parse finding (ADR-004).
func splitFrontmatter(data []byte) (yaml []byte, body []byte, err error) {
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 || strings.TrimRight(lines[0], "\r") != frontmatterDelim {
		return nil, nil, fmt.Errorf("missing frontmatter: file must begin with a %q line", frontmatterDelim)
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], "\r") == frontmatterDelim {
			yaml = []byte(strings.Join(lines[1:i], "\n"))
			body = []byte(strings.Join(lines[i+1:], "\n"))
			return yaml, body, nil
		}
	}
	return nil, nil, fmt.Errorf("unterminated frontmatter: no closing %q line", frontmatterDelim)
}
