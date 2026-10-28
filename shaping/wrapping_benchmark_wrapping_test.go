package shaping

import (
	"fmt"
	"testing"

	"github.com/OpenMellon/typesetting/di"
	"github.com/OpenMellon/typesetting/language"
	tu "github.com/OpenMellon/typesetting/testutils"
	"golang.org/x/image/math/fixed"
)

func BenchmarkWrapping(b *testing.B) {
	for _, langInfo := range benchLangs {
		for _, size := range benchSizes {
			for _, parts := range size.parts {
				b.Run(fmt.Sprintf("%drunes-%s-%dparts", size.runes, langInfo.name, parts), func(b *testing.B) {
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
					var l LineWrapper
					iter := NewSliceIterator(outs)
					b.ResetTimer()
					lines := make([]Line, 1)
					for i := 0; i < b.N; i++ {
						lines, _ = l.WrapParagraph(WrapConfig{}, 100, langInfo.text[:size.runes], iter)
						iter.(*shapedRunSlice).Reset(outs)
					}
					_ = lines
				})
			}
		}
	}
}

// BenchmarkWrappingHappyPath measures the performance when it's obvious that
// the shaped text will fit within the available space without doing any extra
// work.
func BenchmarkWrappingHappyPath(b *testing.B) {
	textInput := []rune("happy path")
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
	iter := NewSliceIterator(out)
	b.ResetTimer()
	var outs []Line
	for i := 0; i < b.N; i++ {
		outs, _ = l.WrapParagraph(WrapConfig{BreakPolicy: Never}, 100, textInput, iter)
		iter.(*shapedRunSlice).Reset(out)
	}
	_ = outs
}

const benchParagraphLatin = `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Porttitor eget dolor morbi non arcu risus quis. Nibh sit amet commodo nulla. Posuere ac ut consequat semper viverra nam libero justo. Risus in hendrerit gravida rutrum quisque. Natoque penatibus et magnis dis parturient montes nascetur. In metus vulputate eu scelerisque felis imperdiet proin fermentum. Mattis rhoncus urna neque viverra. Elit pellentesque habitant morbi tristique. Nisl nunc mi ipsum faucibus vitae aliquet nec. Sed augue lacus viverra vitae congue eu consequat. At quis risus sed vulputate odio ut. Sit amet volutpat consequat mauris nunc congue nisi. Dignissim cras tincidunt lobortis feugiat. Faucibus turpis in eu mi bibendum. Odio aenean sed adipiscing diam donec adipiscing tristique. Fermentum leo vel orci porta non pulvinar. Ut venenatis tellus in metus vulputate eu scelerisque felis imperdiet. Et netus et malesuada fames ac turpis. Venenatis urna cursus eget nunc scelerisque viverra mauris in. Risus ultricies tristique nulla aliquet enim tortor. Risus pretium quam vulputate dignissim suspendisse in. Interdum velit euismod in pellentesque massa placerat duis ultricies lacus. Proin gravida hendrerit lectus a. Auctor augue mauris augue neque gravida in fermentum et. Laoreet sit amet cursus sit amet dictum. In fermentum et sollicitudin ac orci phasellus egestas tellus rutrum. Tempus imperdiet nulla malesuada pellentesque elit eget gravida. Consequat id porta nibh venenatis cras sed. Vulputate ut pharetra sit amet aliquam. Congue mauris rhoncus aenean vel elit. Risus quis varius quam quisque id diam vel quam elementum. Pretium lectus quam id leo in vitae. Sed sed risus pretium quam vulputate dignissim suspendisse in est. Velit laoreet id donec ultrices. Nunc sed velit dignissim sodales ut. Nunc scelerisque viverra mauris in aliquam sem fringilla ut. Sed enim ut sem viverra aliquet eget sit. Convallis posuere morbi leo urna molestie at. Aliquam id diam maecenas ultricies mi eget mauris. Ipsum dolor sit amet consectetur adipiscing elit ut aliquam. Accumsan tortor posuere ac ut consequat semper. Viverra vitae congue eu consequat ac felis donec et odio. Scelerisque in dictum non consectetur a. Consequat nisl vel pretium lectus quam id leo in vitae. Morbi tristique senectus et netus et malesuada fames ac turpis. Ac orci phasellus egestas tellus. Tempus egestas sed sed risus. Ullamcorper morbi tincidunt ornare massa eget egestas purus. Nibh venenatis cras sed felis eget velit.`

