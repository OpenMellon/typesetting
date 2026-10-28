package shaping

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/OpenMellon/typesetting/di"
	"github.com/OpenMellon/typesetting/language"
	"golang.org/x/image/math/fixed"
)

func TestWrapBuffer(t *testing.T) {
	t.Run("new and reset have same state", func(t *testing.T) {
		b1 := wrapBuffer{}
		b1.reset()
		b2 := wrapBuffer{}
		b2.reset()
		defer func() {
			if t.Failed() {
				t.Logf("b1: %s\nb2: %s\n", b1.stats(), b2.stats())
			}
		}()
		b2.reset()
		if !reflect.DeepEqual(b1, b2) {
			t.Errorf("expected new and new+reset buffer to have same fields")
		}
	})
	t.Run("paragraph functions", func(t *testing.T) {
		b1 := wrapBuffer{}
		b1.reset()
		defer func() {
			if t.Failed() {
				t.Logf("b1: %s\n", b1.stats())
			}
		}()
		line := Line{Output{Advance: 10}}
		for i := 0; i < 5; i++ {
			maxLines := cap(b1.paragraph)
			startAddr := &b1.paragraph[0:1][0]
			b1.reset()
			for k := 0; k < maxLines; k++ {
				b1.paragraphAppend(line)
			}
			para := b1.finalParagraph()
			if actAddr := &para[0:1][0]; startAddr != actAddr {
				t.Errorf("expected paragraph to reuse slice starting at %p, got %p", startAddr, actAddr)
			}
		}
		for i := 0; i < 5; i++ {
			maxLines := cap(b1.paragraph)
			startAddr := &b1.paragraph[0:1][0]
			b1.reset()
			for k := 0; k < maxLines+1; k++ {
				b1.paragraphAppend(line)
			}
			para := b1.finalParagraph()
			if actAddr := &para[0:1][0]; startAddr == actAddr {
				t.Errorf("expected paragraph to enlarge slice starting at %p (changing start addres), but got %p", startAddr, actAddr)
			}
		}
		for i := 0; i < 5; i++ {
			startAddr := &b1.paragraph[0:1][0]
			b1.reset()
			para := b1.singleRunParagraph(line[0])
			if actAddr := &para[0:1][0]; startAddr != actAddr {
				t.Errorf("expected singleRunParagraph to reuse slice starting at %p, got %p", startAddr, actAddr)
			}
		}
	})
	t.Run("line building", func(t *testing.T) {
		b := wrapBuffer{}
		b.reset()
		defer func() {
			if t.Failed() {
				t.Logf("b: %s\n", b.stats())
			}
		}()
		b.startLine()
		run := Output{Advance: 10}
		lineLen := 0
		for i := 0; i < cap(b.line)-1; i++ {
			b.candidateAppend(run)
			lineLen++
			if b.hasBest() {
				t.Errorf("no best committed, but hasBest() true")
			}
		}
		best := b.finalizeBest()
		if best != nil {
			t.Errorf("no best committed, but finalizeBest() returned non-nil %#+v", best)
		}
		preCommitAltLen := len(b.alt)
		b.markCandidateBest(run)
		if postCommitAltLen := len(b.alt); preCommitAltLen != postCommitAltLen {
			t.Errorf("modified candidate when committing best, expected len %d, got %d", preCommitAltLen, postCommitAltLen)
		}
		if !b.hasBest() {
			t.Errorf("best committed, but hasBest() false")
		}
		preLineUsed := b.lineUsed
		best = b.finalizeBest()
		postLineUsed := b.lineUsed
		if len(best) != lineLen+1 {
			t.Errorf("expected best candidate to have len %d, got %d", lineLen+1, len(best))
		}
		if used := postLineUsed - preLineUsed; used != len(best) {
			t.Errorf("expected best candidate to use %d capacity of line, used %d", len(best), used)
		}
		if b.lineExhausted {
			t.Errorf("did not expect line to be exhausted yet")
		}
		b.startLine()
		b.candidateAppend(run)
		b.candidateAppend(run)
		if b.hasBest() {
			t.Errorf("no best committed, but hasBest() true")
		}
		b.markCandidateBest()
		if !b.hasBest() {
			t.Errorf("best committed, but hasBest() false")
		}
		preLineUsed = b.lineUsed
		best = b.finalizeBest()
		postLineUsed = b.lineUsed
		if len(best) != 2 {
			t.Errorf("expected best candidate to have len %d, got %d", 2, len(best))
		}
		if used := postLineUsed - preLineUsed; used != 0 {
			t.Errorf("expected best candidate to use %d capacity of line, used %d", 0, used)
		}
		if !b.lineExhausted {
			t.Errorf("expected line to be exhausted")
		}
		preResetCap := cap(b.line)
		b.reset()
		postResetCap := cap(b.line)
		if postResetCap <= preResetCap {
			t.Errorf("expected exhausted line to expand on reset")
		}
		if b.lineExhausted {
			t.Errorf("expected lineExhausted to clear on reset")
		}
	})
}

