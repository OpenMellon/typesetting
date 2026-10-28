package shaping

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"testing/quick"

	"github.com/OpenMellon/typesetting/di"
	"github.com/OpenMellon/typesetting/language"
	"github.com/OpenMellon/typesetting/segmenter"
	"golang.org/x/image/math/fixed"
)

func checkOptions(t *testing.T, runes []rune, options []breakOption) bool {
	t.Helper()
	// Ensure breaks are in valid range.
	for _, o := range options {
		if o.breakAtRune < 0 || o.breakAtRune > len(runes)-1 {
			t.Errorf("breakAtRune out of bounds: %d when len(runes)=%d", o.breakAtRune, len(runes))
			return false
		}
	}
	// Ensure breaks are sorted.
	if !sort.SliceIsSorted(options, func(i, j int) bool {
		return options[i].breakAtRune < options[j].breakAtRune
	}) {
		t.Errorf("breaks are not sorted: %#+v", options)
		return false
	}

	// Ensure breaks are unique.
	m := make([]bool, len(runes))
	for _, o := range options {
		if m[o.breakAtRune] {
			t.Errorf("breaks are not unique: %v is repeated in %#+v", o, m)
			return false
		} else {
			m[o.breakAtRune] = true
		}
	}

	return true
}

func TestRawBreakOptions(t *testing.T) {
	if err := quick.Check(func(runes []rune) bool {
		breaker := newBreaker(&segmenter.Segmenter{}, runes)
		var wordOptions []breakOption
		for b, ok := breaker.nextWordRaw(); ok; b, ok = breaker.nextWordRaw() {
			wordOptions = append(wordOptions, b)
		}
		var graphemeOptions []breakOption
		for b, ok := breaker.nextGraphemeRaw(); ok; b, ok = breaker.nextGraphemeRaw() {
			graphemeOptions = append(graphemeOptions, b)
		}
		return checkOptions(t, runes, wordOptions) && checkOptions(t, runes, graphemeOptions)
	}, nil); err != nil {
		t.Errorf("generated invalid break options: %v", err)
	}
}

// TestWrappingLatinE2E actually performs both text shaping and line wrapping
// on a selection of latin text.
func TestWrappingLatinE2E(t *testing.T) {
	textInput := []rune("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.")
	face := benchEnFace
	var shaper HarfbuzzShaper
	out := []Output{shaper.Shape(Input{
		Text:      textInput,
		RunStart:  0,
		RunEnd:    len(textInput),
		Direction: di.DirectionLTR,
		Face:      face,
		Size:      16 * 72,
		Script:    language.Latin,
		Language:  language.NewLanguage("EN"),
	})}
	var l LineWrapper
	outs, _ := l.WrapParagraph(WrapConfig{BreakPolicy: Never}, 250, textInput, NewSliceIterator(out))
	if len(outs) < 3 {
		t.Errorf("expected %d lines, got %d", 3, len(outs))
	}
}