const benchParagraphArabic = `و سأعرض مثال حي لهذا، من منا لم يتحمل جهد بدني شاق إلا من أجل الحصول على ميزة أو فائدة؟ ولكن من لديه الحق أن ينتقد شخص ما أراد أن يشعر بالسعادة التي لا تشوبها عواقب أليمة أو آخر أراد أن يتجنب الألم الذي ربما تنجم عنه بعض المتعة ؟ علي الجانب الآخر نشجب ونستنكر هؤلاء الرجال المفتونون بنشوة اللحظة الهائمون في رغباتهم فلا يدركون ما يعقبها من الألم والأسي المحتم، واللوم كذلك يشمل هؤلاء الذين أخفقوا في واجباتهم نتيجة لضعف إرادتهم فيتساوي مع هؤلاء الذين يتجنبون وينأون عن تحمل الكدح والألم . من المفترض أن نفرق بين هذه الحالات بكل سهولة ومرونة. في ذاك الوقت عندما تكون قدرتنا علي الاختيار غير مقيدة بشرط وعندما لا نجد ما يمنعنا أن نفعل الأفضل فها نحن نرحب بالسرور والسعادة ونتجنب كل ما يبعث إلينا الألم. في بعض الأحيان ونظراً للالتزامات التي يفرضها علينا الواجب والعمل سنتنازل غالباً ونرفض الشعور بالسرور ونقبل ما يجلبه إلينا الأسى. الإنسان الحكيم عليه أن يمسك زمام الأمور ويختار إما أن يرفض مصادر السعادة من أجل ما هو أكثر أهمية أو يتحمل الألم من أجل ألا يتحمل ما هو أسوأ. و سأعرض مثال حي لهذا، من منا لم يتحمل جهد بدني شاق إلا من أجل الحصول على ميزة أو فائدة؟ ولكن من لديه الحق أن ينتقد شخص ما أراد أن يشعر بالسعادة التي لا تشوبها عواقب أليمة أو آخر أراد أن يتجنب الألم الذي ربما تنجم عنه بعض المتعة ؟ علي الجانب الآخر نشجب ونستنكر هؤلاء الرجال المفتونون بنشوة اللحظة الهائمون في رغباتهم فلا يدركون ما يعقبها من الألم والأسي المحتم، واللوم كذلك يشمل هؤلاء الذين أخفقوا في واجباتهم نتيجة لضعف إرادتهم فيتساوي مع هؤلاء الذين يتجنبون وينأون عن تحمل الكدح والألم . من المفترض أن نفرق بين هذه الحالات بكل سهولة ومرونة. في ذاك الوقت عندما تكون قدرتنا علي الاختيار غير مقيدة بشرط وعندما لا نجد ما يمنعنا أن نفعل الأفضل فها نحن نرحب بالسرور والسعادة ونتجنب كل ما يبعث إلينا الألم. في بعض الأحيان ونظراً للالتزامات التي يفرضها علينا الواجب والعمل سنتنازل غالباً ونرفض الشعور بالسرور ونقبل ما يجلبه إلينا الأسى. الإنسان الحكيم عليه أن يمسك زمام الأمور ويختار إما أن يرفض مصادر السعادة من أجل ما هو أكثر أهمية أو يتحمل الألم من أجل ألا يتحمل ما هو أسوأ.`

