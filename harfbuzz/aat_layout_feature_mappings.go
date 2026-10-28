package harfbuzz

import (
	"github.com/OpenMellon/typesetting/font"
	ot "github.com/OpenMellon/typesetting/font/opentype"
	"github.com/OpenMellon/typesetting/font/opentype/tables"
)

// Mapping from OpenType feature tags to AAT feature names and selectors.
//
// Table data courtesy of Apple.  Converted from mnemonics to integers
// when moving to this file.
var featureMappings = [...]aatFeatureMapping{
	{ot.NewTag('a', 'f', 'r', 'c'), aatLayoutFeatureTypeFractions, aatLayoutFeatureSelectorVerticalFractions, aatLayoutFeatureSelectorNoFractions},
	{ot.NewTag('c', '2', 'p', 'c'), aatLayoutFeatureTypeUpperCase, aatLayoutFeatureSelectorUpperCasePetiteCaps, aatLayoutFeatureSelectorDefaultUpperCase},
	{ot.NewTag('c', '2', 's', 'c'), aatLayoutFeatureTypeUpperCase, aatLayoutFeatureSelectorUpperCaseSmallCaps, aatLayoutFeatureSelectorDefaultUpperCase},
	{ot.NewTag('c', 'a', 'l', 't'), aatLayoutFeatureTypeContextualAlternatives, aatLayoutFeatureSelectorContextualAlternatesOn, aatLayoutFeatureSelectorContextualAlternatesOff},
	{ot.NewTag('c', 'a', 's', 'e'), aatLayoutFeatureTypeCaseSensitiveLayout, aatLayoutFeatureSelectorCaseSensitiveLayoutOn, aatLayoutFeatureSelectorCaseSensitiveLayoutOff},
	{ot.NewTag('c', 'l', 'i', 'g'), aatLayoutFeatureTypeLigatures, aatLayoutFeatureSelectorContextualLigaturesOn, aatLayoutFeatureSelectorContextualLigaturesOff},
	{ot.NewTag('c', 'p', 's', 'p'), aatLayoutFeatureTypeCaseSensitiveLayout, aatLayoutFeatureSelectorCaseSensitiveSpacingOn, aatLayoutFeatureSelectorCaseSensitiveSpacingOff},
	{ot.NewTag('c', 's', 'w', 'h'), aatLayoutFeatureTypeContextualAlternatives, aatLayoutFeatureSelectorContextualSwashAlternatesOn, aatLayoutFeatureSelectorContextualSwashAlternatesOff},
	{ot.NewTag('d', 'l', 'i', 'g'), aatLayoutFeatureTypeLigatures, aatLayoutFeatureSelectorRareLigaturesOn, aatLayoutFeatureSelectorRareLigaturesOff},
	{ot.NewTag('e', 'x', 'p', 't'), aatLayoutFeatureTypeCharacterShape, aatLayoutFeatureSelectorExpertCharacters, 16},
	{ot.NewTag('f', 'r', 'a', 'c'), aatLayoutFeatureTypeFractions, aatLayoutFeatureSelectorDiagonalFractions, aatLayoutFeatureSelectorNoFractions},
	{ot.NewTag('f', 'w', 'i', 'd'), aatLayoutFeatureTypeTextSpacing, aatLayoutFeatureSelectorMonospacedText, 7},
	{ot.NewTag('h', 'a', 'l', 't'), aatLayoutFeatureTypeTextSpacing, aatLayoutFeatureSelectorAltHalfWidthText, 7},
	{ot.NewTag('h', 'i', 's', 't'), 40, 0, 1},
	{ot.NewTag('h', 'k', 'n', 'a'), aatLayoutFeatureTypeAlternateKana, aatLayoutFeatureSelectorAlternateHorizKanaOn, aatLayoutFeatureSelectorAlternateHorizKanaOff},
	{ot.NewTag('h', 'l', 'i', 'g'), aatLayoutFeatureTypeLigatures, aatLayoutFeatureSelectorHistoricalLigaturesOn, aatLayoutFeatureSelectorHistoricalLigaturesOff},
	{ot.NewTag('h', 'n', 'g', 'l'), aatLayoutFeatureTypeTransliteration, aatLayoutFeatureSelectorHanjaToHangul, aatLayoutFeatureSelectorNoTransliteration},
	{ot.NewTag('h', 'o', 'j', 'o'), aatLayoutFeatureTypeCharacterShape, aatLayoutFeatureSelectorHojoCharacters, 16},
	{ot.NewTag('h', 'w', 'i', 'd'), aatLayoutFeatureTypeTextSpacing, aatLayoutFeatureSelectorHalfWidthText, 7},
	{ot.NewTag('i', 't', 'a', 'l'), aatLayoutFeatureTypeItalicCjkRoman, aatLayoutFeatureSelectorCjkItalicRomanOn, aatLayoutFeatureSelectorCjkItalicRomanOff},
	{ot.NewTag('j', 'p', '0', '4'), aatLayoutFeatureTypeCharacterShape, aatLayoutFeatureSelectorJis2004Characters, 16},
	{ot.NewTag('j', 'p', '7', '8'), aatLayoutFeatureTypeCharacterShape, aatLayoutFeatureSelectorJis1978Characters, 16},
	{ot.NewTag('j', 'p', '8', '3'), aatLayoutFeatureTypeCharacterShape, aatLayoutFeatureSelectorJis1983Characters, 16},
	{ot.NewTag('j', 'p', '9', '0'), aatLayoutFeatureTypeCharacterShape, aatLayoutFeatureSelectorJis1990Characters, 16},
	{ot.NewTag('l', 'i', 'g', 'a'), aatLayoutFeatureTypeLigatures, aatLayoutFeatureSelectorCommonLigaturesOn, aatLayoutFeatureSelectorCommonLigaturesOff},
	{ot.NewTag('l', 'n', 'u', 'm'), aatLayoutFeatureTypeNumberCase, aatLayoutFeatureSelectorUpperCaseNumbers, 2},
	{ot.NewTag('m', 'g', 'r', 'k'), aatLayoutFeatureTypeMathematicalExtras, aatLayoutFeatureSelectorMathematicalGreekOn, aatLayoutFeatureSelectorMathematicalGreekOff},
	{ot.NewTag('n', 'l', 'c', 'k'), aatLayoutFeatureTypeCharacterShape, aatLayoutFeatureSelectorNlccharacters, 16},
	{ot.NewTag('o', 'n', 'u', 'm'), aatLayoutFeatureTypeNumberCase, aatLayoutFeatureSelectorLowerCaseNumbers, 2},
	{ot.NewTag('o', 'r', 'd', 'n'), aatLayoutFeatureTypeVerticalPosition, aatLayoutFeatureSelectorOrdinals, aatLayoutFeatureSelectorNormalPosition},
	{ot.NewTag('p', 'a', 'l', 't'), aatLayoutFeatureTypeTextSpacing, aatLayoutFeatureSelectorAltProportionalText, 7},
	{ot.NewTag('p', 'c', 'a', 'p'), aatLayoutFeatureTypeLowerCase, aatLayoutFeatureSelectorLowerCasePetiteCaps, aatLayoutFeatureSelectorDefaultLowerCase},
	{ot.NewTag('p', 'k', 'n', 'a'), aatLayoutFeatureTypeTextSpacing, aatLayoutFeatureSelectorProportionalText, 7},
	{ot.NewTag('p', 'n', 'u', 'm'), aatLayoutFeatureTypeNumberSpacing, aatLayoutFeatureSelectorProportionalNumbers, 4},
	{ot.NewTag('p', 'w', 'i', 'd'), aatLayoutFeatureTypeTextSpacing, aatLayoutFeatureSelectorProportionalText, 7},
	{ot.NewTag('q', 'w', 'i', 'd'), aatLayoutFeatureTypeTextSpacing, aatLayoutFeatureSelectorQuarterWidthText, 7},
	{ot.NewTag('r', 'l', 'i', 'g'), aatLayoutFeatureTypeLigatures, aatLayoutFeatureSelectorRequiredLigaturesOn, aatLayoutFeatureSelectorRequiredLigaturesOff},
	{ot.NewTag('r', 'u', 'b', 'y'), aatLayoutFeatureTypeRubyKana, aatLayoutFeatureSelectorRubyKanaOn, aatLayoutFeatureSelectorRubyKanaOff},
	{ot.NewTag('s', 'i', 'n', 'f'), aatLayoutFeatureTypeVerticalPosition, aatLayoutFeatureSelectorScientificInferiors, aatLayoutFeatureSelectorNormalPosition},
	{ot.NewTag('s', 'm', 'c', 'p'), aatLayoutFeatureTypeLowerCase, aatLayoutFeatureSelectorLowerCaseSmallCaps, aatLayoutFeatureSelectorDefaultLowerCase},
	{ot.NewTag('s', 'm', 'p', 'l'), aatLayoutFeatureTypeCharacterShape, aatLayoutFeatureSelectorSimplifiedCharacters, 16},
	{ot.NewTag('s', 's', '0', '1'), aatLayoutFeatureTypeStylisticAlternatives, aatLayoutFeatureSelectorStylisticAltOneOn, aatLayoutFeatureSelectorStylisticAltOneOff},
	{ot.NewTag('s', 's', '0', '2'), aatLayoutFeatureTypeStylisticAlternatives, aatLayoutFeatureSelectorStylisticAltTwoOn, aatLayoutFeatureSelectorStylisticAltTwoOff},
	{ot.NewTag('s', 's', '0', '3'), aatLayoutFeatureTypeStylisticAlternatives, aatLayoutFeatureSelectorStylisticAltThreeOn, aatLayoutFeatureSelectorStylisticAltThreeOff},
	{ot.NewTag('s', 's', '0', '4'), aatLayoutFeatureTypeStylisticAlternatives, aatLayoutFeatureSelectorStylisticAltFourOn, aatLayoutFeatureSelectorStylisticAltFourOff},
	{ot.NewTag('s', 's', '0', '5'), aatLayoutFeatureTypeStylisticAlternatives, aatLayoutFeatureSelectorStylisticAltFiveOn, aatLayoutFeatureSelectorStylisticAltFiveOff},
	{ot.NewTag('s', 's', '0', '6'), aatLayoutFeatureTypeStylisticAlternatives, aatLayoutFeatureSelectorStylisticAltSixOn, aatLayoutFeatureSelectorStylisticAltSixOff},
	{ot.NewTag('s', 's', '0', '7'), aatLayoutFeatureTypeStylisticAlternatives, aatLayoutFeatureSelectorStylisticAltSevenOn, aatLayoutFeatureSelectorStylisticAltSevenOff},
	{ot.NewTag('s', 's', '0', '8'), aatLayoutFeatureTypeStylisticAlternatives, aatLayoutFeatureSelectorStylisticAltEightOn, aatLayoutFeatureSelectorStylisticAltEightOff},
	{ot.NewTag('s', 's', '0', '9'), aatLayoutFeatureTypeStylisticAlternatives, aatLayoutFeatureSelectorStylisticAltNineOn, aatLayoutFeatureSelectorStylisticAltNineOff},
	{ot.NewTag('s', 's', '1', '0'), aatLayoutFeatureTypeStylisticAlternatives, aatLayoutFeatureSelectorStylisticAltTenOn, aatLayoutFeatureSelectorStylisticAltTenOff},
	{ot.NewTag('s', 's', '1', '1'), aatLayoutFeatureTypeStylisticAlternatives, aatLayoutFeatureSelectorStylisticAltElevenOn, aatLayoutFeatureSelectorStylisticAltElevenOff},
	{ot.NewTag('s', 's', '1', '2'), aatLayoutFeatureTypeStylisticAlternatives, aatLayoutFeatureSelectorStylisticAltTwelveOn, aatLayoutFeatureSelectorStylisticAltTwelveOff},
	{ot.NewTag('s', 's', '1', '3'), aatLayoutFeatureTypeStylisticAlternatives, aatLayoutFeatureSelectorStylisticAltThirteenOn, aatLayoutFeatureSelectorStylisticAltThirteenOff},
	{ot.NewTag('s', 's', '1', '4'), aatLayoutFeatureTypeStylisticAlternatives, aatLayoutFeatureSelectorStylisticAltFourteenOn, aatLayoutFeatureSelectorStylisticAltFourteenOff},
	{ot.NewTag('s', 's', '1', '5'), aatLayoutFeatureTypeStylisticAlternatives, aatLayoutFeatureSelectorStylisticAltFifteenOn, aatLayoutFeatureSelectorStylisticAltFifteenOff},
	{ot.NewTag('s', 's', '1', '6'), aatLayoutFeatureTypeStylisticAlternatives, aatLayoutFeatureSelectorStylisticAltSixteenOn, aatLayoutFeatureSelectorStylisticAltSixteenOff},
	{ot.NewTag('s', 's', '1', '7'), aatLayoutFeatureTypeStylisticAlternatives, aatLayoutFeatureSelectorStylisticAltSeventeenOn, aatLayoutFeatureSelectorStylisticAltSeventeenOff},
	{ot.NewTag('s', 's', '1', '8'), aatLayoutFeatureTypeStylisticAlternatives, aatLayoutFeatureSelectorStylisticAltEighteenOn, aatLayoutFeatureSelectorStylisticAltEighteenOff},
	{ot.NewTag('s', 's', '1', '9'), aatLayoutFeatureTypeStylisticAlternatives, aatLayoutFeatureSelectorStylisticAltNineteenOn, aatLayoutFeatureSelectorStylisticAltNineteenOff},
	{ot.NewTag('s', 's', '2', '0'), aatLayoutFeatureTypeStylisticAlternatives, aatLayoutFeatureSelectorStylisticAltTwentyOn, aatLayoutFeatureSelectorStylisticAltTwentyOff},
	{ot.NewTag('s', 'u', 'b', 's'), aatLayoutFeatureTypeVerticalPosition, aatLayoutFeatureSelectorInferiors, aatLayoutFeatureSelectorNormalPosition},
	{ot.NewTag('s', 'u', 'p', 's'), aatLayoutFeatureTypeVerticalPosition, aatLayoutFeatureSelectorSuperiors, aatLayoutFeatureSelectorNormalPosition},
	{ot.NewTag('s', 'w', 's', 'h'), aatLayoutFeatureTypeContextualAlternatives, aatLayoutFeatureSelectorSwashAlternatesOn, aatLayoutFeatureSelectorSwashAlternatesOff},
	{ot.NewTag('t', 'i', 't', 'l'), aatLayoutFeatureTypeStyleOptions, aatLayoutFeatureSelectorTitlingCaps, aatLayoutFeatureSelectorNoStyleOptions},
	{ot.NewTag('t', 'n', 'a', 'm'), aatLayoutFeatureTypeCharacterShape, aatLayoutFeatureSelectorTraditionalNamesCharacters, 16},
	{ot.NewTag('t', 'n', 'u', 'm'), aatLayoutFeatureTypeNumberSpacing, aatLayoutFeatureSelectorMonospacedNumbers, 4},
	{ot.NewTag('t', 'r', 'a', 'd'), aatLayoutFeatureTypeCharacterShape, aatLayoutFeatureSelectorTraditionalCharacters, 16},
	{ot.NewTag('t', 'w', 'i', 'd'), aatLayoutFeatureTypeTextSpacing, aatLayoutFeatureSelectorThirdWidthText, 7},
	{ot.NewTag('u', 'n', 'i', 'c'), aatLayoutFeatureTypeLetterCase, 14, 15},
	{ot.NewTag('v', 'a', 'l', 't'), aatLayoutFeatureTypeTextSpacing, aatLayoutFeatureSelectorAltProportionalText, 7},
	{ot.NewTag('v', 'e', 'r', 't'), aatLayoutFeatureTypeVerticalSubstitution, aatLayoutFeatureSelectorSubstituteVerticalFormsOn, aatLayoutFeatureSelectorSubstituteVerticalFormsOff},
	{ot.NewTag('v', 'h', 'a', 'l'), aatLayoutFeatureTypeTextSpacing, aatLayoutFeatureSelectorAltHalfWidthText, 7},
	{ot.NewTag('v', 'k', 'n', 'a'), aatLayoutFeatureTypeAlternateKana, aatLayoutFeatureSelectorAlternateVertKanaOn, aatLayoutFeatureSelectorAlternateVertKanaOff},
	{ot.NewTag('v', 'p', 'a', 'l'), aatLayoutFeatureTypeTextSpacing, aatLayoutFeatureSelectorAltProportionalText, 7},
	{ot.NewTag('v', 'r', 't', '2'), aatLayoutFeatureTypeVerticalSubstitution, aatLayoutFeatureSelectorSubstituteVerticalFormsOn, aatLayoutFeatureSelectorSubstituteVerticalFormsOff},
	{ot.NewTag('v', 'r', 't', 'r'), aatLayoutFeatureTypeVerticalSubstitution, 2, 3},
	{ot.NewTag('z', 'e', 'r', 'o'), aatLayoutFeatureTypeTypographicExtras, aatLayoutFeatureSelectorSlashedZeroOn, aatLayoutFeatureSelectorSlashedZeroOff},
}