func TestLineWrapperBreakPolicies(t *testing.T) {
	type testcase struct {
		name      string
		paragraph []rune
	}
	for _, tc := range []testcase{
		{
			name:      "hello world",
			paragraph: []rune("hello, world"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := HarfbuzzShaper{}
			out := s.Shape(Input{
				Text:      tc.paragraph,
				RunStart:  0,
				RunEnd:    len(tc.paragraph),
				Direction: di.DirectionLTR,
				Face:      benchEnFace,
				Size:      fixed.I(16),
				Script:    language.Latin,
				Language:  language.NewLanguage("EN"),
			})
			textWidth := out.Advance.Ceil()
			for maxWidth := textWidth; maxWidth > 0; maxWidth -= max(maxWidth/10, 1) {
				t.Run(fmt.Sprintf("maxWidth%d", maxWidth), func(t *testing.T) {
					for _, policy := range []LineBreakPolicy{WhenNecessary, Never, Always} {
						t.Run(policy.String(), func(t *testing.T) {
							w := LineWrapper{}
							lines, truncated := w.WrapParagraph(WrapConfig{
								BreakPolicy: policy,
							}, maxWidth, tc.paragraph, NewSliceIterator([]Output{out}))
							checkRuneCounts(t, tc.paragraph, lines, truncated)
						})
					}
				})
			}
		})
	}
}

// Checks that grapheme break in ligature are properly
// avoided
func TestGraphemeBreakLigature(t *testing.T) {
	const text1 = "il est affable"
	gls := append(
		glyphs(0, 7),        // simple 1:1 mapping
		ligatureGlyph(8, 2), // ligature
		simpleGlyph(10),     // a
		simpleGlyph(11),     // b
		simpleGlyph(12),     // l
		simpleGlyph(13),     // e
	)
	shapedText := Output{
		Advance: fixed.I(10 * len(gls)),
		Runes:   Range{Count: len([]rune(text1))},
		Glyphs:  gls,
		LineBounds: Bounds{
			Ascent:  fixed.I(10),
			Descent: fixed.I(5),
			// No line gap.
		},
		GlyphBounds: Bounds{
			Ascent: fixed.I(10),
			// No glyphs descend.
		},
	}
	var w LineWrapper
	lines, _ := w.WrapParagraph(WrapConfig{BreakPolicy: Always}, 90, []rune(text1), NewSliceIterator([]Output{shapedText}))
	if L := lines[0][0].Runes.Count; L != 10 {
		t.Errorf("invalid break with ligature %d", L)
	}
}

func (out Output) copy() Output {
	copy := out
	copy.Glyphs = append([]Glyph(nil), out.Glyphs...)
	return copy
}

