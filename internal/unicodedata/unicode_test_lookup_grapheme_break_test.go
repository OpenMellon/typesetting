package unicodedata

import (
	"testing"
	"unicode"

	"github.com/OpenMellon/typesetting/language"
	tu "github.com/OpenMellon/typesetting/testutils"
)

func TestLookupGraphemeBreak(t *testing.T) {
	for _, tt := range graphemeBreakTests {
		tu.Assert(t, LookupGraphemeBreak(tt.args) == tt.want)
	}
}

var wordBreakTests = []struct {
	args rune
	want WordBreak
}{
	// these runes have changed from Unicode v15.0 to 15.1
	{0x6DD, WB_Numeric},
	{0x661, WB_Numeric},

	{'\u000D', WB_NewlineCRLF},            // CARRIAGE RETURN (CR)
	{'\u000A', WB_NewlineCRLF},            // LINE FEED (LF)
	{'\u000B', WB_NewlineCRLF},            // LINE TABULATION
	{'\u000C', WB_NewlineCRLF},            // FORM FEED (FF)
	{'\u0085', WB_NewlineCRLF},            // NEXT LINE (NEL)
	{'\u2028', WB_NewlineCRLF},            // LINE SEPARATOR
	{'\u2029', WB_NewlineCRLF},            // PARAGRAPH SEPARATOR
	{'\U0001F1E6', WB_Regional_Indicator}, // REGIONAL INDICATOR SYMBOL LETTER A
	{'\U0001F1FF', WB_Regional_Indicator}, // REGIONAL INDICATOR SYMBOL LETTER Z
	{'\u3031', WB_Katakana},               // ( 〱 ) VERTICAL KANA REPEAT MARK
	{'\u3032', WB_Katakana},               // ( 〲 ) VERTICAL KANA REPEAT WITH VOICED SOUND MARK
	{'\u3033', WB_Katakana},               // ( 〳 ) VERTICAL KANA REPEAT MARK UPPER HALF
	{'\u3034', WB_Katakana},               // ( 〴 ) VERTICAL KANA REPEAT WITH VOICED SOUND MARK UPPER HALF
	{'\u3035', WB_Katakana},               // ( 〵 ) VERTICAL KANA REPEAT MARK LOWER HALF
	{'\u309B', WB_Katakana},               // ( ゛ ) KATAKANA-HIRAGANA VOICED SOUND MARK
	{'\u309C', WB_Katakana},               // ( ゜ ) KATAKANA-HIRAGANA SEMI-VOICED SOUND MARK
	{'\u30A0', WB_Katakana},               // ( ゠ ) KATAKANA-HIRAGANA DOUBLE HYPHEN
	{'\u30FC', WB_Katakana},               // ( ー ) KATAKANA-HIRAGANA PROLONGED SOUND MARK
	{'\uFF70', WB_Katakana},               // ( ｰ ) HALFWIDTH KATAKANA-HIRAGANA PROLONGED SOUND MARK
	{'\u00B8', WB_ALetter},                // ( ¸ ) CEDILLA
	{'\u02C2', WB_ALetter},                // ( ˂ ) MODIFIER LETTER LEFT ARROWHEAD
	{'\u02C5', WB_ALetter},                // ( ˅ ) MODIFIER LETTER DOWN ARROWHEAD
	{'\u02D2', WB_ALetter},                // ( ˒ ) MODIFIER LETTER CENTRED RIGHT HALF RING
	{'\u02D7', WB_ALetter},                // ( ˗ ) MODIFIER LETTER MINUS SIGN
	{'\u02DE', WB_ALetter},                // ( ˞ ) MODIFIER LETTER RHOTIC HOOK
	{'\u02DF', WB_ALetter},                // ( ˟ ) MODIFIER LETTER CROSS ACCENT
	{'\u02E5', WB_ALetter},                // ( ˥ ) MODIFIER LETTER EXTRA-HIGH TONE BAR
	{'\u02EB', WB_ALetter},                // ( ˫ ) MODIFIER LETTER YANG DEPARTING TONE MARK
	{'\u02ED', WB_ALetter},                // ( ˭ ) MODIFIER LETTER UNASPIRATED
	{'\u02EF', WB_ALetter},                // ( ˯ ) MODIFIER LETTER LOW DOWN ARROWHEAD
	{'\u02FF', WB_ALetter},                // ( ˿ ) MODIFIER LETTER LOW LEFT ARROW
	{'\u055A', WB_ALetter},                // ( ՚ ) ARMENIAN APOSTROPHE
	{'\u055B', WB_ALetter},                // ( ՛ ) ARMENIAN EMPHASIS MARK
	{'\u055C', WB_ALetter},                // ( ՜ ) ARMENIAN EXCLAMATION MARK
	{'\u055E', WB_ALetter},                // ( ՞ ) ARMENIAN QUESTION MARK
	{'\u058A', WB_ALetter},                // ( ֊ ) ARMENIAN HYPHEN
	{'\u05F3', WB_ALetter},                // ( ׳ ) HEBREW PUNCTUATION GERESH
	{'\u070F', WB_ALetter},                // ( ܏ ) SYRIAC ABBREVIATION MARK
	{'\uA708', WB_ALetter},                // ( ꜈ ) MODIFIER LETTER EXTRA-HIGH DOTTED TONE BAR
	{'\uA716', WB_ALetter},                // ( ꜖ ) MODIFIER LETTER EXTRA-LOW LEFT-STEM TONE BAR
	{'\uA720', WB_ALetter},                // (꜠ ) MODIFIER LETTER STRESS AND HIGH TONE
	{'\uA721', WB_ALetter},                // (꜡ ) MODIFIER LETTER STRESS AND LOW TONE
	{'\uA789', WB_ALetter},                // (꞉ ) MODIFIER LETTER COLON
	{'\uA78A', WB_ALetter},                // ( ꞊ ) MODIFIER LETTER SHORT EQUALS SIGN
	{'\uAB5B', WB_ALetter},                // ( ꭛ ) MODIFIER BREVE WITH INVERTED BREVE
	{'\u0027', WB_Single_Quote},           // ( ' ) APOSTROPHE
	{'\u0022', WB_Double_Quote},           // ( " ) QUOTATION MARK
	{'\u002E', WB_MidNumLet},              // ( . ) FULL STOP
	{'\u2018', WB_MidNumLet},              // ( ‘ ) LEFT SINGLE QUOTATION MARK
	{'\u2019', WB_MidNumLet},              // ( ’ ) RIGHT SINGLE QUOTATION MARK
	{'\u2024', WB_MidNumLet},              // ( ․ ) ONE DOT LEADER
	{'\uFE52', WB_MidNumLet},              // ( ﹒ ) SMALL FULL STOP
	{'\uFF07', WB_MidNumLet},              // ( ＇ ) FULLWIDTH APOSTROPHE
	{'\uFF0E', WB_MidNumLet},              // ( ． ) FULLWIDTH FULL STOP
	{'\u003A', WB_MidLetter},              //  ( : ) COLON (used in Swedish)
	{'\u00B7', WB_MidLetter},              //  ( · ) MIDDLE DOT
	{'\u0387', WB_MidLetter},              //  ( · ) GREEK ANO TELEIA
	{'\u055F', WB_MidLetter},              //  ( ՟ ) ARMENIAN ABBREVIATION MARK
	{'\u05F4', WB_MidLetter},              //  ( ״ ) HEBREW PUNCTUATION GERSHAYIM
	{'\u2027', WB_MidLetter},              //  ( ‧ ) HYPHENATION POINT
	{'\uFE13', WB_MidLetter},              //  ( ︓ ) PRESENTATION FORM FOR VERTICAL COLON
	{'\uFE55', WB_MidLetter},              //  ( ﹕ ) SMALL COLON
	{'\uFF1A', WB_MidLetter},              //  ( ： ) FULLWIDTH COLON
	{'\u066C', WB_MidNum},                 // ( ٬ ) ARABIC THOUSANDS SEPARATOR
	{'\uFE50', WB_MidNum},                 // ( ﹐ ) SMALL COMMA
	{'\uFE54', WB_MidNum},                 // ( ﹔ ) SMALL SEMICOLON
	{'\uFF0C', WB_MidNum},                 // ( ， ) FULLWIDTH COMMA
	{'\uFF1B', WB_MidNum},                 // ( ； ) FULLWIDTH SEMICOLON
	{'\u202F', WB_ExtendNumLet},           // NARROW NO-BREAK SPACE (NNBSP)
}

