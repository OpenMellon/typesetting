package shaping

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/OpenMellon/typesetting/di"
	"github.com/OpenMellon/typesetting/font"
	"github.com/OpenMellon/typesetting/language"
	"github.com/OpenMellon/typesetting/segmenter"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
)

func TestWrapping_oneLine(t *testing.T) {
	textInput := []rune("Lorem ipsum") // a simple input that fits on one line
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
	iter := NewSliceIterator(out)
	var l LineWrapper

	outs, _ := l.WrapParagraph(WrapConfig{BreakPolicy: Never}, 250, textInput, iter)
	if len(outs) != 1 {
		t.Errorf("expected one line, got %d", len(outs))
	}

	// the run in iter should have been consumed
	outs, _ = l.WrapParagraph(WrapConfig{BreakPolicy: Never}, 250, textInput, iter)
	if len(outs) != 0 {
		t.Errorf("expected no line, got %d", len(outs))
	}
}

// TestWrappingTruncation checks that the line wrapper's truncation features
// handle some edge cases.
func TestWrappingTruncationEdgeCases(t *testing.T) {
	type testcase struct {
		// name describing what is being tested.
		name string
		// input string to shape.
		input string
		// width to wrap shaped text to.
		wrapWidth int
		// cutInto is a count of how many runs the shaped input should be cut into
		// before wrapping, in order to ensure consistent wrapping behavior regardless
		// of the number of runs involved.
		cutInto int
		// maxLines controls how many lines of text will be wrapped before attempting
		// truncation.
		maxLines int
		// truncator to shape and use as the final run on truncated lines.
		truncator string
		// forceTruncation ensures that the final line in maxLines will show the truncator
		// symbol.
		forceTruncation bool
		// expectedTruncated is the expected count of truncated runes.
		expectedTruncated int
	}
	for _, tc := range []testcase{
		{
			name:              "only run doesn't fit 1 part",
			input:             "mmmmm",
			wrapWidth:         40,
			cutInto:           1,
			maxLines:          1,
			truncator:         "...",
			expectedTruncated: 5,
		},
		{
			name:              "only run doesn't fit 5 parts",
			input:             "mmmmm",
			wrapWidth:         40,
			cutInto:           5,
			maxLines:          1,
			truncator:         "...",
			expectedTruncated: 5,
		},
		{
			name:              "run only fits without truncator 1 part",
			input:             "mmmm",
			wrapWidth:         40,
			cutInto:           1,
			maxLines:          1,
			truncator:         "...",
			expectedTruncated: 0,
		},
		{
			name:              "run only fits without truncator 1 part, forced to truncate",
			input:             "mmmm",
			wrapWidth:         40,
			cutInto:           1,
			maxLines:          1,
			truncator:         "...",
			forceTruncation:   true,
			expectedTruncated: 4,
		},
		{
			name:              "run only fits without truncator 4 part",
			input:             "mmmm",
			wrapWidth:         40,
			cutInto:           4,
			maxLines:          1,
			truncator:         "...",
			expectedTruncated: 0,
		},
		{
			name:              "multi-word run only fits without truncator 1 part",
			input:             "m mm",
			wrapWidth:         40,
			cutInto:           1,
			maxLines:          1,
			truncator:         "...",
			expectedTruncated: 0,
		},
		{
			name:              "multi-word run only fits without truncator 1 part, forced to truncate",
			input:             "m mm",
			wrapWidth:         40,
			cutInto:           1,
			maxLines:          1,
			truncator:         "...",
			forceTruncation:   true,
			expectedTruncated: 0,
		},
		{
			name:              "multi-word run only fits without truncator 4 part",
			input:             "m mm",
			wrapWidth:         40,
			cutInto:           4,
			maxLines:          1,
			truncator:         "...",
			expectedTruncated: 0,
		},
		{
			name:              "multi-word run doesn't fit 1 part",
			input:             "mmm mm",
			wrapWidth:         40,
			cutInto:           1,
			maxLines:          1,
			truncator:         "...",
			expectedTruncated: 2,
		},
		{
			name:              "multi-word run doesn't fit 4 part",
			input:             "mmm mm",
			wrapWidth:         40,
			cutInto:           4,
			maxLines:          1,
			truncator:         "...",
			expectedTruncated: 2,
		},
		{
			name:              "only run doesn't fit and truncator doesn't fit",
			input:             "mm mmm",
			wrapWidth:         40,
			cutInto:           1,
			maxLines:          1,
			truncator:         "mmmmm",
			expectedTruncated: 6,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			inputRunes := []rune(tc.input)
			truncRunes := []rune(tc.truncator)
			var shaper HarfbuzzShaper
			trunc := shaper.Shape(Input{
				Text:      truncRunes,
				RunStart:  0,
				RunEnd:    len(truncRunes),
				Direction: di.DirectionLTR,
				Face:      benchEnFace,
				Size:      fixed.I(10),
				Script:    language.Latin,
				Language:  language.NewLanguage("EN"),
			})
			out := shaper.Shape(Input{
				Text:      inputRunes,
				RunStart:  0,
				RunEnd:    len(inputRunes),
				Direction: di.DirectionLTR,
				Face:      benchEnFace,
				Size:      fixed.I(10),
				Script:    language.Latin,
				Language:  language.NewLanguage("EN"),
			})
			var outs []Output
			if tc.cutInto > 1 {
				outs = cutRunInto(out, tc.cutInto)
			} else {
				outs = append(outs, out)
			}
			var l LineWrapper
			lines, truncatedRunes := l.WrapParagraph(WrapConfig{
				BreakPolicy:        Never,
				Truncator:          trunc,
				TruncateAfterLines: tc.maxLines,
				TextContinues:      tc.forceTruncation,
			}, tc.wrapWidth, inputRunes, NewSliceIterator(outs))
			if truncatedRunes != tc.expectedTruncated {
				t.Errorf("got %d truncated runes when truncation expectation was %d", truncatedRunes, tc.expectedTruncated)
			}
			lastLine := lines[len(lines)-1]
			lastRun := lastLine[len(lastLine)-1]
			shouldTruncate := (tc.expectedTruncated > 0 || tc.forceTruncation)
			trunc.VisualIndex = lastRun.VisualIndex
			trunc.Runes = Range{
				Offset: len(inputRunes) - tc.expectedTruncated,
				Count:  tc.expectedTruncated,
			}
			if lastRunIsTruncator := reflect.DeepEqual(lastRun, trunc); lastRunIsTruncator != shouldTruncate {
				t.Errorf("shouldTruncate = %v, but lastRunIsTruncator = %v", shouldTruncate, lastRunIsTruncator)
			}
		})
	}
}

