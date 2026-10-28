package shaping

import (
	"os"
	"reflect"
	"testing"
	"unicode"

	"github.com/OpenMellon/typesetting/di"
	"github.com/OpenMellon/typesetting/font"
	"github.com/OpenMellon/typesetting/language"
	tu "github.com/OpenMellon/typesetting/testutils"
)

func Test_ignoreFaceChange(t *testing.T) {
	tests := []struct {
		args rune
		want bool
	}{
		{' ', true},
		{'a', false},
		{'\n', true},
		{'\r', true},
		{'\f', true},
		{'\ufe01', true},
		{'\ufe02', true},
		{'\U000E0100', true},
		{'\u06DD', false},
		{'\u200f', true},
	}
	for _, tt := range tests {
		if got := ignoreFaceChange(tt.args); got != tt.want {
			t.Errorf("ignoreFaceChange() = %v, want %v", got, tt.want)
		}
	}
}

// support any rune
type universalCmap struct{ font.Cmap }

func (universalCmap) Lookup(rune) (font.GID, bool) { return 0, true }

type upperCmap struct{ font.Cmap }

func (upperCmap) Lookup(r rune) (font.GID, bool) {
	return 0, unicode.IsUpper(r)
}

type lowerCmap struct{ font.Cmap }

func (lowerCmap) Lookup(r rune) (font.GID, bool) {
	return 0, unicode.IsLower(r)
}

func loadOpentypeFont(t testing.TB, filename string) *font.Face {
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("opening font file: %s", err)
	}
	face, err := font.ParseTTF(file)
	if err != nil {
		t.Fatalf("parsing font file %s: %s", filename, err)
	}
	return face
}

