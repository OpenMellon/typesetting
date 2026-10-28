package harfbuzz

// ported from harfbuzz/src/hb-aat-layout.h  Copyright © 2018 Ebrahim Byagowi, Behdad Esfahbod

// The possible feature types defined for AAT shaping,
// from https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html
type aatLayoutFeatureType = uint16

const (
	// Initial, unset feature type
	// aatLayoutFeatureTypeInvalid = 0xFFFF
	// [All Typographic Features](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type0)
	// aatLayoutFeatureTypeAllTypographic = 0
	// [Ligatures](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type1)
	aatLayoutFeatureTypeLigatures = 1
	// [Cursive Connection](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type2)
	// aatLayoutFeatureTypeCurisveConnection = 2
	// [Letter Case](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type3)
	aatLayoutFeatureTypeLetterCase = 3
	// [Vertical Substitution](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type4)
	aatLayoutFeatureTypeVerticalSubstitution = 4
	// [Linguistic Rearrangement](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type5)
	// aatLayoutFeatureTypeLinguisticRearrangement = 5
	// [Number Spacing](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type6)
	aatLayoutFeatureTypeNumberSpacing = 6
	// [Smart Swash](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type8)
	// aatLayoutFeatureTypeSmartSwashType = 8
	// [Diacritics](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type9)
	// aatLayoutFeatureTypeDiacriticsType = 9
	// [Vertical Position](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type10)
	aatLayoutFeatureTypeVerticalPosition = 10
	// [Fractions](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type11)
	aatLayoutFeatureTypeFractions = 11
	// [Overlapping Characters](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type13)
	// aatLayoutFeatureTypeOverlappingCharactersType = 13
	// [Typographic Extras](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type14)
	aatLayoutFeatureTypeTypographicExtras = 14
	// [Mathematical Extras](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type15)
	aatLayoutFeatureTypeMathematicalExtras = 15
	// [Ornament Sets](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type16)
	// aatLayoutFeatureTypeOrnamentSetsType = 16
	// [Character Alternatives](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type17)
	aatLayoutFeatureTypeCharacterAlternatives = 17
	// [Style Options](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type19)
	aatLayoutFeatureTypeStyleOptions = 19
	// [Character Shape](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type20)
	aatLayoutFeatureTypeCharacterShape = 20
	// [Number Case](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type21)
	aatLayoutFeatureTypeNumberCase = 21
	// [Text Spacing](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type22)
	aatLayoutFeatureTypeTextSpacing = 22
	// [Transliteration](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type23)
	aatLayoutFeatureTypeTransliteration = 23

	// [Ruby Kana](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type28)
	aatLayoutFeatureTypeRubyKana = 28

	// [Italic CJK Roman](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type32)
	aatLayoutFeatureTypeItalicCjkRoman = 32
	// [Case Sensitive Layout](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type33)
	aatLayoutFeatureTypeCaseSensitiveLayout = 33
	// [Alternate Kana](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type34)
	aatLayoutFeatureTypeAlternateKana = 34
	// [Stylistic Alternatives](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type35)
	aatLayoutFeatureTypeStylisticAlternatives = 35
	// [Contextual Alternatives](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type36)
	aatLayoutFeatureTypeContextualAlternatives = 36
	// [Lower Case](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type37)
	aatLayoutFeatureTypeLowerCase = 37
	// [Upper Case](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type38)
	aatLayoutFeatureTypeUpperCase = 38
	// [Language Tag](https://developer.apple.com/fonts/TrueType-Reference-Manual/RM09/AppendixF.html#Type39)
	aatLayoutFeatureTypeLanguageTagType = 39
)

// The selectors defined for specifying AAT feature settings.
type aatLayoutFeatureSelector = uint16
