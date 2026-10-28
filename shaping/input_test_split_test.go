package shaping

import (
	"testing"

	"github.com/OpenMellon/typesetting/di"
	"github.com/OpenMellon/typesetting/font"
	"github.com/OpenMellon/typesetting/language"
	tu "github.com/OpenMellon/typesetting/testutils"
)

func TestSplit(t *testing.T) {
	latinFont := loadOpentypeFont(t, "../font/testdata/Roboto-Regular.ttf")
	arabicFont := loadOpentypeFont(t, "../font/testdata/Amiri-Regular.ttf")
	fm := fixedFontmap{latinFont, arabicFont}

	sideways, upright := di.DirectionTTB, di.DirectionTTB
	sideways.SetSideways(true)
	upright.SetSideways(false)

	var seg Segmenter

	type run struct {
		start, end int
		dir        di.Direction
		script     language.Script
		lang       language.Language
		face       *font.Face
	}
	for _, test := range []struct {
		text         string
		dir          di.Direction
		expectedRuns []run
	}{
		{
			"",
			di.DirectionLTR,
			[]run{{0, 0, di.DirectionLTR, language.Common, "fr", nil}},
		},
		{ // make sure we dont crash
			"\n",
			di.DirectionLTR,
			[]run{{0, 1, di.DirectionLTR, language.Common, "fr", latinFont}},
		},
		{
			"The quick brown fox jumps over the lazy dog.",
			di.DirectionLTR,
			[]run{{0, 44, di.DirectionLTR, language.Latin, "fr", latinFont}},
		},
		{
			"الحب سماء لا تمط غير الأحلام",
			di.DirectionLTR,
			[]run{{0, 28, di.DirectionRTL, language.Arabic, "ar", arabicFont}},
		},
		{
			"The quick سماء שלום لا fox تمط שלום غير the lazy dog.",
			di.DirectionLTR,
			[]run{
				{0, 10, di.DirectionLTR, language.Latin, "fr", latinFont},
				{10, 15, di.DirectionRTL, language.Arabic, "ar", arabicFont},
				{15, 20, di.DirectionRTL, language.Hebrew, "he", latinFont},
				{20, 22, di.DirectionRTL, language.Arabic, "ar", arabicFont},
				{22, 27, di.DirectionLTR, language.Latin, "fr", latinFont},
				{27, 31, di.DirectionRTL, language.Arabic, "ar", arabicFont},
				{31, 36, di.DirectionRTL, language.Hebrew, "he", latinFont},
				{36, 39, di.DirectionRTL, language.Arabic, "ar", arabicFont},
				{39, 53, di.DirectionLTR, language.Latin, "fr", latinFont},
			},
		},
		{
			"الحب سماء brown привет fox تمط jumps привет over غير الأحلام",
			di.DirectionLTR,
			[]run{
				{0, 10, di.DirectionRTL, language.Arabic, "ar", arabicFont},
				{10, 16, di.DirectionLTR, language.Latin, "fr", latinFont},
				{16, 23, di.DirectionLTR, language.Cyrillic, "ru", latinFont},
				{23, 26, di.DirectionLTR, language.Latin, "fr", latinFont},
				{26, 31, di.DirectionRTL, language.Arabic, "ar", arabicFont},
				{31, 37, di.DirectionLTR, language.Latin, "fr", latinFont},
				{37, 44, di.DirectionLTR, language.Cyrillic, "ru", latinFont},
				{44, 48, di.DirectionLTR, language.Latin, "fr", latinFont},
				{48, 60, di.DirectionRTL, language.Arabic, "ar", arabicFont},
			},
		},
		// vertical text
		{
			"A french word",
			di.DirectionTTB,
			[]run{
				{0, 13, sideways, language.Latin, "fr", latinFont},
			},
		},
		{
			"A french word",
			upright,
			[]run{
				{0, 13, upright, language.Latin, "fr", latinFont},
			},
		},
		{
			"with upright \uff21\uff22\uff23",
			di.DirectionTTB,
			[]run{
				{0, 13, sideways, language.Latin, "fr", latinFont},
				{13, 16, upright, language.Latin, "fr", latinFont},
			},
		},
		{
			"ᠬᠦᠮᠦᠨ ᠪᠦ",
			di.DirectionTTB,
			[]run{
				{0, 8, sideways, language.Mongolian, "mn", latinFont},
			},
		},
		{
			"\u200fabc", // the mark should not trigger a face change, even "across" different directions/scripts
			di.DirectionLTR,
			[]run{
				{0, 1, di.DirectionRTL, language.Common, "fr", latinFont},
				{1, 4, di.DirectionLTR, language.Latin, "fr", latinFont},
			},
		},
		{
			"a\u200fabc", // the mark should not trigger a face change, even "across" different directions/scripts
			di.DirectionLTR,
			[]run{
				{0, 1, di.DirectionLTR, language.Latin, "fr", latinFont},
				{1, 2, di.DirectionRTL, language.Common, "fr", latinFont},
				{2, 5, di.DirectionLTR, language.Latin, "fr", latinFont},
			},
		},
	} {
		inputs := seg.Split(Input{
			Text:      []rune(test.text),
			RunEnd:    len([]rune(test.text)),
			Direction: test.dir,

			Size:     10,
			Language: "fr",
		}, fm)
		tu.Assert(t, len(inputs) == len(test.expectedRuns))
		for i, run := range test.expectedRuns {
			got := inputs[i]
			tu.Assert(t, got.RunStart == run.start)
			tu.Assert(t, got.RunEnd == run.end)
			tu.Assert(t, got.Direction == run.dir)
			tu.Assert(t, got.Script == run.script)
			tu.Assert(t, got.Face == run.face)
			tu.Assert(t, got.Language == run.lang)
			// check that input properties are properly copied
			tu.Assert(t, got.Size == 10)
		}

		// check that spliting a "middle" text slice is supported
		inputs = seg.Split(Input{
			Text:      []rune("DUMMY" + test.text + "DUMMY"),
			RunStart:  5,
			RunEnd:    5 + len([]rune(test.text)),
			Direction: test.dir,
		}, fm)
		tu.Assert(t, len(inputs) == len(test.expectedRuns))
		for i, run := range test.expectedRuns {
			got := inputs[i]
			tu.Assert(t, got.RunStart == 5+run.start)
			tu.Assert(t, got.RunEnd == 5+run.end)
			tu.Assert(t, got.Direction == run.dir)
			tu.Assert(t, got.Script == run.script)
			tu.Assert(t, got.Face == run.face)
		}
	}

	// unknown language are preserved
	inputs := seg.Split(Input{
		Text:     []rune("abc"),
		RunEnd:   len([]rune("abc")),
		Language: "xxxx", // unknown
	}, fm)
	tu.Assert(t, len(inputs) == 1)
	tu.Assert(t, inputs[0].Language == "xxxx")

	// empty language are resolved
	inputs = seg.Split(Input{
		Text:     []rune("سماء"),
		RunEnd:   len([]rune("سماء")),
		Language: "", // empty
	}, fm)
	tu.Assert(t, len(inputs) == 1)
	tu.Assert(t, inputs[0].Language == "ar")
}

func TestIssue127(t *testing.T) {
	// regression test for https://github.com/go-text/typesetting/issues/127
	str := []rune("لمّا")
	input := Input{
		Text:      str,
		RunStart:  0,
		RunEnd:    len(str),
		Direction: di.DirectionRTL,
		Language:  language.NewLanguage("ar"),
	}

	inputs := (&Segmenter{}).Split(input, fixedFontmap{benchArFace})
	// make sure Inherited script does no create a new run
	tu.Assert(t, len(inputs) == 1)
}

func Test_enforceLang(t *testing.T) {
	tests := []struct {
		lang language.LangID
		s    language.Script
		want language.LangID
	}{
		{language.LangFr, language.Latin, language.LangFr},
		{language.LangFr, language.Hangul, language.LangKo},
		{language.LangKo, language.Hangul, language.LangKo},
		{language.LangJa, language.Hiragana, language.LangJa},
		{language.LangEn, language.Arabic, language.LangAr},
		{language.LangEn, language.Bopomofo, language.LangEn},
	}
	for _, tt := range tests {
		if got := enforceLang(tt.lang, tt.s); got != tt.want {
			t.Errorf("enforceLang() = %v, want %v", got, tt.want)
		}
	}
}
