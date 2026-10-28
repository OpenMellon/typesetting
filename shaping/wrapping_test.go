package shaping

import (
	"reflect"
	"testing"

	"github.com/OpenMellon/typesetting/di"
	"golang.org/x/image/math/fixed"
)

// glyphs returns a slice of glyphs with clusters from start to
// end. If start is greater than end, the glyphs will be returned
// with descending cluster values.
func glyphs(start, end int) []Glyph {
	inc := 1
	if start > end {
		inc = -inc
	}
	num := max(start, end) - min(start, end) + 1
	g := make([]Glyph, 0, num)
	for i := 0; i < num; i++ {
		g = append(g, simpleGlyph(start))
		start += inc
	}
	return g
}

func TestMapRunesToClusterIndices(t *testing.T) {
	type testcase struct {
		name     string
		dir      di.Direction
		runes    Range
		glyphs   []Glyph
		expected []int
	}
	for _, tc := range []testcase{
		{
			name:  "simple",
			dir:   di.DirectionLTR,
			runes: Range{Count: 5},
			glyphs: []Glyph{
				simpleGlyph(0),
				simpleGlyph(1),
				simpleGlyph(2),
				simpleGlyph(3),
				simpleGlyph(4),
			},
			expected: []int{0, 1, 2, 3, 4},
		},
		{
			name:  "simple offset",
			dir:   di.DirectionLTR,
			runes: Range{Count: 5, Offset: 5},
			glyphs: []Glyph{
				simpleGlyph(5),
				simpleGlyph(6),
				simpleGlyph(7),
				simpleGlyph(8),
				simpleGlyph(9),
			},
			expected: []int{0, 1, 2, 3, 4},
		},
		{
			name:  "simple rtl",
			dir:   di.DirectionRTL,
			runes: Range{Count: 5},
			glyphs: []Glyph{
				simpleGlyph(4),
				simpleGlyph(3),
				simpleGlyph(2),
				simpleGlyph(1),
				simpleGlyph(0),
			},
			expected: []int{4, 3, 2, 1, 0},
		},
		{
			name:  "simple offset rtl",
			dir:   di.DirectionRTL,
			runes: Range{Count: 5, Offset: 5},
			glyphs: []Glyph{
				simpleGlyph(9),
				simpleGlyph(8),
				simpleGlyph(7),
				simpleGlyph(6),
				simpleGlyph(5),
			},
			expected: []int{4, 3, 2, 1, 0},
		},
		{
			name:  "fused clusters",
			dir:   di.DirectionLTR,
			runes: Range{Count: 5},
			glyphs: []Glyph{
				complexGlyph(0, 2, 2),
				complexGlyph(0, 2, 2),
				simpleGlyph(2),
				complexGlyph(3, 2, 2),
				complexGlyph(3, 2, 2),
			},
			expected: []int{0, 0, 2, 3, 3},
		},
		{
			name:  "fused clusters rtl",
			dir:   di.DirectionRTL,
			runes: Range{Count: 5},
			glyphs: []Glyph{
				complexGlyph(3, 2, 2),
				complexGlyph(3, 2, 2),
				simpleGlyph(2),
				complexGlyph(0, 2, 2),
				complexGlyph(0, 2, 2),
			},
			expected: []int{3, 3, 2, 0, 0},
		},
		{
			name:  "ligatures",
			dir:   di.DirectionLTR,
			runes: Range{Count: 5},
			glyphs: []Glyph{
				ligatureGlyph(0, 2),
				simpleGlyph(2),
				ligatureGlyph(3, 2),
			},
			expected: []int{0, 0, 1, 2, 2},
		},
		{
			name:  "ligatures rtl",
			dir:   di.DirectionRTL,
			runes: Range{Count: 5},
			glyphs: []Glyph{
				ligatureGlyph(3, 2),
				simpleGlyph(2),
				ligatureGlyph(0, 2),
			},
			expected: []int{2, 2, 1, 0, 0},
		},
		{
			name:  "expansion",
			dir:   di.DirectionLTR,
			runes: Range{Count: 5},
			glyphs: []Glyph{
				simpleGlyph(0),
				expansionGlyph(1, 3),
				expansionGlyph(1, 3),
				expansionGlyph(1, 3),
				simpleGlyph(2),
				simpleGlyph(3),
				simpleGlyph(4),
			},
			expected: []int{0, 1, 4, 5, 6},
		},
		{
			name:  "expansion rtl",
			dir:   di.DirectionRTL,
			runes: Range{Count: 5},
			glyphs: []Glyph{
				simpleGlyph(4),
				simpleGlyph(3),
				simpleGlyph(2),
				expansionGlyph(1, 3),
				expansionGlyph(1, 3),
				expansionGlyph(1, 3),
				simpleGlyph(0),
			},
			expected: []int{6, 3, 2, 1, 0},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mapping := mapRunesToClusterIndices(tc.dir, tc.runes, tc.glyphs, nil)
			if !reflect.DeepEqual(tc.expected, mapping) {
				t.Errorf("expected %v, got %v", tc.expected, mapping)
			}
			mapping = mapRunesToClusterIndices2(tc.dir, tc.runes, tc.glyphs, nil)
			if !reflect.DeepEqual(tc.expected, mapping) {
				t.Errorf("expected %v, got %v", tc.expected, mapping)
			}
			mapping = mapRunesToClusterIndices3(tc.dir, tc.runes, tc.glyphs, nil)
			if !reflect.DeepEqual(tc.expected, mapping) {
				t.Errorf("expected %v, got %v", tc.expected, mapping)
			}
			for runeIdx, glyphIdx := range tc.expected {
				g := mapRuneToClusterIndex(tc.dir, tc.runes, tc.glyphs, runeIdx)
				if g != glyphIdx {
					t.Errorf("map single, expected rune %d to yield %d, got %d", runeIdx, glyphIdx, g)
				}
			}
		})
	}
}