func TestLookupWordBreak(t *testing.T) {
	for _, tt := range wordBreakTests {
		tu.Assert(t, LookupWordBreak(tt.args) == tt.want)
	}
}

var isWordTests = []struct {
	args rune
	want bool
}{
	{',', false},
	{'.', false},
	{0xffd2, true},
	{0xffd7, true},
	{0xffd8, false},
	{0xffda, true},
	{0xffdc, true},
	{0xffdd, false},
	{0x10000, true},
	{0x1000b, true},
	{0x1000c, false},
	{0x1000d, true},
	{0x10026, true},
}

func TestIsWord(t *testing.T) {
	for _, tt := range isWordTests {
		tu.Assert(t, IsWord(tt.args) == tt.want)
	}
}

var indicConjunctBreakTests = []struct {
	r     rune
	class IndicConjunctBreak
}{
	{-1, 0},
	{1, 0},
	{'a', 0},
	{'9', 0},
	{0x100, 0},
	{0x1a60, ICBLinker},
	{0x1bab, ICBLinker},
	{0xaaf6, ICBLinker},
	{0x10a3f, ICBLinker},
	{0xaae1, ICBConsonant},
	{0xabc0, ICBConsonant},
	{0x10a00, ICBConsonant},
	{0x10a11, ICBConsonant},
	{0x10a15, ICBConsonant},
	{0x10a19, ICBConsonant},
	{0xfe01, ICBExtend},
	{0xfe20, ICBExtend},
	{0xff9e, ICBExtend},
	{0x101fd, ICBExtend},
	{0x10376, ICBExtend},
	{0x10a01, ICBExtend},
	{0x10a05, ICBExtend},
	{0x10a0c, ICBExtend},
	{0x10a38, ICBExtend},

	{0x1a50, 0b10},
	{0x1a51, 0b10},
	{0x1a52, 0b10},
	{0x1a53, 0b10},
	{0x1a54, 0b10},
	{0x1a55, 0b0},
	{0x1a56, 0b100},
	{0x1a57, 0b0},
	{0x1a58, 0b100},
	{0x1a59, 0b100},
	{0x1a5a, 0b100},
	{0x1a5b, 0b100},
	{0x1a5c, 0b100},
	{0x1a5d, 0b100},
	{0x1a5e, 0b100},
	{0x1a5f, 0b0},
	{0x1a60, 0b1},
	{0x1a61, 0b0},
	{0x1a62, 0b100},
	{0x1a63, 0b0},
	{0x1a64, 0b0},
	{0x1a65, 0b100},
	{0x1a66, 0b100},
	{0x1a67, 0b100},
	{0x1a68, 0b100},
	{0x1a69, 0b100},
	{0x1a6a, 0b100},
	{0x1a6b, 0b100},
	{0x1a6c, 0b100},
	{0x1a6d, 0b0},
	{0x1a6e, 0b0},
	{0x1a6f, 0b0},
	{0x1a70, 0b0},
	{0x1a71, 0b0},
	{0x1a72, 0b0},
	{0x1a73, 0b100},
	{0x1a74, 0b100},
	{0x1a75, 0b100},
	{0x1a76, 0b100},
	{0x1a77, 0b100},
	{0x1a78, 0b100},
	{0x1a79, 0b100},
	{0x1a7a, 0b100},
	{0x1a7b, 0b100},
	{0x1a7c, 0b100},
	{0x1a7d, 0b0},
	{0x1a7e, 0b0},
	{0x1a7f, 0b100},
	{0x1a80, 0b0},
	{0x1a81, 0b0},
	{0x1a82, 0b0},
	{0x1a83, 0b0},
	{0x1a84, 0b0},
	{0x1a85, 0b0},
	{0x1a86, 0b0},
	{0x1a87, 0b0},
	{0x1a88, 0b0},
	{0x1a89, 0b0},
	{0x1a8a, 0b0},
	{0x1a8b, 0b0},
	{0x1a8c, 0b0},
	{0x1a8d, 0b0},
	{0x1a8e, 0b0},
	{0x1a8f, 0b0},
	{0x1a90, 0b0},
	{0x1a91, 0b0},
	{0x1a92, 0b0},
	{0x1a93, 0b0},
	{0x1a94, 0b0},
	{0x1a95, 0b0},
	{0x1a96, 0b0},
	{0x1a97, 0b0},
	{0x1a98, 0b0},
	{0x1a99, 0b0},
	{0x1a9a, 0b0},
	{0x1a9b, 0b0},
	{0x1a9c, 0b0},
	{0x1a9d, 0b0},
	{0x1a9e, 0b0},
	{0x1a9f, 0b0},
	{0x1aa0, 0b0},
	{0x1aa1, 0b0},
	{0x1aa2, 0b0},
	{0x1aa3, 0b0},
	{0x1aa4, 0b0},
	{0x1aa5, 0b0},
	{0x1aa6, 0b0},
	{0x1aa7, 0b0},
	{0x1aa8, 0b0},
	{0x1aa9, 0b0},
	{0x1aaa, 0b0},
	{0x1aab, 0b0},
	{0x1aac, 0b0},
	{0x1aad, 0b0},
	{0x1aae, 0b0},
	{0x1aaf, 0b0},
	{0x1ab0, 0b100},
	{0x1ab1, 0b100},
	{0x1ab2, 0b100},
	{0x1ab3, 0b100},
	{0x1ab4, 0b100},
	{0x1ab5, 0b100},
	{0x1ab6, 0b100},
	{0x1ab7, 0b100},
	{0x1ab8, 0b100},
	{0x1ab9, 0b100},
	{0x1aba, 0b100},
	{0x1abb, 0b100},
	{0x1abc, 0b100},
	{0x1abd, 0b100},
	{0x1abe, 0b100},
	{0x1abf, 0b100},
	{0x1ac0, 0b100},
	{0x1ac1, 0b100},
	{0x1ac2, 0b100},
	{0x1ac3, 0b100},
	{0x1ac4, 0b100},
	{0x1ac5, 0b100},
	{0x1ac6, 0b100},
	{0x1ac7, 0b100},
	{0x1ac8, 0b100},
	{0x1ac9, 0b100},
	{0x1aca, 0b100},
	{0x1acb, 0b100},
	{0x1acc, 0b100},
	{0x1acd, 0b100},
	{0x1ace, 0b100},
	{0x1acf, 0b100},
	{0x1ad0, 0b100},
	{0x1ad1, 0b100},
	{0x1ad2, 0b100},
	{0x1ad3, 0b100},
	{0x1ad4, 0b100},
	{0x1ad5, 0b100},
	{0x1ad6, 0b100},
	{0x1ad7, 0b100},
	{0x1ad8, 0b100},
	{0x1ad9, 0b100},
	{0x1ada, 0b100},
	{0x1adb, 0b100},
	{0x1adc, 0b100},
	{0x1add, 0b100},
	{0x1ade, 0b0},
	{0x1adf, 0b0},
	{0x1ae0, 0b100},
	{0x1ae1, 0b100},
	{0x1ae2, 0b100},
	{0x1ae3, 0b100},
	{0x1ae4, 0b100},
	{0x1ae5, 0b100},
	{0x1ae6, 0b100},
	{0x1ae7, 0b100},
	{0x1ae8, 0b100},
	{0x1ae9, 0b100},
	{0x1aea, 0b100},
	{0x1aeb, 0b100},
	{0x1aec, 0b0},
	{0x1aed, 0b0},
	{0x1aee, 0b0},
	{0x1aef, 0b0},
}