const deletedGlyph = 0xFFFF

type aatClassCache = otLayoutMappingCache

/* Note: This context is used for kerning, even without AAT, hence the condition. */

type aatApplyContext struct {
	plan      *otShapePlan
	font      *Font
	face      Face
	buffer    *Buffer
	gdef      tables.GDEF
	ankrTable tables.Ankr

	rangeFlags    []rangeFlags
	subtableFlags GlyphMask

	bufferIsReversed bool
	// Caches
	usingBufferGlyphSet bool
	bufferGlyphSet      intSet // runes or glyphs

	firstSet          intSet        // readonly
	secondSet         intSet        // readonly
	machineClassCache aatClassCache // readonly
}

func newAatApplyContext(plan *otShapePlan, font *Font, buffer *Buffer) *aatApplyContext {
	var out aatApplyContext
	out.plan = plan
	out.font = font
	out.face = font.face
	out.buffer = buffer
	return &out
}

func (c *aatApplyContext) reverseBuffer() {
	c.buffer.Reverse()
	c.bufferIsReversed = !c.bufferIsReversed
}

func (c *aatApplyContext) setupBufferGlyphSet() {
	c.usingBufferGlyphSet = len(c.buffer.Info) >= 4
	if !c.usingBufferGlyphSet {
		return
	}
	c.buffer.collectGlyphs(&c.bufferGlyphSet)
}

