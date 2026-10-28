package harfbuzz

const (
	/* Selectors for #aatLayoutFeatureTypeLigatures */
	// for #aatLayoutFeatureTypeLigatures
	aatLayoutFeatureSelectorRequiredLigaturesOn = 0
	// for #aatLayoutFeatureTypeLigatures
	aatLayoutFeatureSelectorRequiredLigaturesOff = 1
	// for #aatLayoutFeatureTypeLigatures
	aatLayoutFeatureSelectorCommonLigaturesOn = 2
	// for #aatLayoutFeatureTypeLigatures
	aatLayoutFeatureSelectorCommonLigaturesOff = 3
	// for #aatLayoutFeatureTypeLigatures
	aatLayoutFeatureSelectorRareLigaturesOn = 4
	// for #aatLayoutFeatureTypeLigatures
	aatLayoutFeatureSelectorRareLigaturesOff = 5

	// for #aatLayoutFeatureTypeLigatures
	aatLayoutFeatureSelectorContextualLigaturesOn = 18
	// for #aatLayoutFeatureTypeLigatures
	aatLayoutFeatureSelectorContextualLigaturesOff = 19
	// for #aatLayoutFeatureTypeLigatures
	aatLayoutFeatureSelectorHistoricalLigaturesOn = 20
	// for #aatLayoutFeatureTypeLigatures
	aatLayoutFeatureSelectorHistoricalLigaturesOff = 21

	/* Selectors for #aatLayoutFeatureTypeLetterCase */

	// Deprecated
	aatLayoutFeatureSelectorSmallCaps = 3 /* deprecated */

	/* Selectors for #aatLayoutFeatureTypeVerticalSubstitution */
	// for #aatLayoutFeatureTypeVerticalSubstitution
	aatLayoutFeatureSelectorSubstituteVerticalFormsOn = 0
	// for #aatLayoutFeatureTypeVerticalSubstitution
	aatLayoutFeatureSelectorSubstituteVerticalFormsOff = 1

	/* Selectors for #aatLayoutFeatureTypeNumberSpacing */
	// for #aatLayoutFeatureTypeNumberSpacing
	aatLayoutFeatureSelectorMonospacedNumbers = 0
	// for #aatLayoutFeatureTypeNumberSpacing
	aatLayoutFeatureSelectorProportionalNumbers = 1

	/* Selectors for #aatLayoutFeatureTypeVerticalPosition */
	// for #aatLayoutFeatureTypeVerticalPosition
	aatLayoutFeatureSelectorNormalPosition = 0
	// for #aatLayoutFeatureTypeVerticalPosition
	aatLayoutFeatureSelectorSuperiors = 1
	// for #aatLayoutFeatureTypeVerticalPosition
	aatLayoutFeatureSelectorInferiors = 2
	// for #aatLayoutFeatureTypeVerticalPosition
	aatLayoutFeatureSelectorOrdinals = 3
	// for #aatLayoutFeatureTypeVerticalPosition
	aatLayoutFeatureSelectorScientificInferiors = 4

	/* Selectors for #aatLayoutFeatureTypeFractions */
	// for #aatLayoutFeatureTypeFractions
	aatLayoutFeatureSelectorNoFractions = 0
	// for #aatLayoutFeatureTypeFractions
	aatLayoutFeatureSelectorVerticalFractions = 1
	// for #aatLayoutFeatureTypeFractions
	aatLayoutFeatureSelectorDiagonalFractions = 2

	// for #aatLayoutFeatureTypeTypographicExtras
	aatLayoutFeatureSelectorSlashedZeroOn = 4
	// for #aatLayoutFeatureTypeTypographicExtras
	aatLayoutFeatureSelectorSlashedZeroOff = 5

	/* Selectors for #aatLayoutFeatureTypeMathematicalExtras */
	// for #aatLayoutFeatureTypeMathematicalExtras
	aatLayoutFeatureSelectorMathematicalGreekOn = 10
	// for #aatLayoutFeatureTypeMathematicalExtras
	aatLayoutFeatureSelectorMathematicalGreekOff = 11

	/* Selectors for #aatLayoutFeatureTypeStyleOptions */
	// for #aatLayoutFeatureTypeStyleOptions
	aatLayoutFeatureSelectorNoStyleOptions = 0

	// for #aatLayoutFeatureTypeStyleOptions
	aatLayoutFeatureSelectorTitlingCaps = 4

	/* Selectors for #aatLayoutFeatureTypeCharacterShape */
	// for #aatLayoutFeatureTypeCharacterShape
	aatLayoutFeatureSelectorTraditionalCharacters = 0
	// for #aatLayoutFeatureTypeCharacterShape
	aatLayoutFeatureSelectorSimplifiedCharacters = 1
	// for #aatLayoutFeatureTypeCharacterShape
	aatLayoutFeatureSelectorJis1978Characters = 2
	// for #aatLayoutFeatureTypeCharacterShape
	aatLayoutFeatureSelectorJis1983Characters = 3
	// for #aatLayoutFeatureTypeCharacterShape
	aatLayoutFeatureSelectorJis1990Characters = 4

	// for #aatLayoutFeatureTypeCharacterShape
	aatLayoutFeatureSelectorExpertCharacters = 10
	// for #aatLayoutFeatureTypeCharacterShape
	aatLayoutFeatureSelectorJis2004Characters = 11
	// for #aatLayoutFeatureTypeCharacterShape
	aatLayoutFeatureSelectorHojoCharacters = 12
	// for #aatLayoutFeatureTypeCharacterShape
	aatLayoutFeatureSelectorNlccharacters = 13
	// for #aatLayoutFeatureTypeCharacterShape
	aatLayoutFeatureSelectorTraditionalNamesCharacters = 14

	/* Selectors for #aatLayoutFeatureTypeNumberCase */
	// for #aatLayoutFeatureTypeNumberCase
	aatLayoutFeatureSelectorLowerCaseNumbers = 0
	// for #aatLayoutFeatureTypeNumberCase
	aatLayoutFeatureSelectorUpperCaseNumbers = 1

	/* Selectors for #aatLayoutFeatureTypeTextSpacing */
	// for #aatLayoutFeatureTypeTextSpacing
	aatLayoutFeatureSelectorProportionalText = 0
	// for #aatLayoutFeatureTypeTextSpacing
	aatLayoutFeatureSelectorMonospacedText = 1
	// for #aatLayoutFeatureTypeTextSpacing
	aatLayoutFeatureSelectorHalfWidthText = 2
	// for #aatLayoutFeatureTypeTextSpacing
	aatLayoutFeatureSelectorThirdWidthText = 3
	// for #aatLayoutFeatureTypeTextSpacing
	aatLayoutFeatureSelectorQuarterWidthText = 4
	// for #aatLayoutFeatureTypeTextSpacing
	aatLayoutFeatureSelectorAltProportionalText = 5
	// for #aatLayoutFeatureTypeTextSpacing
	aatLayoutFeatureSelectorAltHalfWidthText = 6

	/* Selectors for #aatLayoutFeatureTypeTransliteration */
	// for #aatLayoutFeatureTypeTransliteration
	aatLayoutFeatureSelectorNoTransliteration = 0
	// for #aatLayoutFeatureTypeTransliteration
	aatLayoutFeatureSelectorHanjaToHangul = 1

	/* Selectors for #aatLayoutFeatureTypeRubyKana */
	// for #aatLayoutFeatureTypeRubyKana
	aatLayoutFeatureSelectorRubyKanaOn = 2
	// for #aatLayoutFeatureTypeRubyKana
	aatLayoutFeatureSelectorRubyKanaOff = 3

	/* Selectors for #aatLayoutFeatureTypeItalicCjkRoman */
	// for #aatLayoutFeatureTypeItalicCjkRoman
	aatLayoutFeatureSelectorCjkItalicRomanOn = 2
	// for #aatLayoutFeatureTypeItalicCjkRoman
	aatLayoutFeatureSelectorCjkItalicRomanOff = 3

	/* Selectors for #aatLayoutFeatureTypeCaseSensitiveLayout */
	// for #aatLayoutFeatureTypeCaseSensitiveLayout
	aatLayoutFeatureSelectorCaseSensitiveLayoutOn = 0
	// for #aatLayoutFeatureTypeCaseSensitiveLayout
	aatLayoutFeatureSelectorCaseSensitiveLayoutOff = 1
	// for #aatLayoutFeatureTypeCaseSensitiveLayout
	aatLayoutFeatureSelectorCaseSensitiveSpacingOn = 2
	// for #aatLayoutFeatureTypeCaseSensitiveLayout
	aatLayoutFeatureSelectorCaseSensitiveSpacingOff = 3

	/* Selectors for #aatLayoutFeatureTypeAlternateKana */
	// for #aatLayoutFeatureTypeAlternateKana
	aatLayoutFeatureSelectorAlternateHorizKanaOn = 0
	// for #aatLayoutFeatureTypeAlternateKana
	aatLayoutFeatureSelectorAlternateHorizKanaOff = 1
	// for #aatLayoutFeatureTypeAlternateKana
	aatLayoutFeatureSelectorAlternateVertKanaOn = 2
	// for #aatLayoutFeatureTypeAlternateKana
	aatLayoutFeatureSelectorAlternateVertKanaOff = 3

	/* Selectors for #aatLayoutFeatureTypeStylisticAlternatives */

	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltOneOn = 2
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltOneOff = 3
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltTwoOn = 4
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltTwoOff = 5
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltThreeOn = 6
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltThreeOff = 7
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltFourOn = 8
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltFourOff = 9
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltFiveOn = 10
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltFiveOff = 11
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltSixOn = 12
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltSixOff = 13
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltSevenOn = 14
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltSevenOff = 15
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltEightOn = 16
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltEightOff = 17
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltNineOn = 18
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltNineOff = 19
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltTenOn = 20
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltTenOff = 21
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltElevenOn = 22
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltElevenOff = 23
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltTwelveOn = 24
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltTwelveOff = 25
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltThirteenOn = 26
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltThirteenOff = 27
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltFourteenOn = 28
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltFourteenOff = 29
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltFifteenOn = 30
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltFifteenOff = 31
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltSixteenOn = 32
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltSixteenOff = 33
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltSeventeenOn = 34
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltSeventeenOff = 35
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltEighteenOn = 36
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltEighteenOff = 37
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltNineteenOn = 38
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltNineteenOff = 39
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltTwentyOn = 40
	// for #aatLayoutFeatureTypeStylisticAlternatives
	aatLayoutFeatureSelectorStylisticAltTwentyOff = 41

	/* Selectors for #aatLayoutFeatureTypeContextualAlternatives */
	// for #aatLayoutFeatureTypeContextualAlternatives
	aatLayoutFeatureSelectorContextualAlternatesOn = 0
	// for #aatLayoutFeatureTypeContextualAlternatives
	aatLayoutFeatureSelectorContextualAlternatesOff = 1
	// for #aatLayoutFeatureTypeContextualAlternatives
	aatLayoutFeatureSelectorSwashAlternatesOn = 2
	// for #aatLayoutFeatureTypeContextualAlternatives
	aatLayoutFeatureSelectorSwashAlternatesOff = 3
	// for #aatLayoutFeatureTypeContextualAlternatives
	aatLayoutFeatureSelectorContextualSwashAlternatesOn = 4
	// for #aatLayoutFeatureTypeContextualAlternatives
	aatLayoutFeatureSelectorContextualSwashAlternatesOff = 5

	/* Selectors for #aatLayoutFeatureTypeLowerCase */
	// for #aatLayoutFeatureTypeLowerCase
	aatLayoutFeatureSelectorDefaultLowerCase = 0
	// for #aatLayoutFeatureTypeLowerCase
	aatLayoutFeatureSelectorLowerCaseSmallCaps = 1
	// for #aatLayoutFeatureTypeLowerCase
	aatLayoutFeatureSelectorLowerCasePetiteCaps = 2

	/* Selectors for #aatLayoutFeatureTypeUpperCase */
	// for #aatLayoutFeatureTypeUpperCase
	aatLayoutFeatureSelectorDefaultUpperCase = 0
	// for #aatLayoutFeatureTypeUpperCase
	aatLayoutFeatureSelectorUpperCaseSmallCaps = 1
	// for #aatLayoutFeatureTypeUpperCase
	aatLayoutFeatureSelectorUpperCasePetiteCaps = 2
)
