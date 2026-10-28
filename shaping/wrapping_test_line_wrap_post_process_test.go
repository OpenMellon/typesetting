package shaping

import (
	"math"
	"testing"

	"github.com/OpenMellon/typesetting/di"
	"github.com/OpenMellon/typesetting/language"
	tu "github.com/OpenMellon/typesetting/testutils"
	"golang.org/x/image/math/fixed"
)

func TestLineWrapPostProcess(t *testing.T) {
	output := []Output{
		{
			Advance: 10,
			Glyphs: []Glyph{
				{
					Width:        10,
					XAdvance:     10,
					ClusterIndex: 0,
					RuneCount:    1,
					GlyphCount:   1,
				},
			},
			Direction: di.DirectionLTR,
			Runes: Range{
				Offset: 0,
				Count:  1,
			},
		},
		{
			Advance:   20,
			Direction: di.DirectionRTL,
			Glyphs: []Glyph{
				{
					XAdvance:     10,
					Width:        10,
					ClusterIndex: 1,
					RuneCount:    1,
					GlyphCount:   1,
				},
				{
					XAdvance:     10,
					Width:        0,
					ClusterIndex: 2,
					RuneCount:    1,
					GlyphCount:   1,
				},
			},
			Runes: Range{
				Offset: 1,
				Count:  2,
			},
		},
		{
			Advance:   10,
			Direction: di.DirectionRTL,
			Glyphs: []Glyph{
				{
					XAdvance:     10,
					Width:        10,
					ClusterIndex: 3,
					RuneCount:    1,
					GlyphCount:   1,
				},
			},
			Runes: Range{
				Offset: 3,
				Count:  1,
			},
		},
	}
	t.Run("LTR", func(t *testing.T) {
		// Wrap as LTR, which places the space visually at the end of the line.
		lines, _ := (&LineWrapper{}).WrapParagraph(
			WrapConfig{
				BreakPolicy: Always,
				Direction:   di.DirectionLTR,
			}, 50, []rune("XX X"), NewSliceIterator(output),
		)
		if len(lines) > 1 {
			t.Errorf("expected %d lines, got %d", 1, len(lines))
		}
		line := lines[0]
		adv := fixed.Int26_6(0)
		for _, run := range line {
			adv += run.Advance
		}
		if adv > fixed.Int26_6(30) {
			t.Errorf("expected total advance %d, got %d", 30, int(adv))
		}
	})
	t.Run("RTL", func(t *testing.T) {
		// Wrap as RTL, which places the space visually within the line.
		lines, _ := (&LineWrapper{}).WrapParagraph(
			WrapConfig{
				BreakPolicy: Always,
				Direction:   di.DirectionRTL,
			}, 50, []rune("XX X"), NewSliceIterator(output),
		)
		if len(lines) > 1 {
			t.Errorf("expected %d lines, got %d", 1, len(lines))
		}
		line := lines[0]
		adv := fixed.Int26_6(0)
		for _, run := range line {
			adv += run.Advance
		}
		if adv > fixed.Int26_6(40) {
			t.Errorf("expected total advance %d, got %d", 30, int(adv))
		}
	})
}

func TestComputeBidiOrdering(t *testing.T) {
	type testcase struct {
		name                string
		input               []Output
		direction           di.Direction
		expectedVisualOrder []int
	}
	for _, tc := range []testcase{
		{
			name:      "ltr",
			direction: di.DirectionLTR,
			input: []Output{
				{Direction: di.DirectionLTR},
				{Direction: di.DirectionLTR},
				{Direction: di.DirectionLTR},
			},
			expectedVisualOrder: []int{0, 1, 2},
		},
		{
			name:      "rtl",
			direction: di.DirectionRTL,
			input: []Output{
				{Direction: di.DirectionRTL},
				{Direction: di.DirectionRTL},
				{Direction: di.DirectionRTL},
			},
			expectedVisualOrder: []int{2, 1, 0},
		},
		{
			name:      "bidi-ltr",
			direction: di.DirectionLTR,
			input: []Output{
				{Direction: di.DirectionLTR},
				{Direction: di.DirectionRTL},
				{Direction: di.DirectionRTL},
				{Direction: di.DirectionRTL},
				{Direction: di.DirectionLTR},
			},
			expectedVisualOrder: []int{0, 3, 2, 1, 4},
		},
		{
			name:      "bidi-ltr-complex",
			direction: di.DirectionLTR,
			input: []Output{
				{Direction: di.DirectionRTL},
				{Direction: di.DirectionRTL},
				{Direction: di.DirectionLTR},
				{Direction: di.DirectionRTL},
				{Direction: di.DirectionRTL},
				{Direction: di.DirectionLTR},
				{Direction: di.DirectionRTL},
				{Direction: di.DirectionRTL},
				{Direction: di.DirectionLTR},
				{Direction: di.DirectionRTL},
				{Direction: di.DirectionRTL},
			},
			expectedVisualOrder: []int{1, 0, 2, 4, 3, 5, 7, 6, 8, 10, 9},
		},
		{
			name:      "bidi-rtl",
			direction: di.DirectionRTL,
			input: []Output{
				{Direction: di.DirectionRTL},
				{Direction: di.DirectionLTR},
				{Direction: di.DirectionLTR},
				{Direction: di.DirectionLTR},
				{Direction: di.DirectionRTL},
			},
			expectedVisualOrder: []int{4, 1, 2, 3, 0},
		},
		{
			name:      "bidi-rtl-complex",
			direction: di.DirectionRTL,
			input: []Output{
				{Direction: di.DirectionLTR},
				{Direction: di.DirectionLTR},
				{Direction: di.DirectionRTL},
				{Direction: di.DirectionLTR},
				{Direction: di.DirectionLTR},
				{Direction: di.DirectionRTL},
				{Direction: di.DirectionLTR},
				{Direction: di.DirectionLTR},
				{Direction: di.DirectionRTL},
				{Direction: di.DirectionLTR},
				{Direction: di.DirectionLTR},
			},
			expectedVisualOrder: []int{9, 10, 8, 6, 7, 5, 3, 4, 2, 0, 1},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			computeBidiOrdering(tc.direction, tc.input)
			for visualIndex, logicalIndex := range tc.expectedVisualOrder {
				if tc.input[logicalIndex].VisualIndex != int32(visualIndex) {
					t.Errorf("line[%d]: expected visual index %v, got %v", logicalIndex, visualIndex, tc.input[logicalIndex].VisualIndex)
				}
			}
		})
	}
}

