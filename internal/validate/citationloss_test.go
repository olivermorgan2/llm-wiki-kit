package validate

import (
	"reflect"
	"testing"
)

// evidence wraps body markdown under a single `## Evidence` context heading.
func evidence(body string) []byte {
	return []byte("## Evidence\n\n" + body + "\n")
}

// TestCitationLoss covers the URL/bundle/malformed target classes over a single
// evidence context (ADR-008 sub-decision 6). Each row states source vs staged
// evidence bodies and the expected net-lost targets (source order, first-seen
// spelling). opts default to the single "Evidence" context unless overridden.
func TestCitationLoss(t *testing.T) {
	evi := Options{EvidenceSections: []string{"Evidence"}}
	cases := []struct {
		name   string
		source string
		staged string
		opts   Options
		want   []string
	}{
		{ // (1) drop a cited URL from the evidence context
			name:   "drop cited url is a loss",
			source: "## Evidence\n\n[a](https://example.com/p)\n",
			staged: "## Evidence\n\n(removed)\n",
			opts:   evi,
			want:   []string{"https://example.com/p"},
		},
		{ // (2) move within the same evidence context is not a loss
			name:   "move within same context is nil",
			source: "## Evidence\n\nSee [a](https://example.com/p) here.\n",
			staged: "## Evidence\n\nOther prose.\n\nThen [a](https://example.com/p).\n",
			opts:   evi,
			want:   nil,
		},
		{ // (3) move between two designated evidence contexts is not a net loss (D2 union)
			name:   "move between two evidence contexts is nil (union semantics)",
			source: "## Evidence\n\n[a](notes.md)\n\n## Sources\n\nprose\n",
			staged: "## Evidence\n\nprose\n\n## Sources\n\n[a](notes.md)\n",
			opts:   Options{EvidenceSections: []string{"Evidence", "Sources"}},
			want:   nil,
		},
		{ // (4a) fragment-only spelling change normalizes equal
			name:   "fragment drop normalizes equal is nil",
			source: "## Evidence\n\n[a](a.md#frag)\n",
			staged: "## Evidence\n\n[a](a.md)\n",
			opts:   evi,
			want:   nil,
		},
		{ // (4b) leading-slash spelling change normalizes equal
			name:   "root-absolute vs relative normalizes equal is nil",
			source: "## Evidence\n\n[a](/notes/a.md)\n",
			staged: "## Evidence\n\n[a](notes/a.md)\n",
			opts:   evi,
			want:   nil,
		},
		{ // (5) URLs are compared verbatim: host case differs => loss
			name:   "url host case is not folded so it is a loss",
			source: "## Evidence\n\n[a](https://X.com/p)\n",
			staged: "## Evidence\n\n[a](https://x.com/p)\n",
			opts:   evi,
			want:   []string{"https://X.com/p"},
		},
		{ // (6) dropping one of two duplicate occurrences is not a loss (set semantics)
			name:   "drop duplicate occurrence is nil",
			source: "## Evidence\n\n[a](a.md) and [b](a.md)\n",
			staged: "## Evidence\n\n[a](a.md)\n",
			opts:   evi,
			want:   nil,
		},
		{ // (7) target dropped from evidence but kept as a navigational link is a loss
			name:   "kept only outside evidence is a loss",
			source: "## Evidence\n\n[a](a.md)\n",
			staged: "## Evidence\n\nprose\n\n## Nav\n\n[a](a.md)\n",
			opts:   evi,
			want:   []string{"a.md"},
		},
		{ // (8) all citation classes count: a dropped malformed target is a loss
			name:   "drop malformed mailto target is a loss",
			source: "## Evidence\n\n[a](mailto:x@y.com)\n",
			staged: "## Evidence\n\nprose\n",
			opts:   evi,
			want:   []string{"mailto:x@y.com"},
		},
		{ // (10) empty EvidenceSections is always nil (zero-cost default)
			name:   "empty evidence sections is nil",
			source: "## Evidence\n\n[a](a.md)\n",
			staged: "## Evidence\n\nprose\n",
			opts:   Options{},
			want:   nil,
		},
		{ // (11) an image link in evidence is never a citation
			name:   "drop image link is nil",
			source: "## Evidence\n\n![a](img.png)\n",
			staged: "## Evidence\n\nprose\n",
			opts:   evi,
			want:   nil,
		},
		{ // (12) lost targets keep source order (not sorted)
			name:   "lost order follows source order",
			source: "## Evidence\n\n[a](https://b.com/1) [b](https://a.com/2)\n",
			staged: "## Evidence\n\nprose\n",
			opts:   evi,
			want:   []string{"https://b.com/1", "https://a.com/2"},
		},
		{ // first-seen spelling: two occurrences normalize equal, drop both
			name:   "first-seen spelling is reported",
			source: "## Evidence\n\n[a](a.md#one) then [b](a.md#two)\n",
			staged: "## Evidence\n\nprose\n",
			opts:   evi,
			want:   []string{"a.md#one"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := CitationLoss([]byte(tc.source), []byte(tc.staged), tc.opts)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("CitationLoss = %v, want %v", got, tc.want)
			}
		})
	}
}

// (9) A bundle-escaping `../` target resolves through the repo-path class when an
// anchor is present: dropping it is a loss, keeping it is not, and moving/keeping
// it is keyed on the repo-relative path.
func TestCitationLossRepoClassWithAnchor(t *testing.T) {
	opts := Options{
		EvidenceSections: []string{"Evidence"},
		BundleDir:        "wiki",
		RepoResolve:      repoStub(map[string]RepoStatus{"shared/doc.md": RepoFound}),
	}
	// Unchanged repo-class citation: no loss.
	if got := CitationLoss(evidence("[a](../shared/doc.md#s)"), evidence("[a](../shared/doc.md#s)"), opts); got != nil {
		t.Errorf("unchanged repo citation loss = %v, want nil", got)
	}
	// Dropped repo-class citation: a loss carrying its first-seen spelling.
	if got := CitationLoss(evidence("[a](../shared/doc.md#s)"), evidence("prose"), opts); !reflect.DeepEqual(got, []string{"../shared/doc.md#s"}) {
		t.Errorf("dropped repo citation loss = %v, want [../shared/doc.md#s]", got)
	}
}

// (9 cont.) Without a repoResolve anchor a bundle-escaping target degrades to the
// malformed class consistently on both sides, so an unchanged one is still nil.
func TestCitationLossRepoClassNoAnchorConsistent(t *testing.T) {
	opts := Options{EvidenceSections: []string{"Evidence"}}
	if got := CitationLoss(evidence("[a](../shared/doc.md)"), evidence("[a](../shared/doc.md)"), opts); got != nil {
		t.Errorf("unchanged malformed-degraded citation loss = %v, want nil", got)
	}
	if got := CitationLoss(evidence("[a](../shared/doc.md)"), evidence("prose"), opts); !reflect.DeepEqual(got, []string{"../shared/doc.md"}) {
		t.Errorf("dropped malformed-degraded citation loss = %v, want [../shared/doc.md]", got)
	}
}