func (c *aatApplyContext) bufferIntersectsMachine() bool {
	if c.usingBufferGlyphSet {
		return c.bufferGlyphSet.Intersects(c.firstSet)
	}

	// Faster for shorter buffers.
	for _, info := range c.buffer.Info {
		if c.firstSet.hasGlyph(info.Glyph) {
			return true
		}
	}
	return false
}

func (c *aatApplyContext) outputGlyphs(glyphs []GID) bool {
	if c.usingBufferGlyphSet {
		c.bufferGlyphSet.addGlyphs(glyphs)
	}
	for _, glyph := range glyphs {
		if glyph == deletedGlyph {
			c.buffer.scratchFlags |= bsfAatHasDeleted
			c.buffer.cur(0).setAatDeleted()
		} else {
			if c.gdef.GlyphClassDef != nil {
				c.buffer.cur(0).glyphProps = c.gdef.GlyphProps(gID(glyph))
			}
		}
		c.buffer.outputGlyphIndex(glyph)
	}
	return true
}

func (c *aatApplyContext) replaceGlyph(glyph GID) {
	if glyph == deletedGlyph {
		c.buffer.scratchFlags |= bsfAatHasDeleted
		c.buffer.cur(0).setAatDeleted()
	}

	if c.usingBufferGlyphSet {
		c.bufferGlyphSet.addGlyph(glyph)
	}
	if c.gdef.GlyphClassDef != nil {
		c.buffer.cur(0).glyphProps = c.gdef.GlyphProps(gID(glyph))
	}
	c.buffer.replaceGlyphIndex(glyph)
}