func TestRequiredBreaks(t *testing.T) {
	// assume single run, 1 to 1 rune <-> glyph mapping
	shapedText := func(text []rune, line Line) string {
		tu.Assert(t, len(line) == 1)
		runLine := line[0]
		var s []rune
		for _, g := range runLine.Glyphs {
			s = append(s, text[g.ClusterIndex])
		}
		return string(s)
	}

	text := []rune("The\nquick\nbrown\nfox")

	face := loadOpentypeFont(t, "../font/testdata/UbuntuMono-R.ttf")
	run := (&HarfbuzzShaper{}).Shape(Input{
		Text:   text,
		Face:   face,
		Size:   72,
		RunEnd: len(text),
	})
	tu.Assert(t, run.Glyphs[0].XAdvance == fixed.I(1))

	lines, _ := (&LineWrapper{}).WrapParagraph(WrapConfig{BreakPolicy: Always}, 100, text, NewSliceIterator([]Output{run}))

	expected := []string{
		"The\n",
		"quick\n",
		"brown\n",
		"fox",
	}
	if len(lines) != len(expected) {
		t.Errorf("expected %d lines, got %d", len(expected), len(lines))
	}
	tu.Assert(t, len(lines) == 4)
	for i, line := range lines {
		s := shapedText(text, line)
		if s != expected[i] {
			t.Errorf("expected %s, got %s", expected[i], s)
		}
	}
}

func TestTrimmedTrailingWhitespace(t *testing.T) {
	face := loadOpentypeFont(t, "../font/testdata/UbuntuMono-R.ttf")
	text := []rune("The quick")
	run := (&HarfbuzzShaper{}).Shape(Input{
		Text:   text,
		Face:   face,
		Size:   72,
		RunEnd: len(text),
	})
	tu.Assert(t, run.Glyphs[0].XAdvance == fixed.I(1) && run.Glyphs[3].XAdvance == fixed.I(1))

	var wrapper LineWrapper

	wrapper.Prepare(WrapConfig{BreakPolicy: WhenNecessary}, text, NewSliceIterator([]Output{run.copy()}))
	line, _ := wrapper.WrapNextLine(4) // cut right after the space
	tu.Assert(t, line.NextLine == 4)
	tu.Assert(t, line.Line[0].Advance == fixed.I(3))           // the space is collapsed
	tu.Assert(t, line.TrimmedTrailingWhitespace == fixed.I(1)) // and we know how much was collapsed

	wrapper.Prepare(WrapConfig{BreakPolicy: WhenNecessary, DisableTrailingWhitespaceTrim: true}, text, NewSliceIterator([]Output{run.copy()}))
	line, _ = wrapper.WrapNextLine(4) // cut right after the space
	tu.Assert(t, line.NextLine == 4)
	tu.Assert(t, line.Line[0].Advance == fixed.I(4)) // the space is not collapsed
	tu.Assert(t, line.TrimmedTrailingWhitespace == fixed.I(0))
}

func TestMaxWidthRouding(t *testing.T) {
	text := []rune("a simple word") // odd number of letters (13)

	face := loadOpentypeFont(t, "../font/testdata/UbuntuMono-R.ttf")
	run := (&HarfbuzzShaper{}).Shape(Input{
		Text:   text,
		Face:   face,
		Size:   36,
		RunEnd: len(text),
	})
	tu.Assert(t, run.Glyphs[0].XAdvance == fixed.I(1)/2)
	tu.Assert(t, run.Advance == 13*fixed.I(1)/2)

	wr := LineWrapper{}

	wr.Prepare(WrapConfig{}, text, NewSliceIterator([]Output{run}))
	line, _ := wr.WrapNextLine(run.Advance.Floor())
	tu.Assert(t, line.NextLine == 9)

	wr.Prepare(WrapConfig{}, text, NewSliceIterator([]Output{run}))
	line, _ = wr.WrapNextLineF(run.Advance)
	tu.Assert(t, line.NextLine == 13)
}

func TestWrapping_oneLine_overflow_bug(t *testing.T) {
	maxWidth := math.MaxInt

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

	outs, _ := l.WrapParagraph(WrapConfig{BreakPolicy: Never}, maxWidth, textInput, iter)
	if len(outs) != 1 {
		t.Errorf("expected one line, got %d", len(outs))
	}

	// the run in iter should have been consumed
	outs, _ = l.WrapParagraph(WrapConfig{BreakPolicy: Never}, maxWidth, textInput, iter)
	if len(outs) != 0 {
		t.Errorf("expected no line, got %d", len(outs))
	}

	l.Prepare(WrapConfig{BreakPolicy: Never}, textInput, NewSliceIterator(out))
	_, done := l.WrapNextLine(maxWidth)
	tu.Assert(t, done)
}
