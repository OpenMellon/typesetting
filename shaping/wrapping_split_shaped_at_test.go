package shaping

import (
	"fmt"
	"reflect"
	"testing"

	tu "github.com/OpenMellon/typesetting/testutils"
)

// splitShapedAt splits a single shaped output into multiple. It splits
// on each provided glyph index in indices, with the index being the end of
// a slice range (so it's exclusive). You can think of the index as the
// first glyph of the next output.
func splitShapedAt(shaped Output, indices ...glyphIndex) []Output {
	numOut := len(indices) + 1
	outputs := make([]Output, 0, numOut)
	start := 0
	runeOffset := shaped.Runes.Offset
	for _, i := range indices {
		newOut := shaped
		newOut.Glyphs = newOut.Glyphs[start:i]
		newOut.Runes.Offset = runeOffset
		newOut.Runes.Count = 0
		cluster := -1
		for _, g := range newOut.Glyphs {
			if cluster == g.ClusterIndex {
				continue
			}
			cluster = g.ClusterIndex
			newOut.Runes.Count += g.RuneCount
		}
		runeOffset += newOut.Runes.Count
		newOut.RecalculateAll()
		outputs = append(outputs, newOut)
		start = i
	}
	newOut := shaped
	newOut.Glyphs = newOut.Glyphs[start:]
	newOut.Runes.Offset = runeOffset
	newOut.Runes.Count = shaped.Runes.Count + shaped.Runes.Offset - newOut.Runes.Offset
	newOut.RecalculateAll()
	outputs = append(outputs, newOut)
	return outputs
}