// TestTruncationWithBreaking tests cases of interest involving both text truncation and
// line breaking policies.
func TestTruncationWithBreaking(t *testing.T) {
	inputText := []rune("Fortunately, there's another way to build streams that is nearly API-compatible with the current approach. You can model them as computational Directed Acyclic Graphs and build a structure of nodes and edges that process values. Nodes essentially run a function on their input edges and propagate the results to their output edges. Edges act as the synchronization primitive in this scheme, ensuring that nodes are able to execute safely in parallel with one another.")
	inputRun := Input{
		Text:      inputText,
		RunStart:  0,
		RunEnd:    len(inputText),
		Direction: di.DirectionLTR,
		Face:      benchEnFace,
		Size:      fixed.I(16),
		Script:    language.Latin,
		Language:  language.NewLanguage("EN"),
	}
	truncatorRun := Input{
		Text:      []rune("…"),
		RunStart:  0,
		RunEnd:    1,
		Direction: di.DirectionLTR,
		Face:      benchEnFace,
		Size:      fixed.I(16),
		Script:    language.Latin,
		Language:  language.NewLanguage("EN"),
	}
	var shaper HarfbuzzShaper
	output := shaper.Shape(inputRun)
	truncator := shaper.Shape(truncatorRun)
	type testcase struct {
		name      string
		width     int
		truncated int
		policy    LineBreakPolicy
	}

	for _, tc := range []testcase{
		{
			name:      "Never",
			width:     936,
			policy:    Never,
			truncated: 341,
		},
		{
			name:      "WhenNecessary",
			width:     936,
			policy:    WhenNecessary,
			truncated: 341,
		},
		{
			name:      "Always",
			width:     936,
			policy:    Always,
			truncated: 341,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var wrapper LineWrapper
			lines, truncated := wrapper.WrapParagraph(WrapConfig{
				BreakPolicy:        tc.policy,
				TruncateAfterLines: 1,
				Truncator:          truncator,
			}, tc.width, inputText, NewSliceIterator([]Output{output.copy()}))
			if truncated != tc.truncated {
				t.Errorf("expected %d truncated runes, got %d (total %d)", tc.truncated, truncated, len(inputText))
			}
			checkRuneCounts(t, inputText, lines, truncated)
		})
	}
}