func TestInclusiveRange(t *testing.T) {
	type testcase struct {
		name string
		// inputs
		dir         di.Direction
		start       int
		breakAfter  int
		runeToGlyph []int
		numGlyphs   int
		// expected outputs
		gs, ge int
	}
	for _, tc := range []testcase{
		{
			name:        "simple at start",
			dir:         di.DirectionLTR,
			numGlyphs:   5,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{0, 1, 2, 3, 4},
			gs:          0,
			ge:          2,
		},
		{
			name:        "simple in middle",
			dir:         di.DirectionLTR,
			numGlyphs:   5,
			start:       1,
			breakAfter:  3,
			runeToGlyph: []int{0, 1, 2, 3, 4},
			gs:          1,
			ge:          3,
		},
		{
			name:        "simple at end",
			dir:         di.DirectionLTR,
			numGlyphs:   5,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{0, 1, 2, 3, 4},
			gs:          2,
			ge:          4,
		},
		{
			name:        "simple at start rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   5,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{4, 3, 2, 1, 0},
			gs:          2,
			ge:          4,
		},
		{
			name:        "simple in middle rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   5,
			start:       1,
			breakAfter:  3,
			runeToGlyph: []int{4, 3, 2, 1, 0},
			gs:          1,
			ge:          3,
		},
		{
			name:        "simple at end rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   5,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{4, 3, 2, 1, 0},
			gs:          0,
			ge:          2,
		},
		{
			name:        "fused clusters at start",
			dir:         di.DirectionLTR,
			numGlyphs:   5,
			start:       0,
			breakAfter:  1,
			runeToGlyph: []int{0, 0, 2, 3, 3},
			gs:          0,
			ge:          1,
		},
		{
			name:        "fused clusters start and middle",
			dir:         di.DirectionLTR,
			numGlyphs:   5,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{0, 0, 2, 3, 3},
			gs:          0,
			ge:          2,
		},
		{
			name:        "fused clusters middle and end",
			dir:         di.DirectionLTR,
			numGlyphs:   5,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{0, 0, 2, 3, 3},
			gs:          2,
			ge:          4,
		},
		{
			name:        "fused clusters at end",
			dir:         di.DirectionLTR,
			numGlyphs:   5,
			start:       3,
			breakAfter:  4,
			runeToGlyph: []int{0, 0, 2, 3, 3},
			gs:          3,
			ge:          4,
		},
		{
			name:        "fused clusters at start rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   5,
			start:       0,
			breakAfter:  1,
			runeToGlyph: []int{3, 3, 2, 0, 0},
			gs:          3,
			ge:          4,
		},
		{
			name:        "fused clusters start and middle rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   5,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{3, 3, 2, 0, 0},
			gs:          2,
			ge:          4,
		},
		{
			name:        "fused clusters middle and end rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   5,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{3, 3, 2, 0, 0},
			gs:          0,
			ge:          2,
		},
		{
			name:        "fused clusters at end rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   5,
			start:       3,
			breakAfter:  4,
			runeToGlyph: []int{3, 3, 2, 0, 0},
			gs:          0,
			ge:          1,
		},
		{
			name:        "ligatures at start",
			dir:         di.DirectionLTR,
			numGlyphs:   3,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{0, 0, 1, 2, 2},
			gs:          0,
			ge:          1,
		},
		{
			name:        "ligatures in middle",
			dir:         di.DirectionLTR,
			numGlyphs:   3,
			start:       2,
			breakAfter:  2,
			runeToGlyph: []int{0, 0, 1, 2, 2},
			gs:          1,
			ge:          1,
		},
		{
			name:        "ligatures at end",
			dir:         di.DirectionLTR,
			numGlyphs:   3,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{0, 0, 1, 2, 2},
			gs:          1,
			ge:          2,
		},
		{
			name:        "ligatures at start rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   3,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{2, 2, 1, 0, 0},
			gs:          1,
			ge:          2,
		},
		{
			name:        "ligatures in middle rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   3,
			start:       2,
			breakAfter:  2,
			runeToGlyph: []int{2, 2, 1, 0, 0},
			gs:          1,
			ge:          1,
		},
		{
			name:        "ligatures at end rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   3,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{2, 2, 1, 0, 0},
			gs:          0,
			ge:          1,
		},
		{
			name:        "expansion at start",
			dir:         di.DirectionLTR,
			numGlyphs:   7,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{0, 1, 4, 5, 6},
			gs:          0,
			ge:          4,
		},
		{
			name:        "expansion in middle",
			dir:         di.DirectionLTR,
			numGlyphs:   7,
			start:       1,
			breakAfter:  3,
			runeToGlyph: []int{0, 1, 4, 5, 6},
			gs:          1,
			ge:          5,
		},
		{
			name:        "expansion at end",
			dir:         di.DirectionLTR,
			numGlyphs:   7,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{0, 1, 4, 5, 6},
			gs:          4,
			ge:          6,
		},
		{
			name:        "expansion at start rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   7,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{6, 3, 2, 1, 0},
			gs:          2,
			ge:          6,
		},
		{
			name:        "expansion in middle rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   7,
			start:       1,
			breakAfter:  3,
			runeToGlyph: []int{6, 3, 2, 1, 0},
			gs:          1,
			ge:          5,
		},
		{
			name:        "expansion at end rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   7,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{6, 3, 2, 1, 0},
			gs:          0,
			ge:          2,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gs, ge := inclusiveGlyphRange(tc.dir, tc.start, tc.breakAfter, tc.runeToGlyph, tc.numGlyphs)
			if gs != tc.gs {
				t.Errorf("glyphStart mismatch, got %d, expected %d", gs, tc.gs)
			}
			if ge != tc.ge {
				t.Errorf("glyphEnd mismatch, got %d, expected %d", ge, tc.ge)
			}
		})
	}
}