func TestWrapLine(t *testing.T) {
	type expected struct {
		line     Line
		nextLine int
		done     bool
	}
	type testcase struct {
		name      string
		shaped    []Output
		paragraph []rune
		maxWidth  int
		expected  []expected
	}
	for _, tc := range []testcase{
		{
			name:      "simple",
			shaped:    []Output{shapedText1},
			paragraph: []rune(text1),
			maxWidth:  40,
			expected: []expected{
				{
					line:     []Output{splitShapedAt(shapedText1, 5)[0]},
					nextLine: 5,
					done:     false,
				},
				{
					line:     []Output{splitShapedAt(shapedText1, 5, 9)[1]},
					nextLine: 9,
					done:     false,
				},
				{
					line:     []Output{splitShapedAt(shapedText1, 9, 12)[1]},
					nextLine: 12,
					done:     false,
				},
				{
					line:     []Output{splitShapedAt(shapedText1, 12)[1]},
					nextLine: 15,
					done:     true,
				},
			},
		},
		{
			// This test uses the same input text as the previous, but chops
			// every glyph into its own run. This simulates text that changes
			// font or style every glyph.
			name:      "simple in pieces 1",
			shaped:    splitShapedAt(shapedText1, 1, 2, 3, 4, 5),
			paragraph: []rune(text1),
			maxWidth:  40,
			expected: []expected{
				{
					line:     splitShapedAt(shapedText1, 1, 2, 3, 4, 5)[:5],
					nextLine: 5,
					done:     false,
				},
			},
		},
		{
			// This test uses the same test strategy as the previous, but divides
			// the run into segments that do not align evenly with line break
			// candidates. This forces the wrapper to break one run across lines.
			name:      "simple in pieces 2",
			shaped:    splitShapedAt(shapedText1, 3, 6),
			paragraph: []rune(text1),
			maxWidth:  40,
			expected: []expected{
				{
					line:     splitShapedAt(shapedText1, 3, 5)[:2],
					nextLine: 5,
					done:     false,
				},
				{
					line:     splitShapedAt(shapedText1, 5, 6, 9)[1:3],
					nextLine: 9,
					done:     false,
				},
			},
		},
		{
			name:      "simple rtl",
			shaped:    []Output{shapedText3},
			paragraph: []rune(text3),
			maxWidth:  40,
			expected: []expected{
				{
					line: []Output{
						withRange(splitShapedAt(shapedText3, 10)[1],
							Range{Count: 5}),
					},
					nextLine: 5,
					done:     false,
				},
				{
					line: []Output{
						withRange(splitShapedAt(shapedText3, 7, 10)[1],
							Range{Offset: 5, Count: 5}),
					},
					nextLine: 10,
					done:     false,
				},
				{
					line: []Output{
						withRange(splitShapedAt(shapedText3, 2, 7)[1],
							Range{Offset: 10, Count: 5}),
					},
					nextLine: 15,
					done:     false,
				},
				{
					line: []Output{
						withRange(splitShapedAt(shapedText3, 2)[0],
							Range{Offset: 15, Count: 4}),
					},
					nextLine: 19,
					done:     true,
				},
			},
		},
		{
			name:      "simple bidi",
			shaped:    shapedBidiText1,
			paragraph: []rune(bidiText1),
			maxWidth:  110,
			expected: []expected{
				{
					line: []Output{
						shapedBidiText1[0],
						withRange(splitShapedAt(shapedBidiText1[1], 5)[1],
							Range{Offset: 6, Count: 5}),
					},
					nextLine: 11,
					done:     false,
				},
				{
					line: []Output{
						withRange(splitShapedAt(shapedBidiText1[1], 5)[0],
							Range{Offset: 11, Count: 5}),
						shapedBidiText1[2],
					},
					nextLine: 20,
					done:     true,
				},
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
			expected: []expected{
				{
					line: []Output{
						shapedOneWord,
					},
					nextLine: 4,
					done:     true,
				},
			},
		},
		{
			name:      "simple fast-path",
			shaped:    []Output{shapedText1},
			paragraph: []rune(text1),
			maxWidth:  200,
			expected: []expected{
				{
					line:     Line{shapedText1},
					nextLine: 15,
					done:     true,
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				line WrappedLine
				done bool
				l    LineWrapper
			)
			l.Prepare(WrapConfig{BreakPolicy: Never}, tc.paragraph, NewSliceIterator(tc.shaped))
			// Iterate every line declared in the test case expectations. This
			// allows test cases to be exhaustive if they need to wihtout forcing
			// every case to wrap entire paragraphs.
			for lineNumber, expected := range tc.expected {
				line, done = l.WrapNextLine(tc.maxWidth)
				compareLines(t, lineNumber, expected.line, line.Line)
				if done != expected.done {
					t.Errorf("done mismatch! expected %v, got %v", expected.done, done)
				}

				tu.AssertC(t, line.NextLine == expected.nextLine, fmt.Sprintf("expected %d, got %d", expected.nextLine, line.NextLine))

				if expected.done { // check WrapNextLine is now a no-op
					line, done = l.WrapNextLine(200)
					if line.Line != nil || !done {
						t.Errorf("expect nil output for WrapNextLine, got %p and %v", line.Line, done)
					}
				}
			}
		})
	}
}

func compareLines(t *testing.T, lineNumber int, expected, actual Line) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Errorf("line %d: expected %d runs, got %d", lineNumber, len(expected), len(actual))
		return
	}
	for i := range expected {
		expected := expected[i]
		actual := actual[i]
		if len(expected.Glyphs) != len(actual.Glyphs) {
			t.Errorf("line %d: run %d: expected %d glyphs, got %d", lineNumber, i, len(expected.Glyphs), len(actual.Glyphs))
			return
		}
		for ii := range expected.Glyphs {
			eg := expected.Glyphs[ii]
			ag := actual.Glyphs[ii]
			if eg != ag {
				t.Errorf("line: %d: run %d: glyph %d: expected\n%#+v\ngot\n%#+v", lineNumber, i, ii, eg, ag)
			}
		}
		expected.Glyphs = nil
		actual.Glyphs = nil
		actual.VisualIndex = expected.VisualIndex
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("line %d: run %d: expected\n%#+v\ngot\n%#+v", lineNumber, i, expected, actual)
		}
	}
}