func (c *aatApplyContext) deleteGlyph() {
	c.buffer.scratchFlags |= bsfAatHasDeleted
	c.buffer.cur(0).setAatDeleted()
	c.buffer.replaceGlyphIndex(deletedGlyph)
}

func (c *aatApplyContext) replace_glyph_inplace(i int, glyph gID) {
	c.buffer.Info[i].Glyph = GID(glyph)
	if c.usingBufferGlyphSet {
		c.bufferGlyphSet.addGlyph(GID(glyph))
	}
	if c.gdef.GlyphClassDef != nil {
		c.buffer.Info[i].glyphProps = c.gdef.GlyphProps(glyph)
	}
}

func (c *aatApplyContext) hasAnyFlags(flag GlyphMask) bool {
	for _, fl := range c.rangeFlags {
		if fl.flags&flag != 0 {
			return true
		}
	}
	return false
}

// execute the state machine in AAT tables
type stateTableDriver struct {
	machine font.AATStateTable
}

func newStateTableDriver(machine font.AATStateTable, _ Face) stateTableDriver {
	return stateTableDriver{machine: machine}
}

func (s *stateTableDriver) getClass(glyph GID, cache *aatClassCache) uint16 {
	key := uint16(glyph)
	if v, ok := cache.get(key); ok {
		return v
	}
	v := s.machine.GetClass(glyph)
	cache.set(key, v)
	return v
}

// implemented by the subtables
type driverContext interface {
	inPlace() bool
	isActionable(entry tables.AATStateEntry) bool
	transition(buffer *Buffer, s stateTableDriver, entry tables.AATStateEntry)
}
