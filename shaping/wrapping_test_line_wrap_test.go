package shaping

import (
	"reflect"
	"testing"

	"github.com/OpenMellon/typesetting/di"
	"golang.org/x/image/math/fixed"
)

func TestLineWrap(t *testing.T) {
	emptyString := []Output{{
		Size:       fixed.I(100),
		LineBounds: Bounds{Ascent: fixed.I(-100), Descent: fixed.I(50)},
	}}
	type testcase struct {
		name      string
		shaped    []Output
		paragraph []rune
		maxWidth  int
		expected  []Line
	}
	for _, tc := range []testcase{
		{
			// Wrapping the empty string should produce a run with no glyphs/runes but
			// with valid line bounds.
			name:      "empty string",
			shaped:    emptyString,
			paragraph: []rune{},
			maxWidth:  1000,
			expected:  []Line{emptyString},
		},
		{
			// Wrapping the empty string should produce a run with no glyphs/runes but
			// with valid line bounds (even if the available width is zero).
			name:      "empty string zero width",
			shaped:    emptyString,
			paragraph: []rune{},
			maxWidth:  0,
			expected:  []Line{emptyString},
		},
		{
			// This test case verifies that no line breaks occur if they are not
			// necessary, and that the proper Offsets are reported in the output.
			name:      "all one line",
			shaped:    []Output{shapedText1},
			paragraph: []rune(text1),
			maxWidth:  1000,
			expected: []Line{
				[]Output{
					withRange(shapedText1, Range{
						Count: len([]rune(text1)),
					}),
				},
			},
		},
		{
			// This test case verifies that trailing whitespace characters on a
			// line do not just disappear if it's the first line.
			name:      "trailing whitespace",
			shaped:    []Output{shapedText1Trailing},
			paragraph: []rune(text1Trailing),
			maxWidth:  1000,
			expected: []Line{
				[]Output{
					withRange(shapedText1Trailing, Range{
						Count: len([]rune(text1)) + 1,
					}),
				},
			},
		},
		{
			// This test case verifies that the line wrapper rejects line break
			// candidates that would split a glyph cluster.
			name: "reject mid-cluster line breaks",
			shaped: []Output{
				{
					Advance: fixed.I(10 * 3),
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
						simpleGlyph(0),
						complexGlyph(1, 2, 2),
						complexGlyph(1, 2, 2),
					},
					Runes: Range{Count: 3},
				},
			},
			// This unicode data was discovered in a testing/quick failure
			// for widget.Editor. It has the property that the middle two
			// runes form a harfbuzz cluster but also have a legal UAX#14
			// segment break between them.
			paragraph: []rune{0xa8e58, 0x3a4fd, 0x119dd},
			maxWidth:  20,
			expected: []Line{
				[]Output{
					withRange(
						Output{
							Direction: di.DirectionLTR,
							Advance:   fixed.I(10),
							LineBounds: Bounds{
								Ascent:  fixed.I(10),
								Descent: fixed.I(5),
							},
							GlyphBounds: Bounds{
								Ascent: fixed.I(10),
							},
							Glyphs: []Glyph{
								simpleGlyph(0),
							},
						},
						Range{
							Count: 1,
						},
					),
				},
				[]Output{
					withRange(
						Output{
							Direction: di.DirectionLTR,
							Advance:   fixed.I(20),
							LineBounds: Bounds{
								Ascent:  fixed.I(10),
								Descent: fixed.I(5),
							},
							GlyphBounds: Bounds{
								Ascent: fixed.I(10),
							},
							Glyphs: []Glyph{
								complexGlyph(1, 2, 2),
								complexGlyph(1, 2, 2),
							},
						},
						Range{
							Count:  2,
							Offset: 1,
						},
					),
				},
			},
		},
		{
			// This test case verifies that the line wrapper rejects line break
			// candidates that would split a glyph cluster at non-zero offsets
			// within the shaped text.
			name: "reject mid-cluster line breaks at non-zero offsets",
			shaped: []Output{
				{
					Advance:     fixed.I(10),
					LineBounds:  Bounds{Ascent: fixed.I(10), Descent: fixed.I(5)},
					GlyphBounds: Bounds{Ascent: fixed.I(10)},
					Glyphs: []Glyph{
						simpleGlyph(0),
					},
					Runes: Range{Count: 1},
				},
				{
					Advance:     fixed.I(10),
					LineBounds:  Bounds{Ascent: fixed.I(10), Descent: fixed.I(5)},
					GlyphBounds: Bounds{Ascent: fixed.I(10)},
					Glyphs: []Glyph{
						simpleGlyph(1),
					},
					Runes: Range{Count: 1, Offset: 1},
				},
				{
					Advance:     fixed.I(10 * 2),
					LineBounds:  Bounds{Ascent: fixed.I(10), Descent: fixed.I(5)},
					GlyphBounds: Bounds{Ascent: fixed.I(10)},
					Glyphs: []Glyph{
						complexGlyph(2, 2, 2),
						complexGlyph(2, 2, 2),
					},
					Runes: Range{Count: 2, Offset: 2},
				},
				{
					Advance:    fixed.I(10),
					LineBounds: Bounds{Ascent: fixed.I(10), Descent: fixed.I(5)},
					GlyphBounds: Bounds{
						Ascent: fixed.I(10),
					},
					Glyphs: []Glyph{
						simpleGlyph(4),
					},
					Runes: Range{Count: 1, Offset: 4},
				},
			},
			// This unicode data was discovered in a fuzz test failure
			// for Gio's text shaper.
			paragraph: []rune{1593, 48, 32, 1474, 48},
			maxWidth:  40,
			expected: []Line{
				[]Output{
					{
						Advance:     fixed.I(10),
						LineBounds:  Bounds{Ascent: fixed.I(10), Descent: fixed.I(5)},
						GlyphBounds: Bounds{Ascent: fixed.I(10)},
						Glyphs: []Glyph{
							simpleGlyph(0),
						},
						Runes: Range{Count: 1},
					},
					{
						Advance:     fixed.I(10),
						LineBounds:  Bounds{Ascent: fixed.I(10), Descent: fixed.I(5)},
						GlyphBounds: Bounds{Ascent: fixed.I(10)},
						Glyphs: []Glyph{
							simpleGlyph(1),
						},
						Runes: Range{Count: 1, Offset: 1},
					},
					{
						Advance:     fixed.I(10 * 2),
						LineBounds:  Bounds{Ascent: fixed.I(10), Descent: fixed.I(5)},
						GlyphBounds: Bounds{Ascent: fixed.I(10)},
						Glyphs: []Glyph{
							complexGlyph(2, 2, 2),
							complexGlyph(2, 2, 2),
						},
						Runes: Range{Count: 2, Offset: 2},
					},
					{
						Advance:     fixed.I(10),
						LineBounds:  Bounds{Ascent: fixed.I(10), Descent: fixed.I(5)},
						GlyphBounds: Bounds{Ascent: fixed.I(10)},
						Glyphs: []Glyph{
							simpleGlyph(4),
						},
						Runes: Range{Count: 1, Offset: 4},
					},
				},
			},
		},
		{
			// This test case verifies that line breaking does occur, and that
			// all lines have proper offsets.
			name:      "line break on last word",
			shaped:    []Output{shapedText1},
			paragraph: []rune(text1),
			maxWidth:  120,
			expected: []Line{
				[]Output{
					withRange(
						splitShapedAt(shapedText1, 12)[0],
						Range{
							Count: len([]rune(text1)) - 3,
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText1, 12)[1],
						Range{
							Offset: len([]rune(text1)) - 3,
							Count:  3,
						},
					),
				},
			},
		},
		{
			// This test case verifies that many line breaks still result in
			// correct offsets. This test also ensures that leading whitespace
			// is correctly hidden on lines after the first.
			name:      "line break several times",
			shaped:    []Output{shapedText1},
			paragraph: []rune(text1),
			maxWidth:  70,
			expected: []Line{
				[]Output{
					withRange(
						splitShapedAt(shapedText1, 5)[0],
						Range{
							Count: 5,
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText1, 5, 12)[1],
						Range{
							Offset: 5,
							Count:  7,
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText1, 12)[1],
						Range{
							Offset: 12,
							Count:  3,
						},
					),
				},
			},
		},
		{
			// This test case verifies baseline offset math for more complicated input.
			name:      "all one line 2",
			shaped:    []Output{shapedText2},
			paragraph: []rune(text2),
			maxWidth:  1000,
			expected: []Line{
				[]Output{
					withRange(
						shapedText2,
						Range{
							Count: len([]rune(text2)),
						},
					),
				},
			},
		},
		{
			// This test case verifies that offset accounting correctly handles complex
			// input across line breaks. It is legal to line-break within words composed
			// of more than one script, so this test expects that to occur.
			name:      "line break several times 2",
			shaped:    []Output{shapedText2},
			paragraph: []rune(text2),
			maxWidth:  40,
			expected: []Line{
				[]Output{
					withRange(
						splitShapedAt(shapedText2, 4)[0],
						Range{
							Count: len([]rune("안П你 ")),
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText2, 4, 8)[1],
						Range{
							Count:  len([]rune("ligDROP 안П")),
							Offset: len([]rune("안П你 ")),
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText2, 8, 11)[1],
						Range{
							Count:  len([]rune("你 ligDROP")),
							Offset: len([]rune("안П你 ligDROP 안П")),
						},
					),
				},
			},
		},
		{
			// This test case verifies baseline offset math for complex RTL input.
			name:      "all one line 3",
			shaped:    []Output{shapedText3},
			paragraph: []rune(text3),
			maxWidth:  1000,
			expected: []Line{
				[]Output{
					withRange(
						shapedText3,
						Range{
							Count: len([]rune(text3)),
						},
					),
				},
			},
		},
		{
			// This test case verifies line wrapping logic in RTL mode.
			name:      "line break once [RTL]",
			shaped:    []Output{shapedText3},
			paragraph: []rune(text3),
			maxWidth:  100,
			expected: []Line{
				[]Output{
					withRange(
						splitShapedAt(shapedText3, 7)[1],
						Range{
							Count: len([]rune("שלום أهلا ")),
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText3, 7)[0],
						Range{
							Count:  len([]rune("שלום أهلا")),
							Offset: len([]rune("שלום أهلا ")),
						},
					),
				},
			},
		},
		{
			// This test case verifies line wrapping logic in RTL mode.
			name:      "line break several times [RTL]",
			shaped:    []Output{shapedText3},
			paragraph: []rune(text3),
			maxWidth:  50,
			expected: []Line{
				[]Output{
					withRange(
						splitShapedAt(shapedText3, 10)[1],
						Range{
							Count: len([]rune("שלום ")),
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText3, 7, 10)[1],
						Range{
							Count:  len([]rune("أهلا ")),
							Offset: len([]rune("שלום ")),
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText3, 2, 7)[1],
						Range{
							Count:  len([]rune("שלום ")),
							Offset: len([]rune("שלום أهلا ")),
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText3, 2)[0],
						Range{
							Count:  len([]rune("أهلا")),
							Offset: len([]rune("שלום أهلا שלום ")),
						},
					),
				},
			},
		},
		{
			// This test case verifies the behavior of the line wrapper for multi-run
			// shaped input.
			name:      "multiple input runs 1",
			shaped:    splitShapedMultiInput1,
			paragraph: []rune(multiInputText1),
			maxWidth:  20,
			expected: []Line{
				splitShapedAt(shapedMultiInputText1, 3)[:1],
				splitShapedAt(shapedMultiInputText1, 3, 4, 6)[1:3],
				splitShapedAt(shapedMultiInputText1, 6)[1:],
			},
		},
		{
			name:      "multiple input runs 2",
			shaped:    splitShapedMultiInput1,
			paragraph: []rune(multiInputText1),
			maxWidth:  30,
			expected: []Line{
				splitShapedAt(shapedMultiInputText1, 3)[:1],
				splitShapedAt(shapedMultiInputText1, 3, 4, 6)[1:3],
				splitShapedAt(shapedMultiInputText1, 6)[1:],
			},
		},
		{
			name:      "multiple input runs 3",
			shaped:    splitShapedMultiInput1,
			paragraph: []rune(multiInputText1),
			maxWidth:  40,
			expected: []Line{
				splitShapedAt(shapedMultiInputText1, 3)[:1],
				splitShapedAt(shapedMultiInputText1, 3, 4, 6)[1:3],
				splitShapedAt(shapedMultiInputText1, 6)[1:],
			},
		},
		{
			name:      "multiple input runs 4",
			shaped:    splitShapedMultiInput1,
			paragraph: []rune(multiInputText1),
			maxWidth:  50,
			expected: []Line{
				splitShapedAt(shapedMultiInputText1, 3)[:1],
				splitShapedAt(shapedMultiInputText1, 3, 4, 6)[1:],
			},
		},
		{
			name:      "multiple input runs 5",
			shaped:    splitShapedMultiInput1,
			paragraph: []rune(multiInputText1),
			maxWidth:  60,
			expected: []Line{
				splitShapedAt(shapedMultiInputText1, 4, 6)[:2],
				splitShapedAt(shapedMultiInputText1, 6)[1:],
			},
		},
		{
			name:      "multiple input runs 6",
			shaped:    splitShapedMultiInput1,
			paragraph: []rune(multiInputText1),
			maxWidth:  70,
			expected: []Line{
				splitShapedAt(shapedMultiInputText1, 4, 6)[:2],
				splitShapedAt(shapedMultiInputText1, 6)[1:],
			},
		},
		{
			name:      "multiple input runs 7",
			shaped:    splitShapedMultiInput1,
			paragraph: []rune(multiInputText1),
			maxWidth:  80,
			expected: []Line{
				splitShapedMultiInput1,
			},
		},
		{
			// This is regression test. One of the benchmarks would hit this case in which
			// the line wrap candidate was exactly at the end of the available runs and
			// would increment the currentRun off of the end of the available runs.
			name:      "one word",
			shaped:    []Output{shapedOneWord},
			paragraph: []rune(oneWord),
			maxWidth:  10,
			expected: []Line{
				{
					shapedOneWord,
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var l LineWrapper
			outs, _ := l.WrapParagraph(WrapConfig{BreakPolicy: Never}, tc.maxWidth, tc.paragraph, NewSliceIterator(tc.shaped))

			if len(tc.expected) != len(outs) {
				t.Errorf("expected %d lines, got %d", len(tc.expected), len(outs))
				return
			}
			for lineNum := range tc.expected {
				expectedLine := tc.expected[lineNum]
				actualLine := outs[lineNum]
				if len(expectedLine) != len(actualLine) {
					t.Errorf("expected %d runs in line %d, got %d", len(expectedLine), lineNum, len(actualLine))
					continue
				}
				for runNum := range expectedLine {
					expectedRun := expectedLine[runNum]
					actualRun := actualLine[runNum]
					lenE := len(expectedRun.Glyphs)
					lenO := len(actualRun.Glyphs)
					if lenE != lenO {
						t.Errorf("line %d run %d: expected %d glyphs, got %d", lineNum, runNum, lenE, lenO)
						continue
					}
					for k := range expectedRun.Glyphs {
						e := expectedRun.Glyphs[k]
						o := actualRun.Glyphs[k]
						if !reflect.DeepEqual(e, o) {
							t.Errorf("line %d: glyph mismatch at index %d, expected: %#v, got %#v", runNum, k, e, o)
						}
					}
					if expectedRun.Runes != actualRun.Runes {
						t.Errorf("line %d: expected %#v offsets, got %#v", runNum, expectedRun.Runes, actualRun.Runes)
					}
					if expectedRun.Direction != actualRun.Direction {
						t.Errorf("line %d: expected %v direction, got %v", runNum, expectedRun.Direction, actualRun.Direction)
					}
					// Reduce the verbosity of the reflect mismatch since we already
					// compared the glyphs.
					expectedRun.Glyphs = nil
					actualRun.Glyphs = nil
					actualRun.VisualIndex = expectedRun.VisualIndex
					if !reflect.DeepEqual(expectedRun, actualRun) {
						t.Errorf("line %d: expected: %#v, got %#v", runNum, expectedRun, actualRun)
					}
				}
			}
		})
	}
}

// simpleGlyph returns a simple square glyph with the provided cluster
// value.
func simpleGlyph(cluster int) Glyph {
	return complexGlyph(cluster, 1, 1)
}

// ligatureGlyph returns a simple square glyph with the provided cluster
// value and number of runes.
func ligatureGlyph(cluster, runes int) Glyph {
	return complexGlyph(cluster, runes, 1)
}

// expansionGlyph returns a simple square glyph with the provided cluster
// value and number of glyphs.
func expansionGlyph(cluster, glyphs int) Glyph {
	return complexGlyph(cluster, 1, glyphs)
}

// complexGlyph returns a simple square glyph with the provided cluster
// value, number of associated runes, and number of glyphs in the cluster.
func complexGlyph(cluster, runes, glyphs int) Glyph {
	return Glyph{
		Width:        fixed.I(10),
		Height:       fixed.I(10),
		Advance:      fixed.I(10),
		YBearing:     fixed.I(10),
		ClusterIndex: cluster,
		GlyphCount:   glyphs,
		RuneCount:    runes,
	}
}