func TestIndicConjunctBreak(t *testing.T) {
	for _, test := range indicConjunctBreakTests {
		tu.Assert(t, LookupIndicConjunctBreak(test.r) == test.class)
	}
}

var mirTests = []struct {
	args rune
	want rune
}{
	{'a', 'a'},
	{'\u2231', '\u2231'},
	{'\u0028', '\u0029'},
	{'\u0029', '\u0028'},
	{'\u22E0', '\u22E1'},
	/* Some characters that do NOT mirror */
	{0x0020, 0x0020},
	{0x0041, 0x0041},
	{0x00F0, 0x00F0},
	{0x27CC, 0x27CC},
	{0xE01EF, 0xE01EF},
	{0x1D7C3, 0x1D7C3},
	{0x100000, 0x100000},
	/* Some characters that do mirror */
	{0x0029, 0x0028},
	{0x0028, 0x0029},
	{0x003E, 0x003C},
	{0x003C, 0x003E},
	{0x005D, 0x005B},
	{0x005B, 0x005D},
	{0x007D, 0x007B},
	{0x007B, 0x007D},
	{0x00BB, 0x00AB},
	{0x00AB, 0x00BB},
	{0x226B, 0x226A},
	{0x226A, 0x226B},
	{0x22F1, 0x22F0},
	{0x22F0, 0x22F1},
	{0xFF60, 0xFF5F},
	{0xFF5F, 0xFF60},
	{0xFF63, 0xFF62},
	{0xFF62, 0xFF63},
	/* Unicode-6.1 character additions */
	{0x27CB, 0x27CD},
	/* Unicode-11.0 character additions */
	{0x2BFE, 0x221F},
	{0x111111, 0x111111},
}