func TestGraphemeBreakingRegression(t *testing.T) {
	bidiText2 := []rune("renنيتذدagلaaiua.ئPocttأior رادرs")
	inputRuns := []Input{
		{Text: bidiText2, RunStart: 0, RunEnd: 3, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 3, RunEnd: 8, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 8, RunEnd: 10, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 10, RunEnd: 11, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 11, RunEnd: 16, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 16, RunEnd: 18, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 18, RunEnd: 23, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 23, RunEnd: 24, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 24, RunEnd: 27, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 27, RunEnd: 32, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 32, RunEnd: 33, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
	}
	var shaper HarfbuzzShaper
	shaped := []Output{}
	for _, run := range inputRuns {
		shaped = append(shaped, shaper.Shape(run))
	}
	var wrapper LineWrapper
	maxWidth := 0
	for _, run := range shaped {
		maxWidth += run.Advance.Ceil()
	}
	// Wrap at increasingly unreasonable widths, checking for sane wrapping decisions.
	for maxWidth := maxWidth; maxWidth > 0; maxWidth-- {
		t.Run(fmt.Sprintf("maxWidth=%d", maxWidth), func(t *testing.T) {
			runs := make([]Output, len(shaped))
			for i, run := range shaped {
				runs[i] = run.copy()
			}
			lines, truncated := wrapper.WrapParagraph(WrapConfig{
				BreakPolicy:                   Always,
				DisableTrailingWhitespaceTrim: true,
			}, maxWidth, bidiText2, NewSliceIterator(runs))
			checkRuneCounts(t, bidiText2, lines, truncated)

			for i, baseLine := range lines[:len(lines)-1] {
				nextLine := lines[i+1]
				baseAdv := fixed.Int26_6(0)
				for _, run := range baseLine {
					baseAdv += run.Advance
				}
				nextAdv := fixed.Int26_6(0)
				for _, run := range nextLine {
					nextAdv += run.Advance
				}
				if total := (baseAdv + nextAdv).Ceil(); total <= maxWidth {
					t.Errorf("lines[%d] and lines[%d] could have fit on the same line (total width %d)", i, i+1, total)
				}
			}

			if maxWidth == 1 {
				// Check that every grapheme cluster was wrapped properly to its own line.
				seg := segmenter.Segmenter{}
				seg.Init(bidiText2)
				iter := seg.GraphemeIterator()
				lengths := []int{}
				for iter.Next() {
					cluster := iter.Grapheme()
					lengths = append(lengths, len(cluster.Text))
				}
				if len(lengths) != len(lines) {
					t.Errorf("got %d lines but there are %d grapheme clusters", len(lines), len(lengths))
				}
				for lineNum, line := range lines {
					totalRunes := 0
					for _, run := range line {
						totalRunes += run.Runes.Count
					}
					if totalRunes != lengths[lineNum] {
						t.Errorf("lines[%d] has %d runes, but grapheme cluster at index %d has %d", lineNum, totalRunes, lineNum, lengths[lineNum])
					}
				}
			}
		})
	}
}

func BenchmarkMapping(b *testing.B) {
	type wrapfunc func(di.Direction, Range, []Glyph, []glyphIndex) []glyphIndex
	for _, langInfo := range benchLangs {
		for _, size := range []int{10, 100, 1000} {
			for impl, f := range map[string]wrapfunc{
				// "v1": mapRunesToClusterIndices,
				// "v2": mapRunesToClusterIndices2,
				"v3": mapRunesToClusterIndices3,
			} {
				b.Run(fmt.Sprintf("%drunes-%s-%s", size, langInfo.name, impl), func(b *testing.B) {
					var shaper HarfbuzzShaper
					out := shaper.Shape(Input{
						Text:      langInfo.text[:size],
						RunStart:  0,
						RunEnd:    size,
						Direction: langInfo.dir,
						Face:      langInfo.face,
						Size:      16 * 72,
						Script:    langInfo.script,
						Language:  langInfo.lang,
					})
					var m []glyphIndex
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						m = f(out.Direction, out.Runes, out.Glyphs, m)
					}
					_ = m
				})
			}
		}
	}
}