func TestSplitByFontGlyphs(t *testing.T) {
	type args struct {
		input          Input
		availableFaces []*font.Face
	}

	universalFont := font.NewFace(&font.Font{Cmap: universalCmap{}})
	lowerFont := font.NewFace(&font.Font{Cmap: lowerCmap{}})
	upperFont := font.NewFace(&font.Font{Cmap: upperCmap{}})

	latinFont := loadOpentypeFont(t, "../font/testdata/Roboto-Regular.ttf")
	arabicFont := loadOpentypeFont(t, "../font/testdata/Amiri-Regular.ttf")
	englishArabic := []rune("Hello " + "تثذرزسشص" + "world" + "لمنهويء")

	tests := []struct {
		name string
		args args
		want []Input
	}{
		{
			"no font change",
			args{
				input: Input{
					Text:     []rune("a simple text"),
					RunStart: 0, RunEnd: len("a simple text"),
				},
				availableFaces: []*font.Face{universalFont},
			},
			[]Input{
				{
					Text:     []rune("a simple text"),
					RunStart: 0, RunEnd: len("a simple text"),
					Face: universalFont,
				},
			},
		},
		{
			"one change no spaces",
			args{
				input: Input{
					Text:     []rune("aaaAAA"),
					RunStart: 0, RunEnd: len("aaaAAA"),
				},
				availableFaces: []*font.Face{lowerFont, upperFont},
			},
			[]Input{
				{
					Text:     []rune("aaaAAA"),
					RunStart: 0, RunEnd: 3,
					Face: lowerFont,
				},
				{
					Text:     []rune("aaaAAA"),
					RunStart: 3, RunEnd: 6,
					Face: upperFont,
				},
			},
		},
		{
			"one change with spaces",
			args{
				input: Input{
					Text:     []rune("aaa AAA "),
					RunStart: 0, RunEnd: len("aaa AAA "),
				},
				availableFaces: []*font.Face{lowerFont, upperFont},
			},
			[]Input{
				{
					Text:     []rune("aaa AAA "),
					RunStart: 0, RunEnd: 4,
					Face: lowerFont,
				},
				{
					Text:     []rune("aaa AAA "),
					RunStart: 4, RunEnd: 8,
					Face: upperFont,
				},
			},
		},
		{
			"no font matched 1",
			args{
				input: Input{
					Text:     []rune("__"),
					RunStart: 0, RunEnd: len("__"),
				},
				availableFaces: []*font.Face{lowerFont, upperFont},
			},
			[]Input{
				{
					Text:     []rune("__"),
					RunStart: 0, RunEnd: 2,
					Face: lowerFont,
				},
			},
		},
		{
			"no font matched 2",
			args{
				input: Input{
					Text:     []rune("__"),
					RunStart: 0, RunEnd: len("__"),
				},
				availableFaces: []*font.Face{upperFont, lowerFont},
			},
			[]Input{
				{
					Text:     []rune("__"),
					RunStart: 0, RunEnd: 2,
					Face: upperFont,
				},
			},
		},
		{
			"mixed english arabic",
			args{
				input: Input{
					Text:     englishArabic,
					RunStart: 0, RunEnd: len(englishArabic),
				},
				availableFaces: []*font.Face{latinFont, arabicFont},
			},
			[]Input{
				{
					Text:     englishArabic,
					RunStart: 0, RunEnd: 6,
					Face: latinFont,
				},
				{
					Text:     englishArabic,
					RunStart: 6, RunEnd: 14,
					Face: arabicFont,
				},
				{
					Text:     englishArabic,
					RunStart: 14, RunEnd: 19,
					Face: latinFont,
				},
				{
					Text:     englishArabic,
					RunStart: 19, RunEnd: 26,
					Face: arabicFont,
				},
			},
		},
		{
			"no change on starting space",
			args{
				input: Input{
					Text:     []rune(" غير الأحلام"),
					RunStart: 0, RunEnd: len([]rune(" غير الأحلام")),
				},
				availableFaces: []*font.Face{latinFont, arabicFont},
			},
			[]Input{
				{
					Text:     []rune(" غير الأحلام"),
					RunStart: 0, RunEnd: len([]rune(" غير الأحلام")),
					Face: arabicFont,
				},
			},
		},
		{
			"entirely whitespace",
			args{
				input: Input{
					Text:     []rune("   "),
					RunStart: 0, RunEnd: 3,
				},
				availableFaces: []*font.Face{latinFont, arabicFont},
			},
			[]Input{
				{
					Text:     []rune("   "),
					RunStart: 0, RunEnd: 3,
					Face: latinFont,
				},
			},
		},
		{
			"no change on ending space",
			args{
				input: Input{
					Text:     []rune(" غير الأحلام "),
					RunStart: 0, RunEnd: len([]rune(" غير الأحلام ")),
				},
				availableFaces: []*font.Face{latinFont, arabicFont},
			},
			[]Input{
				{
					Text:     []rune(" غير الأحلام "),
					RunStart: 0, RunEnd: len([]rune(" غير الأحلام ")),
					Face: arabicFont,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SplitByFontGlyphs(tt.args.input, tt.args.availableFaces); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitByFontGlyphs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitBidi(t *testing.T) {
	ltrSource := []rune("The quick brown fox jumps over the lazy dog.")
	rtlSource := []rune("الحب سماء لا تمط غير الأحلام")
	bidiSource := []rune("The quick سماء שלום لا fox تمط שלום غير the lazy dog.")
	bidi2Source := []rune("الحب سماء brown привет fox تمط jumps привет over غير الأحلام")
	type run struct {
		start, end int
		dir        di.Direction
	}
	for _, test := range []struct {
		text             []rune
		defaultDirection di.Direction
		expectedRuns     []run
	}{
		{
			text:             ltrSource,
			defaultDirection: di.DirectionLTR,
			expectedRuns: []run{
				{0, len(ltrSource), di.DirectionLTR},
			},
		},
		{
			text:             ltrSource,
			defaultDirection: di.DirectionTTB,
			expectedRuns: []run{
				{0, len(ltrSource), di.DirectionTTB},
			},
		},
		{
			text:             ltrSource,
			defaultDirection: di.DirectionRTL,
			expectedRuns: []run{
				{0, len(ltrSource) - 1, di.DirectionLTR},
				{len(ltrSource) - 1, len(ltrSource), di.DirectionRTL},
			},
		},
		{
			text:             ltrSource,
			defaultDirection: di.DirectionBTT,
			expectedRuns: []run{
				{0, len(ltrSource) - 1, di.DirectionTTB},
				{len(ltrSource) - 1, len(ltrSource), di.DirectionBTT},
			},
		},
		{
			text:             rtlSource,
			defaultDirection: di.DirectionRTL,
			expectedRuns: []run{
				{0, len(rtlSource), di.DirectionRTL},
			},
		},
		{
			text:             rtlSource,
			defaultDirection: di.DirectionBTT,
			expectedRuns: []run{
				{0, len(rtlSource), di.DirectionBTT},
			},
		},
		{
			text:             bidiSource,
			defaultDirection: di.DirectionLTR,
			expectedRuns: []run{
				// spaces are assigned to LTR runs
				{0, 10, di.DirectionLTR},
				{10, 22, di.DirectionRTL},
				{22, 27, di.DirectionLTR},
				{27, 39, di.DirectionRTL},
				{39, 53, di.DirectionLTR},
			},
		},
		{
			text:             bidi2Source,
			defaultDirection: di.DirectionLTR,
			// spaces are assigned to RTL runs
			expectedRuns: []run{
				{0, 10, di.DirectionRTL},
				{10, 26, di.DirectionLTR},
				{26, 31, di.DirectionRTL},
				{31, 48, di.DirectionLTR},
				{48, 60, di.DirectionRTL},
			},
		},
	} {
		var seg Segmenter
		seg.splitByBidi(Input{Text: test.text, RunEnd: len(test.text), Direction: test.defaultDirection})
		tu.AssertC(t, len(seg.output) == len(test.expectedRuns), string(test.text))
		for i, run := range test.expectedRuns {
			got := seg.output[i]
			tu.Assert(t, got.RunStart == run.start)
			tu.Assert(t, got.RunEnd == run.end)
			tu.Assert(t, got.Direction == run.dir)
		}
	}
}

func TestSplitScript(t *testing.T) {
	ltrSource := []rune("The quick brown fox jumps over the lazy dog.")
	rtlSource := []rune("الحب سماء لا تمط غير الأحلام")
	mixedLTRSource := []rune("The quick привет")
	mixedRTLSource := []rune("تمط שלום غي")
	commonSource := []rune("()[](][ gamma") // Common at first
	commonSource2 := []rune("gamma (Γ) est une lettre")
	commonSource3 := []rune("gamma (Γ [п] Γ) est une lettre") // nested delimiters
	withInherited := []rune("لمّا")
	type run struct {
		start, end int
		script     language.Script
	}
	for _, test := range []struct {
		text         []rune
		expectedRuns []run
	}{
		{ltrSource, []run{
			{0, len(ltrSource), language.Latin},
		}},
		{rtlSource, []run{
			{0, len(rtlSource), language.Arabic},
		}},
		{mixedLTRSource, []run{
			{0, 10, language.Latin},
			{10, 16, language.Cyrillic},
		}},
		{mixedRTLSource, []run{
			{0, 4, language.Arabic},
			{4, 9, language.Hebrew},
			{9, 11, language.Arabic},
		}},
		{commonSource, []run{
			{0, 13, language.Latin},
		}},
		{commonSource2, []run{
			{0, 7, language.Latin},
			{7, 8, language.Greek},
			{8, 24, language.Latin},
		}},
		{commonSource3, []run{
			{0, 7, language.Latin},
			{7, 10, language.Greek},
			{10, 11, language.Cyrillic},
			{11, 14, language.Greek},
			{14, 30, language.Latin},
		}},
		{withInherited, []run{
			{0, 4, language.Arabic},
		}},
	} {
		var seg Segmenter
		seg.splitByBidi(Input{Text: test.text, RunEnd: len(test.text), Direction: di.DirectionLTR})
		tu.Assert(t, len(seg.output) == 1)
		seg.input, seg.output = seg.output, seg.input

		seg.splitByScript()
		tu.Assert(t, len(seg.output) == len(test.expectedRuns))
		for i, run := range test.expectedRuns {
			got := seg.output[i]
			tu.Assert(t, got.RunStart == run.start)
			tu.Assert(t, got.RunEnd == run.end)
			tu.Assert(t, got.Script == run.script)
		}
	}

	// Verifies that Arabic diacritics (which usually have
	// script 'Inherited') are correctly clustered with their base Arabic letters,
	// rather than being split into a separate shaping run.
	for _, test := range []struct {
		text         []rune
		expectedRuns []run
	}{
		{
			// Arabic Letter + Diacritic
			// \u0628 => BEH
			// \u0650 => KASRA (Diacritic)
			[]rune{'\u0628', '\u0650'},
			[]run{{0, 2, language.Arabic}},
		},
		{
			// Arabic Word with Multiple Diacritics
			[]rune{
				'\u0628', // BEH
				'\u0650', // KASRA
				'\u0633', // SEEN
				'\u0652', // SUKUN
				'\u0645', // MEEM
				'\u0650', // KASRA
			},
			[]run{{0, 6, language.Arabic}},
		},
		{
			// Mixed Script (CONTROL Case) #1
			// Arabic Letter + Latin Letter
			// THESE MUST SPLIT TO 2.
			[]rune{'\u0628', 'A'},
			[]run{
				{0, 1, language.Arabic},
				{1, 2, language.Latin},
			},
		},
		{
			// Mixed Script (CONTROL Case) #2
			// Arabic Letter + Diacritic + Diacritic + Latin Letter + Arabic Letter + Diacritic
			// THESE MUST SPLIT TO 3.
			[]rune{'\u0628', '\u0651', '\u0650', 'A', '\u0628', '\u0650'},
			[]run{
				{0, 3, language.Arabic},
				{3, 4, language.Latin},
				{4, 6, language.Arabic},
			},
		},
		{
			// Mixed Script (A little 'stress' test)
			// Latin 's' + Arabic Kasra + Latin 'r' + Arabic Fatha
			// this is technically valid unicode!
			// the diacritics should inherit "Latin"
			[]rune{'s', '\u0651', '\u0650', 'r', '\u064E'},
			[]run{{0, 5, language.Latin}},
		},
	} {
		var seg Segmenter
		seg.splitByBidi(Input{Text: test.text, RunEnd: len(test.text), Direction: di.DirectionLTR})
		seg.input, seg.output = seg.output, seg.input

		seg.splitByScript()
		tu.Assert(t, len(seg.output) == len(test.expectedRuns))
		for i, run := range test.expectedRuns {
			got := seg.output[i]
			tu.Assert(t, got.RunStart == run.start)
			tu.Assert(t, got.RunEnd == run.end)
			tu.Assert(t, got.Script == run.script)
		}
	}
}

func TestSplitVertOrientation(t *testing.T) {
	type run struct {
		start, end int
		sideways   bool
	}
	for _, test := range []struct {
		text         []rune
		expectedRuns []run
	}{
		{
			[]rune("A regular latin sentence."),
			[]run{
				{0, 25, true},
			},
		},
		{
			[]rune("ごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみ"),
			[]run{
				{0, 44, false}, // Hiragana is upright
			},
		},
		{
			[]rune("ごさざ.しじす;ずせ"),
			[]run{
				{0, 10, false}, // Hiragana is upright
			},
		},
		{
			[]rune("もつ青いそら…"),
			[]run{
				{0, 2, false}, // Hiragana is upright
				{2, 3, false},
				{3, 7, false}, // Hiragana is upright
			},
		},
	} {
		var seg Segmenter
		seg.input = []Input{{Text: test.text, RunEnd: len(test.text), Direction: di.DirectionTTB}}

		seg.splitByScript()
		seg.input, seg.output = seg.output, seg.input
		seg.output = seg.output[:0]
		seg.splitByVertOrientation()
		tu.Assert(t, len(seg.output) == len(test.expectedRuns))
		for i, run := range test.expectedRuns {
			got := seg.output[i]
			tu.Assert(t, got.RunStart == run.start)
			tu.Assert(t, got.RunEnd == run.end)
			tu.Assert(t, got.Direction.HasVerticalOrientation())
			tu.Assert(t, got.Direction.IsSideways() == run.sideways)
		}
	}
}