func TestLookupMirrorChar(t *testing.T) {
	for _, tt := range mirTests {
		if got := LookupMirrorChar(tt.args); got != tt.want {
			t.Errorf("LookupMirrorChar() got = %v, want %v", got, tt.want)
		}
	}
}

func TestLookupVerticalOrientation(t *testing.T) {
	tests := []struct {
		s              language.Script
		r              rune
		wantIsSideways bool
	}{
		{language.Cyrillic, '\u0400', true},
		{language.Latin, 'A', true},
		{language.Latin, '\uFF21', false},
		{language.Katakana, 'も', false},
		{language.Katakana, '\uFF89', true},
		{language.Hangul, '\uFFAB', true},
	}
	for _, tt := range tests {
		if gotIsSideways := LookupVerticalOrientation(tt.s).Orientation(tt.r); gotIsSideways != tt.wantIsSideways {
			t.Errorf("LookupVerticalOrientation(%s) = %v, want %v", string(tt.r), gotIsSideways, tt.wantIsSideways)
		}
	}
}

func BenchmarkLookups(b *testing.B) {
	b.Run("GeneralCategory unicode.RangeTable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range generalCategoryTests {
				_ = lookupGeneralCategory(test.args)
			}
		}
	})
	b.Run("GeneralCategory packtable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range generalCategoryTests {
				_ = LookupGeneralCategory(test.args)
			}
		}
	})

	b.Run("Combining class unicode.RangeTable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range generalCategoryTests {
				_ = lookupCombiningClass(test.args)
			}
		}
	})
	b.Run("Combining class packtable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range generalCategoryTests {
				_ = LookupCombiningClass(test.args)
			}
		}
	})

	b.Run("Mirroring unicode.RangeTable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range generalCategoryTests {
				_ = lookupMirrorChar(test.args)
			}
		}
	})
	b.Run("Mirroring packtable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range generalCategoryTests {
				_ = LookupMirrorChar(test.args)
			}
		}
	})

	b.Run("Decompose map", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range decomposeTests {
				_, _, _ = decompose(test.ab)
			}
		}
	})
	b.Run("Decompose packtable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range decomposeTests {
				_, _, _ = Decompose(test.ab)
			}
		}
	})
	b.Run("Compose map", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range composeTests {
				_, _ = compose_(test.a, test.b)
			}
		}
	})
	b.Run("Compose packtable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range composeTests {
				_, _ = Compose(test.a, test.b)
			}
		}
	})

	b.Run("IsExtendedPictographic unicode.RangeTable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range extendedPictoTests {
				_ = unicode.Is(Extended_Pictographic, test.r)
			}
		}
	})
	b.Run("IsExtendedPictographic packtab", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range extendedPictoTests {
				_ = IsExtendedPictographic(test.r)
			}
		}
	})

	b.Run("IsLargeEastAsian unicode.RangeTable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range eastAsianWidthTests {
				_ = unicode.Is(LargeEastAsian, test.r)
			}
		}
	})
	b.Run("IsLargeEastAsian packtab", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range eastAsianWidthTests {
				_ = IsLargeEastAsian(test.r)
			}
		}
	})

	b.Run("LineBreak unicode.RangeTable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range lineBreakTests {
				_ = lookupLineBreak(test.args)
			}
		}
	})
	b.Run("LineBreak packtab", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range lineBreakTests {
				_ = LookupLineBreak(test.args)
			}
		}
	})

	b.Run("GraphemeBreak unicode.RangeTable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range indicConjunctBreakTests {
				_ = lookupGraphemeBreak(test.r)
			}
		}
	})
	b.Run("GraphemeBreak packtab", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range indicConjunctBreakTests {
				_ = LookupIndicConjunctBreak(test.r)
			}
		}
	})

	b.Run("WordBreak unicode.RangeTable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range wordBreakTests {
				_ = lookupWordBreak(test.args)
			}
		}
	})
	b.Run("WordBreak packtab", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range wordBreakTests {
				_ = LookupIndicConjunctBreak(test.args)
			}
		}
	})

	b.Run("IsWord unicode.RangeTable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range isWordTests {
				_ = unicode.Is(Word, test.args)
			}
		}
	})
	b.Run("IsWord packtab", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range isWordTests {
				_ = LookupIndicConjunctBreak(test.args)
			}
		}
	})

	b.Run("IndicConjunctBreak unicode.RangeTable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range indicConjunctBreakTests {
				_ = lookupIndicConjunctBreak(test.r)
			}
		}
	})
	b.Run("IndicConjunctBreak packtab", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range indicConjunctBreakTests {
				_ = LookupIndicConjunctBreak(test.r)
			}
		}
	})
	b.Run("Combining class packtable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range generalCategoryTests {
				_ = LookupCombiningClass(test.args)
			}
		}
	})
	b.Run("Combining class unicode.RangeTable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range generalCategoryTests {
				_ = lookupCombiningClass(test.args)
			}
		}
	})
}

// legacy code, used (for now) as comparison

var allCategories = [...]*unicode.RangeTable{
	unicode.Cc,
	unicode.Cf,
	unicode.Co,
	unicode.Cs,
	unicode.Ll,
	unicode.Lm,
	unicode.Lo,
	unicode.Lt,
	unicode.Lu,
	unicode.Mc,
	unicode.Me,
	unicode.Mn,
	unicode.Nd,
	unicode.Nl,
	unicode.No,
	unicode.Pc,
	unicode.Pd,
	unicode.Pe,
	unicode.Pf,
	unicode.Pi,
	unicode.Po,
	unicode.Ps,
	unicode.Sc,
	unicode.Sk,
	unicode.Sm,
	unicode.So,
	unicode.Zl,
	unicode.Zp,
	unicode.Zs,
}

// simple implementation using standard library,
// here for benchmark reference
func lookupGeneralCategory(r rune) *unicode.RangeTable {
	for _, table := range allCategories {
		if unicode.Is(table, r) {
			return table
		}
	}
	return nil
}