func withRange(output Output, runes Range) Output {
	output.Runes = runes
	return output
}

var (
	oneWord       = "aaaa"
	shapedOneWord = Output{
		Advance: fixed.I(10 * len([]rune(oneWord))),
		LineBounds: Bounds{
			Ascent:  fixed.I(10),
			Descent: fixed.I(5),
			// No line gap.
		},
		GlyphBounds: Bounds{
			Ascent: fixed.I(10),
			// No glyphs descend.
		},
		Glyphs: glyphs(0, len(oneWord)-1),
		Runes: Range{
			Count: len([]rune(oneWord)),
		},
	}
	// Assume the simple case of 1:1:1 glyph:rune:byte for this input.
	text1       = "text one is ltr"
	shapedText1 = Output{
		Advance: fixed.I(10 * len([]rune(text1))),
		LineBounds: Bounds{
			Ascent:  fixed.I(10),
			Descent: fixed.I(5),
			// No line gap.
		},
		GlyphBounds: Bounds{
			Ascent: fixed.I(10),
			// No glyphs descend.
		},
		Glyphs: glyphs(0, 14),
		Runes: Range{
			Count: len([]rune(text1)),
		},
	}
	text1Trailing       = text1 + " "
	shapedText1Trailing = func() Output {
		out := shapedText1
		out.Runes.Count++
		out.Glyphs = append(out.Glyphs, simpleGlyph(len(out.Glyphs)))
		out.RecalculateAll()
		return out
	}()
	// Test M:N:O glyph:rune:byte for this input.
	// The substring `lig` is shaped as a ligature.
	// The substring `DROP` is not shaped at all.
	text2       = "안П你 ligDROP 안П你 ligDROP"
	shapedText2 = Output{
		// There are 11 glyphs shaped for this string.
		Advance: fixed.I(10 * 11),
		LineBounds: Bounds{
			Ascent:  fixed.I(10),
			Descent: fixed.I(5),
			// No line gap.
		},
		GlyphBounds: Bounds{
			Ascent: fixed.I(10),
			// No glyphs descend.
		},
		Glyphs: []Glyph{
			0: simpleGlyph(0),      // 안        - 4 bytes
			1: simpleGlyph(1),      // П         - 3 bytes
			2: simpleGlyph(2),      // 你        - 4 bytes
			3: simpleGlyph(3),      // <space>   - 1 byte
			4: ligatureGlyph(4, 7), // lig       - 3 runes, 3 bytes
			// DROP                   - 4 runes, 4 bytes (included in ligature rune count)
			5:  simpleGlyph(11),      // <space> - 1 byte
			6:  simpleGlyph(12),      // 안      - 4 bytes
			7:  simpleGlyph(13),      // П       - 3 bytes
			8:  simpleGlyph(14),      // 你      - 4 bytes
			9:  simpleGlyph(15),      // <space> - 1 byte
			10: ligatureGlyph(16, 7), // lig     - 3 runes, 3 bytes
			// DROP                   - 4 runes, 4 bytes (included in ligature rune count)
		},
		Runes: Range{
			Count: len([]rune(text2)),
		},
	}
	// Test RTL languages.
	text3       = "שלום أهلا שלום أهلا"
	shapedText3 = Output{
		// There are 15 glyphs shaped for this string.
		Advance: fixed.I(10 * 15),
		LineBounds: Bounds{
			Ascent:  fixed.I(10),
			Descent: fixed.I(5),
			// No line gap.
		},
		GlyphBounds: Bounds{
			Ascent: fixed.I(10),
			// No glyphs descend.
		},
		Glyphs: []Glyph{
			0: ligatureGlyph(16, 3), // LIGATURE of three runes:
			//                         ا - 3 bytes
			//                         ل - 3 bytes
			//                         ه - 3 bytes
			1: simpleGlyph(15),     // أ - 3 bytes
			2: simpleGlyph(14),     // <space> - 1 byte
			3: simpleGlyph(13),     // ם - 3 bytes
			4: simpleGlyph(12),     // ו - 3 bytes
			5: simpleGlyph(11),     // ל - 3 bytes
			6: simpleGlyph(10),     // ש - 3 bytes
			7: simpleGlyph(9),      // <space> - 1 byte
			8: ligatureGlyph(6, 3), // LIGATURE of three runes:
			//                         ا - 3 bytes
			//                         ل - 3 bytes
			//                         ه - 3 bytes
			9:  simpleGlyph(5), //     أ - 3 bytes
			10: simpleGlyph(4), //     <space> - 1 byte
			11: simpleGlyph(3), //     ם - 3 bytes
			12: simpleGlyph(2), //     ו - 3 bytes
			13: simpleGlyph(1), //     ל - 3 bytes
			14: simpleGlyph(0), //     ש - 3 bytes
		},
		Direction: di.DirectionRTL,
		Runes: Range{
			Count: len([]rune(text3)),
		},
	}
	multiInputText1       = "aa aa aa"
	shapedMultiInputText1 = Output{
		Advance: fixed.I(10 * len([]rune(multiInputText1))),
		LineBounds: Bounds{
			Ascent:  fixed.I(10),
			Descent: fixed.I(5),
			// No line gap.
		},
		GlyphBounds: Bounds{
			Ascent: fixed.I(10),
			// No glyphs descend.
		},
		Glyphs: glyphs(0, len([]rune(multiInputText1))-1),
		Runes:  Range{Count: len([]rune(multiInputText1))},
	}
	splitShapedMultiInput1 = splitShapedAt(shapedMultiInputText1, 4, 6)

	bidiText1       = "hello أهلا שלום test"
	shapedBidiText1 = []Output{
		{
			// LTR initial segment
			Advance:   fixed.I(10 * len([]rune("hello "))),
			Direction: di.DirectionLTR,
			Runes: Range{
				Count: len([]rune("hello ")),
			},
			LineBounds: Bounds{
				Ascent:  fixed.I(10),
				Descent: fixed.I(5),
			},
			GlyphBounds: Bounds{
				Ascent: fixed.I(10),
			},
			Glyphs: glyphs(0, len([]rune("hello "))-1),
		},
		{
			// RTL middle segment
			Advance:   fixed.I(10 * len([]rune("أهلا שלום "))),
			Direction: di.DirectionRTL,
			Runes: Range{
				Offset: len([]rune("hello ")),
				Count:  len([]rune("أهلا שלום ")),
			},
			LineBounds: Bounds{
				Ascent:  fixed.I(10),
				Descent: fixed.I(5),
			},
			GlyphBounds: Bounds{
				Ascent: fixed.I(10),
			},
			Glyphs: glyphs(len([]rune("hello أهلا שלום "))-1, len([]rune("hello "))),
		},
		{
			// LTR final segment
			Advance:   fixed.I(10 * len([]rune("test"))),
			Direction: di.DirectionLTR,
			Runes: Range{
				Offset: len([]rune("hello أهلا שלום ")),
				Count:  len([]rune("test")),
			},
			LineBounds: Bounds{
				Ascent:  fixed.I(10),
				Descent: fixed.I(5),
			},
			GlyphBounds: Bounds{
				Ascent: fixed.I(10),
			},
			Glyphs: glyphs(len([]rune("hello أهلا שלום ")), len([]rune("hello أهلا שלום test"))-1),
		},
	}
)