// benchLangInfo describes the language configuration for a text shaping input.
type benchLangInfo struct {
	name   string
	dir    di.Direction
	script language.Script
	lang   language.Language
	face   *font.Face
	text   []rune
}

// benchSizeConfig describes a size of benchmarking input and a set of configurations
// for cutting the runes into quantities of equal-sized parts.
type benchSizeConfig struct {
	runes int
	parts []int
}

// benchArFace is an arabic font face for use in benchmarks.
var benchArFace = func() *font.Face {
	data, err := os.ReadFile("../font/testdata/Amiri-Regular.ttf")
	if err != nil {
		panic(err)
	}
	arFace, err := font.ParseTTF(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	return arFace
}()

// benchEnFace is a latin font face for use in benchmarks.
var benchEnFace = func() *font.Face {
	enFace, err := font.ParseTTF(bytes.NewReader(goregular.TTF))
	if err != nil {
		panic(err)
	}
	return enFace
}()

var benchLangs = []benchLangInfo{
	{
		name:   "arabic",
		dir:    di.DirectionRTL,
		script: language.Arabic,
		lang:   language.NewLanguage("AR"),
		face:   benchArFace,
		text:   []rune(benchParagraphArabic),
	},
	{
		name:   "latin",
		dir:    di.DirectionLTR,
		script: language.Latin,
		lang:   language.NewLanguage("EN"),
		face:   benchEnFace,
		text:   []rune(benchParagraphLatin),
	},
}

var benchSizes = []benchSizeConfig{
	{runes: 10, parts: []int{1, 10}},
	{runes: 100, parts: []int{1, 10, 100}},
	{runes: 1000, parts: []int{1, 10, 100, 1000}},
}

// cutRunInto divides the run into [parts] of same size (with the last part absorbing any remainder).
func cutRunInto(run Output, parts int) []Output {
	var outs []Output
	mapping := mapRunesToClusterIndices3(run.Direction, run.Runes, run.Glyphs, nil)
	runesPerPart := run.Runes.Count / parts
	partStart := 0
	for i := 0; i < parts-1; i++ {
		outs = append(outs, cutRun(run, mapping, partStart, partStart+runesPerPart-1, false))
		partStart += runesPerPart
	}
	outs = append(outs, cutRun(run, mapping, partStart, run.Runes.Count-1, false))
	return outs
}

// TestCutRunInto ensures that the cutRunInto helper function actually cuts the run into the
// right pieces.
func TestCutRunInto(t *testing.T) {
	for _, langInfo := range benchLangs {
		for _, size := range benchSizes {
			for _, parts := range size.parts {
				var shaper HarfbuzzShaper
				out := shaper.Shape(Input{
					Text:      langInfo.text[:size.runes],
					RunStart:  0,
					RunEnd:    size.runes,
					Direction: langInfo.dir,
					Face:      langInfo.face,
					Size:      16 * 72,
					Script:    langInfo.script,
					Language:  langInfo.lang,
				})
				outs := cutRunInto(out, parts)
				accountedRunes := make([]int, size.runes)
				maxRune := -1
				for _, part := range outs {
					for i := part.Runes.Offset; i < part.Runes.Count+part.Runes.Offset; i++ {
						accountedRunes[i]++
						if i > maxRune {
							maxRune = i
						}
					}
				}
				if maxRune != size.runes-1 {
					t.Errorf("maximum rune in cut result is %d, expected %d", maxRune, size.runes-1)
				}
				for runeIdx, count := range accountedRunes {
					if count != 1 {
						t.Errorf("rune at position %d seen %d times", runeIdx, count)
					}
				}
			}
		}
	}
}

func (w *wrapBuffer) stats() string {
	return fmt.Sprintf("paragraph: %d(%d), line: %d(%d), used: %d, exhausted: %v, alt: %d(%d)", len(w.paragraph), cap(w.paragraph), len(w.line), cap(w.line), w.lineUsed, w.lineExhausted, len(w.alt), cap(w.alt))
}