// TestWrappingBidiRegression checks a specific regression discovered within the Gio test suite.
func TestWrappingBidiRegression(t *testing.T) {
	type testcase struct {
		name     string
		text     []rune
		inputs   []Output
		maxWidth int
	}
	ltrText := []rune("The quick brown fox jumps over the lazy dog.")
	ltrTextRuns := []Input{
		{RunStart: 0, RunEnd: len(ltrText), Script: language.Latin, Direction: di.DirectionLTR},
	}
	bidiText := []rune("الحب سماء brown привет fox تمط jumps привет over غير الأحلام")
	bidiTextRuns := []Input{
		{RunStart: 0, RunEnd: 10, Script: language.Arabic, Direction: di.DirectionRTL},
		{RunStart: 10, RunEnd: 16, Script: language.Latin, Direction: di.DirectionLTR},
		{RunStart: 16, RunEnd: 23, Script: language.Cyrillic, Direction: di.DirectionLTR},
		{RunStart: 23, RunEnd: 26, Script: language.Latin, Direction: di.DirectionLTR},
		{RunStart: 26, RunEnd: 31, Script: language.Arabic, Direction: di.DirectionRTL},
		{RunStart: 31, RunEnd: 37, Script: language.Latin, Direction: di.DirectionLTR},
		{RunStart: 37, RunEnd: 44, Script: language.Cyrillic, Direction: di.DirectionLTR},
		{RunStart: 44, RunEnd: 48, Script: language.Latin, Direction: di.DirectionLTR},
		{RunStart: 48, RunEnd: 60, Script: language.Arabic, Direction: di.DirectionRTL},
	}
	bidiText2 := []rune("د عرمثال dstي met لم aqل جدmوpمg lرe dرd  لو عل ميrةsdiduntut lab renنيتذدagلaaiua.ئPocttأior رادرsاي mيrbلmnonaيdتد ماةعcلخ.")
	bidiTextRuns2 := []Input{
		{Text: bidiText2, RunStart: 0, RunEnd: 9, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 9, RunEnd: 12, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 12, RunEnd: 14, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 14, RunEnd: 17, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 17, RunEnd: 21, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 21, RunEnd: 23, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 23, RunEnd: 27, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 27, RunEnd: 28, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 28, RunEnd: 29, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 29, RunEnd: 30, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 30, RunEnd: 31, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 31, RunEnd: 34, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 34, RunEnd: 35, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 35, RunEnd: 38, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 38, RunEnd: 39, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 39, RunEnd: 40, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 40, RunEnd: 50, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 50, RunEnd: 51, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 51, RunEnd: 52, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 52, RunEnd: 69, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 69, RunEnd: 74, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 74, RunEnd: 76, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 76, RunEnd: 77, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 77, RunEnd: 82, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 82, RunEnd: 84, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 84, RunEnd: 89, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 89, RunEnd: 90, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 90, RunEnd: 93, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 93, RunEnd: 98, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 98, RunEnd: 99, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 99, RunEnd: 102, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 102, RunEnd: 103, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 103, RunEnd: 104, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 104, RunEnd: 106, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 106, RunEnd: 107, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 107, RunEnd: 112, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 112, RunEnd: 113, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 113, RunEnd: 114, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 114, RunEnd: 121, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 121, RunEnd: 122, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 122, RunEnd: 125, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
	}
	truncator := (&HarfbuzzShaper{}).Shape(Input{
		Text:      []rune("…"),
		RunStart:  0,
		RunEnd:    len([]rune("…")),
		Direction: di.DirectionLTR,
		Face:      benchEnFace,
		Size:      fixed.I(16),
		Script:    language.Latin,
		Language:  language.NewLanguage("EN"),
	})
	shapeInputs := func(inputs []Input) []Output {
		var shaper HarfbuzzShaper
		out := make([]Output, len(inputs))
		for i, input := range inputs {
			out[i] = shaper.Shape(input)
		}
		return out
	}
	applyDefaultsAndShape := func(textInput []rune, runs []Input) []Output {
		enFace := benchEnFace
		arFace := benchArFace

		ppem := fixed.I(16)
		arLang := language.NewLanguage("AR")

		for i := range runs {
			runs[i].Text = textInput
			runs[i].Size = ppem
			if runs[i].Direction == di.DirectionRTL {
				runs[i].Face = arFace
			} else {
				runs[i].Face = enFace
			}
			// Even though the text sample is mixed, the overall document language is arabic.
			runs[i].Language = arLang
		}
		return shapeInputs(runs)
	}
	for _, tc := range []testcase{
		{
			name:     "simple ltr",
			text:     ltrText,
			inputs:   applyDefaultsAndShape(ltrText, ltrTextRuns),
			maxWidth: 100,
		},
		{
			name:     "many-run bidi",
			text:     bidiText,
			inputs:   applyDefaultsAndShape(bidiText, bidiTextRuns),
			maxWidth: 100,
		},
		{
			name:     "complex bidi",
			text:     bidiText2,
			inputs:   shapeInputs(bidiTextRuns2),
			maxWidth: 200,
		},
		{
			// This test case was observed failing Gio's text shaping fuzzer. It causes our line
			// wrapper to drop an entire run of text for some reason.
			name:     "bidi regression",
			text:     []rune("000000000000000 00000000 ٰ00000"),
			inputs:   regressionRuns,
			maxWidth: 121,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, policy := range []LineBreakPolicy{Always, Never, WhenNecessary} {
				for _, truncation := range []bool{true, false} {
					t.Run(fmt.Sprintf("graphemeBreak=%s_truncate=%v", policy, truncation), func(t *testing.T) {
						var l LineWrapper
						truncateAfter := 0
						if truncation {
							truncateAfter = 1
						}
						lines, truncated := l.WrapParagraph(WrapConfig{
							BreakPolicy:        policy,
							TruncateAfterLines: truncateAfter,
							Truncator:          truncator,
						}, tc.maxWidth, tc.text, NewSliceIterator(tc.inputs))
						if truncated != 0 && !truncation {
							t.Errorf("did not expect truncation, got truncated=%d", truncated)
						}
						checkRuneCounts(t, tc.text, lines, truncated)
					})
				}
			}
		})
	}
}