func TestLineWrapperBreakSpecific(t *testing.T) {
	type expectation struct {
		name           string
		config         WrapConfig
		runeCounts     []int
		truncatedCount int
	}
	type testcase struct {
		expectations []expectation
		// This truncator will be inserted automatically into each expectation's WrapConfig
		// to avoid redundantly specifying it.
		truncator []rune
		maxWidth  int
		paragraph []rune
	}
	for _, tc := range []testcase{
		{
			paragraph: []rune("hello, world"),
			maxWidth:  80,
			expectations: []expectation{
				{
					name:       "WhenNecessary",
					config:     WrapConfig{BreakPolicy: WhenNecessary},
					runeCounts: []int{7, 5},
				},
				{
					name:           "WhenNecessary-Truncate",
					config:         WrapConfig{BreakPolicy: WhenNecessary, TruncateAfterLines: 1},
					runeCounts:     []int{11},
					truncatedCount: 1,
				},
				{
					name:       "Always",
					config:     WrapConfig{BreakPolicy: Always},
					runeCounts: []int{11, 1},
				},
				{
					name:           "Always-Truncate",
					config:         WrapConfig{BreakPolicy: Always, TruncateAfterLines: 1},
					runeCounts:     []int{11},
					truncatedCount: 1,
				},
				{
					name:       "Never",
					config:     WrapConfig{BreakPolicy: Never},
					runeCounts: []int{7, 5},
				},
				{
					name:           "Never-Truncate",
					config:         WrapConfig{BreakPolicy: Never, TruncateAfterLines: 1},
					runeCounts:     []int{7},
					truncatedCount: 5,
				},
			},
		},
		{
			paragraph: []rune("hello, world"),
			maxWidth:  80,
			truncator: []rune("…"),
			expectations: []expectation{
				{
					name:       "WhenNecessary",
					config:     WrapConfig{BreakPolicy: WhenNecessary},
					runeCounts: []int{7, 5},
				},
				{
					name:           "WhenNecessary-Truncate",
					config:         WrapConfig{BreakPolicy: WhenNecessary, TruncateAfterLines: 1},
					runeCounts:     []int{8},
					truncatedCount: 4,
				},
				{
					name:       "Always",
					config:     WrapConfig{BreakPolicy: Always},
					runeCounts: []int{11, 1},
				},
				{
					name:           "Always-Truncate",
					config:         WrapConfig{BreakPolicy: Always, TruncateAfterLines: 1},
					runeCounts:     []int{8},
					truncatedCount: 4,
				},
				{
					name:       "Never",
					config:     WrapConfig{BreakPolicy: Never},
					runeCounts: []int{7, 5},
				},
				{
					name:           "Never-Truncate",
					config:         WrapConfig{BreakPolicy: Never, TruncateAfterLines: 1},
					runeCounts:     []int{7},
					truncatedCount: 5,
				},
			},
		},
		{
			paragraph: []rune("hello, world"),
			maxWidth:  60,
			expectations: []expectation{
				{
					name:       "WhenNecessary",
					config:     WrapConfig{BreakPolicy: WhenNecessary},
					runeCounts: []int{7, 5},
				},
				{
					name:           "WhenNecessary-Truncate",
					config:         WrapConfig{BreakPolicy: WhenNecessary, TruncateAfterLines: 1},
					runeCounts:     []int{8},
					truncatedCount: 4,
				},
				{
					name:       "Always",
					config:     WrapConfig{BreakPolicy: Always},
					runeCounts: []int{8, 4},
				},
				{
					name:           "Always-Truncate",
					config:         WrapConfig{BreakPolicy: Always, TruncateAfterLines: 1},
					runeCounts:     []int{8},
					truncatedCount: 4,
				},
				{
					name:       "Never",
					config:     WrapConfig{BreakPolicy: Never},
					runeCounts: []int{7, 5},
				},
				{
					name:           "Never-Truncate",
					config:         WrapConfig{BreakPolicy: Never, TruncateAfterLines: 1},
					runeCounts:     []int{7},
					truncatedCount: 5,
				},
			},
		},
		{
			paragraph: []rune("hello, world"),
			// maxWidth of 56 with this truncator results in ~40px of usable space for the text if the
			// truncator is present. The text "hello, " needs a little more space than this even if the
			// space is shortened to width zero.
			maxWidth:  56,
			truncator: []rune("…"),
			expectations: []expectation{
				{
					// We are forced to wrap within a word here.
					name:           "WhenNecessary-Truncate",
					config:         WrapConfig{BreakPolicy: WhenNecessary, TruncateAfterLines: 1},
					runeCounts:     []int{5},
					truncatedCount: 7,
				},
				{
					// We are forced to wrap within a word here.
					name:           "Always-Truncate",
					config:         WrapConfig{BreakPolicy: Always, TruncateAfterLines: 1},
					runeCounts:     []int{5},
					truncatedCount: 7,
				},
				{
					// Since we can't wrap within words in this configuration, we truncate all runes.
					name:           "Never-Truncate",
					config:         WrapConfig{BreakPolicy: Never, TruncateAfterLines: 1},
					runeCounts:     []int{},
					truncatedCount: 12,
				},
			},
		},
		{
			// The first word "hello, " doesn't quite fit in 40 pixels. These tests exercise this boundary
			// condition with various break policies and truncation policies.
			paragraph: []rune("hello, world"),
			maxWidth:  40,
			expectations: []expectation{
				{
					name:       "WhenNecessary",
					config:     WrapConfig{BreakPolicy: WhenNecessary},
					runeCounts: []int{5, 2, 5},
				},
				{
					name:           "WhenNecessary-Truncate",
					config:         WrapConfig{BreakPolicy: WhenNecessary, TruncateAfterLines: 1},
					runeCounts:     []int{5},
					truncatedCount: 7,
				},
				{
					name:       "Always",
					config:     WrapConfig{BreakPolicy: Always},
					runeCounts: []int{5, 6, 1},
				},
				{
					name:           "Always-Truncate",
					config:         WrapConfig{BreakPolicy: Always, TruncateAfterLines: 1},
					runeCounts:     []int{5},
					truncatedCount: 7,
				},
				{
					name:       "Never",
					config:     WrapConfig{BreakPolicy: Never},
					runeCounts: []int{7, 5},
				},
				{
					name:           "Never-Truncate",
					config:         WrapConfig{BreakPolicy: Never, TruncateAfterLines: 1},
					runeCounts:     []int{},
					truncatedCount: 12,
				},
			},
		},
		{
			paragraph: []rune("hello, world"),
			maxWidth:  33,
			expectations: []expectation{
				{
					name:       "WhenNecessary",
					config:     WrapConfig{BreakPolicy: WhenNecessary},
					runeCounts: []int{4, 3, 4, 1},
				},
				{
					name:           "WhenNecessary-Truncate",
					config:         WrapConfig{BreakPolicy: WhenNecessary, TruncateAfterLines: 1},
					runeCounts:     []int{4},
					truncatedCount: 8,
				},
				{
					name:           "WhenNecessary-Truncate2",
					config:         WrapConfig{BreakPolicy: WhenNecessary, TruncateAfterLines: 2},
					runeCounts:     []int{4, 4},
					truncatedCount: 4,
				},
				{
					name:       "Always",
					config:     WrapConfig{BreakPolicy: Always},
					runeCounts: []int{4, 4, 4},
				},
				{
					name:           "Always-Truncate",
					config:         WrapConfig{BreakPolicy: Always, TruncateAfterLines: 1},
					runeCounts:     []int{4},
					truncatedCount: 8,
				},
				{
					name:           "Always-Truncate2",
					config:         WrapConfig{BreakPolicy: Always, TruncateAfterLines: 2},
					runeCounts:     []int{4, 4},
					truncatedCount: 4,
				},
				{
					name:       "Never",
					config:     WrapConfig{BreakPolicy: Never},
					runeCounts: []int{7, 5},
				},
				{
					name:           "Never-Truncate",
					config:         WrapConfig{BreakPolicy: Never, TruncateAfterLines: 1},
					runeCounts:     []int{0},
					truncatedCount: 12,
				},
				{
					name:           "Never-Truncate2",
					config:         WrapConfig{BreakPolicy: Never, TruncateAfterLines: 2},
					runeCounts:     []int{7},
					truncatedCount: 5,
				},
			},
		},
		{
			paragraph: []rune("hello, world"),
			maxWidth:  9,
			expectations: []expectation{
				{
					name:       "WhenNecessary",
					config:     WrapConfig{BreakPolicy: WhenNecessary},
					runeCounts: []int{1, 1, 2, 1, 2, 1, 1, 1, 1, 1},
				},
				{
					name:       "Always",
					config:     WrapConfig{BreakPolicy: Always},
					runeCounts: []int{1, 1, 2, 1, 2, 1, 1, 1, 1, 1},
				},
				{
					name:       "Never",
					config:     WrapConfig{BreakPolicy: Never},
					runeCounts: []int{7, 5},
				},
			},
		},
		{
			paragraph: []rune("hello, world"),
			maxWidth:  1,
			expectations: []expectation{
				{
					name:       "WhenNecessary",
					config:     WrapConfig{BreakPolicy: WhenNecessary},
					runeCounts: []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				},
				{
					name:       "Always",
					config:     WrapConfig{BreakPolicy: Always},
					runeCounts: []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				},
				{
					name:       "Never",
					config:     WrapConfig{BreakPolicy: Never},
					runeCounts: []int{7, 5},
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("%q-width%d-truncator%q", string(tc.paragraph), tc.maxWidth, string(tc.truncator)), func(t *testing.T) {
			s := HarfbuzzShaper{}
			out := s.Shape(Input{
				Text:      tc.paragraph,
				RunStart:  0,
				RunEnd:    len(tc.paragraph),
				Direction: di.DirectionLTR,
				Face:      benchEnFace,
				Size:      fixed.I(16),
				Script:    language.Latin,
				Language:  language.NewLanguage("EN"),
			})
			shapedTrunc := s.Shape(Input{
				Text:      tc.truncator,
				RunStart:  0,
				RunEnd:    len(tc.truncator),
				Direction: di.DirectionLTR,
				Face:      benchEnFace,
				Size:      fixed.I(16),
				Script:    language.Latin,
				Language:  language.NewLanguage("EN"),
			})
			w := LineWrapper{}
			for _, expec := range tc.expectations {
				t.Run(expec.name, func(t *testing.T) {
					expec.config.Truncator = shapedTrunc
					lines, truncated := w.WrapParagraph(expec.config, tc.maxWidth, tc.paragraph, NewSliceIterator([]Output{out.copy()}))
					actualCounts := checkRuneCounts(t, tc.paragraph, lines, truncated)
					if truncated != expec.truncatedCount {
						t.Errorf("expected %d truncated runes, got %d", expec.truncatedCount, truncated)
					}
					for lineNo, expected := range expec.runeCounts {
						if len(actualCounts) <= lineNo {
							t.Errorf("expected %d lines, got %d", len(expec.runeCounts), len(actualCounts))
							break
						}
						actual := actualCounts[lineNo]
						if actual != expected {
							t.Errorf("line %d: expected %d runes, got %d", lineNo, expected, actual)
						}
					}
				})
			}
		})
	}
}