var regressionRuns = []Output{
	{
		Advance: 5490,
		Size:    640,
		Glyphs: []Glyph{
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 0, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 1, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 2, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 3, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 4, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 5, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 6, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 7, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 8, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 9, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 10, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 11, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 12, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 13, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 14, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
		},
		LineBounds: Bounds{
			Ascent: 879, Descent: -472, Gap: 0,
		},
		GlyphBounds: Bounds{
			Ascent: 464, Descent: -6, Gap: 0,
		},
		Direction: 0x0,
		Runes: Range{
			Offset: 0, Count: 15,
		},
	},
	{
		Advance: 166,
		Size:    640,
		Glyphs: []Glyph{
			{
				Width: 0, Height: 0, XBearing: 0, YBearing: 0, XAdvance: 166, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 15, RuneCount: 1, GlyphCount: 1, GlyphID: 0x3, Mask: 0x80000000,
			},
		},
		LineBounds: Bounds{
			Ascent: 879, Descent: -472, Gap: 0,
		},
		GlyphBounds: Bounds{
			Ascent: 0, Descent: 0, Gap: 0,
		},
		Direction: 0x1,
		Runes: Range{
			Offset: 15, Count: 1,
		},
	},
	{
		Advance: 2928,
		Size:    640,
		Glyphs: []Glyph{
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 16, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 17, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 18, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 19, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 20, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 21, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 22, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 23, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
		},
		LineBounds: Bounds{
			Ascent: 879, Descent: -472, Gap: 0,
		},
		GlyphBounds: Bounds{
			Ascent: 464, Descent: -6, Gap: 0,
		},
		Direction: 0x0,
		Runes: Range{
			Offset: 16, Count: 8,
		},
	},
	{
		Advance: 166,
		Size:    640,
		Glyphs: []Glyph{
			{
				Width: 0, Height: 0, XBearing: 0, YBearing: 0, XAdvance: 166, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 24, RuneCount: 2, GlyphCount: 2, GlyphID: 0x3, Mask: 0x80000000,
			},
			{
				Width: 39, Height: -111, XBearing: 29, YBearing: 495, XAdvance: 0, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 24, RuneCount: 2, GlyphCount: 2, GlyphID: 0x57, Mask: 0x80000000,
			},
		},
		LineBounds: Bounds{
			Ascent: 879, Descent: -472, Gap: 0,
		},
		GlyphBounds: Bounds{
			Ascent: 495, Descent: 0, Gap: 0,
		},
		Direction: 0x1,
		Runes: Range{
			Offset: 24, Count: 2,
		},
	},
	{
		Advance: 1830,
		Size:    640,
		Glyphs: []Glyph{
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 26, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 27, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 28, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 29, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 30, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
		},
		LineBounds: Bounds{
			Ascent: 879, Descent: -472, Gap: 0,
		},
		GlyphBounds: Bounds{
			Ascent: 464, Descent: -6, Gap: 0,
		},
		Direction: 0x0,
		Runes: Range{
			Offset: 26, Count: 5,
		},
	},
}

func TestTrailingSpace(t *testing.T) {
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

	text := []rune(" The quick brown fox")

	face := loadOpentypeFont(t, "../font/testdata/UbuntuMono-R.ttf")
	run := (&HarfbuzzShaper{}).Shape(Input{
		Text:   text,
		Face:   face,
		Size:   72,
		RunEnd: len(text),
	})
	tu.Assert(t, run.Glyphs[0].XAdvance == fixed.I(1))

	lines, _ := (&LineWrapper{}).WrapParagraph(WrapConfig{BreakPolicy: Always}, 5, text, NewSliceIterator([]Output{run}))

	expected := []string{
		" The ",
		"quick ",
		"brown ",
		"fox",
	}
	tu.Assert(t, len(lines) == 4)
	for i, line := range lines {
		s := shapedText(text, line)
		tu.Assert(t, s == expected[i])
	}
}