func checkRuneCounts(t *testing.T, source []rune, lines []Line, truncated int) []int {
	t.Helper()
	counts := []int{}
	totalRunes := 0
	alreadyFailed := t.Failed()
	for lineIdx, line := range lines {
		lineTotalRunes := 0
		for runIdx, run := range line {
			if truncated > 0 && totalRunes == len(source)-truncated && run.Runes.Offset == totalRunes {
				// Skip the truncator run.
				continue
			}
			if run.Runes.Offset != totalRunes {
				if !alreadyFailed {
					for i := 0; i <= lineIdx; i++ {
						line := lines[i]
						for k := 0; (i < lineIdx && k < len(line)) || (i == lineIdx && k < runIdx); k++ {
							run := line[k]
							t.Errorf("lines[%d][%d].Runes.Offset=%d, Count=%d", i, k, run.Runes.Offset, run.Runes.Count)
						}
					}
					alreadyFailed = true
				}
				t.Errorf("lines[%d][%d].Runes.Offset=%d, Count=%d, expected Offset %d", lineIdx, runIdx, run.Runes.Offset, run.Runes.Count, totalRunes)
			}
			totalRunes += run.Runes.Count
			lineTotalRunes += run.Runes.Count
		}
		counts = append(counts, lineTotalRunes)
	}
	if len(source)-truncated != totalRunes {
		t.Errorf("expected %d runes total, got %d", len(source)-truncated, totalRunes)
	}
	return counts
}

// TestWrappingTruncation checks that the line wrapper's truncation features
// behave as expected.
func TestWrappingTruncation(t *testing.T) {
	textInput := []rune("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.")
	face := benchEnFace
	var shaper HarfbuzzShaper
	out := []Output{shaper.Shape(Input{
		Text:      textInput,
		RunStart:  0,
		RunEnd:    len(textInput),
		Direction: di.DirectionLTR,
		Face:      face,
		Size:      fixed.I(16),
		Script:    language.Latin,
		Language:  language.NewLanguage("EN"),
	})}
	var l LineWrapper
	outs, _ := l.WrapParagraph(WrapConfig{BreakPolicy: Never}, 250, textInput, NewSliceIterator(out))
	untruncatedCount := len(outs)

	for _, truncator := range []Output{
		{}, // No truncator.
		shaper.Shape(Input{ // Multi-rune truncator.
			Text:      []rune("..."),
			RunStart:  0,
			RunEnd:    len([]rune("...")),
			Direction: di.DirectionLTR,
			Face:      face,
			Size:      fixed.I(16),
			Script:    language.Latin,
			Language:  language.NewLanguage("EN"),
		}),
		shaper.Shape(Input{ // Single-rune truncator.
			Text:      []rune("…"),
			RunStart:  0,
			RunEnd:    len([]rune("…")),
			Direction: di.DirectionLTR,
			Face:      face,
			Size:      fixed.I(16),
			Script:    language.Latin,
			Language:  language.NewLanguage("EN"),
		}),
	} {
		for i := untruncatedCount + 1; i > 0; i-- {
			wc := WrapConfig{
				BreakPolicy:        Never,
				TruncateAfterLines: i,
				Truncator:          truncator,
			}
			newLines, truncated := l.WrapParagraph(wc, 250, textInput, NewSliceIterator(out))
			lineCount := len(newLines)
			t.Logf("wrapping with max lines=%d, untruncatedCount=%d", i, untruncatedCount)
			if i < untruncatedCount {
				if lineCount != i {
					t.Errorf("expected %d lines, got %d", i, lineCount)
				}
				if truncated < 1 {
					t.Errorf("expected lines to indicate truncation")
				}
				lastLine := newLines[len(newLines)-1]
				lastRun := lastLine[len(lastLine)-1]
				expected := truncator
				expected.Runes = Range{
					Offset: len(textInput) - truncated,
					Count:  truncated,
				}
				expected.VisualIndex = lastRun.VisualIndex
				if !reflect.DeepEqual(lastRun, expected) {
					t.Errorf("expected truncator as last run")
				}
			} else if i >= untruncatedCount {
				if lineCount != untruncatedCount {
					t.Errorf("expected %d lines, got %d", untruncatedCount, lineCount)
				}
				if truncated > 0 {
					t.Errorf("expected lines to indicate no truncation")
				}
			}
			runeCount := 0
			for _, line := range newLines {
				for _, run := range line {
					runeCount += run.Runes.Count
				}
			}
			if runeCount != len(textInput) {
				t.Errorf("expected %d runes total, got %d output", len(textInput), runeCount)
			}
		}
	}
}
