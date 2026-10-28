package harfbuzz

import (
	"fmt"
	"sort"

	"github.com/OpenMellon/typesetting/font"
	ot "github.com/OpenMellon/typesetting/font/opentype"
	"github.com/OpenMellon/typesetting/font/opentype/tables"
)

// https://www.unicode.org/reports/tr53/
var modifierCombiningMarks = [...]rune{
	0x0654, /* ARABIC HAMZA ABOVE */
	0x0655, /* ARABIC HAMZA BELOW */
	0x0658, /* ARABIC MARK NOON GHUNNA */
	0x06DC, /* ARABIC SMALL HIGH SEEN */
	0x06E3, /* ARABIC SMALL LOW SEEN */
	0x06E7, /* ARABIC SMALL HIGH YEH */
	0x06E8, /* ARABIC SMALL HIGH NOON */
	0x08CA, /* ARABIC SMALL HIGH FARSI YEH */
	0x08CB, /* ARABIC SMALL HIGH YEH BARREE WITH TWO DOTS BELOW */
	0x08CD, /* ARABIC SMALL HIGH ZAH */
	0x08CE, /* ARABIC LARGE ROUND DOT ABOVE */
	0x08CF, /* ARABIC LARGE ROUND DOT BELOW */
	0x08D3, /* ARABIC SMALL LOW WAW */
	0x08F3, /* ARABIC SMALL HIGH WAW */
}

func infoIsMcm(info *GlyphInfo) bool {
	u := info.codepoint
	for i := 0; i < len(modifierCombiningMarks); i++ {
		if u == modifierCombiningMarks[i] {
			return true
		}
	}
	return false
}

func (cs *complexShaperArabic) reorderMarks(_ *otShapePlan, buffer *Buffer, start, end int) {
	info := buffer.Info

	if debugMode {
		fmt.Printf("ARABIC - Reordering marks from %d to %d\n", start, end)
	}

	i := start
	for cc := uint8(220); cc <= 230; cc += 10 {
		if debugMode {
			fmt.Printf("ARABIC - Looking for combining class %d's starting at %d\n", cc, i)
		}
		for i < end && info[i].getModifiedCombiningClass() < cc {
			i++
		}
		if debugMode {
			fmt.Printf("ARABIC - Looking for %d's stopped at %d\n", cc, i)
		}

		if i == end {
			break
		}

		if info[i].getModifiedCombiningClass() > cc {
			continue
		}

		j := i
		for j < end && info[j].getModifiedCombiningClass() == cc && infoIsMcm(&info[j]) {
			j++
		}

		if i == j {
			continue
		}

		if debugMode {
			fmt.Printf("ARABIC - Found %d's from %d to %d", cc, i, j)
			// shift it!
			fmt.Printf("ARABIC - Shifting %d's: %d %d", cc, i, j)
		}

		var temp [shapeComplexMaxCombiningMarks]GlyphInfo
		//  assert (j - i <= len (temp));
		buffer.mergeClusters(start, j)
		copy(temp[:j-i], info[i:])
		copy(info[start+j-i:], info[start:i])
		copy(info[start:], temp[:j-i])

		/* Renumber CC such that the reordered sequence is still sorted.
		 * 22 and 26 are chosen because they are smaller than all Arabic categories,
		 * and are folded back to 220/230 respectively during fallback mark positioning.
		 *
		 * We do this because the CGJ-handling logic in the normalizer relies on
		 * mark sequences having an increasing order even after this reordering.
		 * https://github.com/harfbuzz/harfbuzz/issues/554
		 * This, however, does break some obscure sequences, where the normalizer
		 * might compose a sequence that it should not.  For example, in the seequence
		 * ALEF, HAMZAH, MADDAH, we should NOT try to compose ALEF+MADDAH, but with this
		 * renumbering, we will. */
		newStart := start + j - i
		newCc := mcc26
		if cc == 220 {
			newCc = mcc26
		}
		for start < newStart {
			info[start].setModifiedCombiningClass(newCc)
			start++
		}

		i = j
	}
}

// Features ordered the same as the entries in [arabicShaping] rows,
// followed by rlig.  Don't change.
// We currently support one subtable per lookup, and one lookup
// per feature.  But we allow duplicate features, so we use that!
var arabicFallbackFeatures = [...]ot.Tag{
	ot.NewTag('i', 'n', 'i', 't'),
	ot.NewTag('m', 'e', 'd', 'i'),
	ot.NewTag('f', 'i', 'n', 'a'),
	ot.NewTag('i', 's', 'o', 'l'),
	ot.NewTag('r', 'l', 'i', 'g'),
	ot.NewTag('r', 'l', 'i', 'g'),
	ot.NewTag('r', 'l', 'i', 'g'),
}

// used to sort both array at the same time
type jointGlyphs struct {
	glyphs, substitutes []gID
}

func (a jointGlyphs) Len() int { return len(a.glyphs) }
func (a jointGlyphs) Swap(i, j int) {
	a.glyphs[i], a.glyphs[j] = a.glyphs[j], a.glyphs[i]
	a.substitutes[i], a.substitutes[j] = a.substitutes[j], a.substitutes[i]
}
func (a jointGlyphs) Less(i, j int) bool { return a.glyphs[i] < a.glyphs[j] }

func arabicFallbackSynthesizeLookupSingle(ft *Font, featureIndex int) *lookupGSUB {
	var glyphs, substitutes []gID

	// populate arrays
	for u := rune(firstArabicShape); u <= lastArabicShape; u++ {
		s := rune(arabicShaping[u-firstArabicShape][featureIndex])
		uGlyph, hasU := ft.face.NominalGlyph(u)
		sGlyph, hasS := ft.face.NominalGlyph(s)

		if s == 0 || !hasU || !hasS || uGlyph == sGlyph || uGlyph > 0xFFFF || sGlyph > 0xFFFF {
			continue
		}

		glyphs = append(glyphs, gID(uGlyph))
		substitutes = append(substitutes, gID(sGlyph))
	}

	if len(glyphs) == 0 {
		return nil
	}

	sort.Stable(jointGlyphs{glyphs: glyphs, substitutes: substitutes})

	return &lookupGSUB{
		LookupOptions: font.LookupOptions{Flag: otIgnoreMarks},
		Subtables: []tables.GSUBLookup{
			tables.SingleSubs{Data: tables.SingleSubstData2{
				Coverage:           tables.Coverage1{Glyphs: glyphs},
				SubstituteGlyphIDs: substitutes,
			}},
		},
	}
}

// used to sort both array at the same time
type glyphsIndirections struct {
	glyphs       []gID
	indirections []int
}

func (a glyphsIndirections) Len() int { return len(a.glyphs) }
func (a glyphsIndirections) Swap(i, j int) {
	a.glyphs[i], a.glyphs[j] = a.glyphs[j], a.glyphs[i]
	a.indirections[i], a.indirections[j] = a.indirections[j], a.indirections[i]
}
func (a glyphsIndirections) Less(i, j int) bool { return a.glyphs[i] < a.glyphs[j] }

func arabicFallbackSynthesizeLookupLigature(ft *Font, ligatureTable []arabicTableEntry, lookupFlags uint16) *lookupGSUB {
	var (
		firstGlyphs            []gID
		firstGlyphsIndirection []int // original index into ArabicLigatures
	)

	// Populate arrays

	// sort out the first-glyphs
	for firstGlyphIdx, lig := range ligatureTable {
		firstGlyph, ok := ft.face.NominalGlyph(lig.First)
		if !ok {
			continue
		}
		firstGlyphs = append(firstGlyphs, gID(firstGlyph))
		firstGlyphsIndirection = append(firstGlyphsIndirection, firstGlyphIdx)
	}

	if len(firstGlyphs) == 0 {
		return nil
	}

	sort.Stable(glyphsIndirections{glyphs: firstGlyphs, indirections: firstGlyphsIndirection})

	var out tables.LigatureSubs
	out.Coverage = tables.Coverage1{Glyphs: firstGlyphs}

	// now that the first-glyphs are sorted, walk again, populate ligatures.
	for _, firstGlyphIdx := range firstGlyphsIndirection {
		ligs := ligatureTable[firstGlyphIdx].Ligatures
		var ligatureSet tables.LigatureSet
		for _, v := range ligs {
			ligatureU := v.ligature
			ligatureGlyph, hasLigature := ft.face.NominalGlyph(ligatureU)
			if !hasLigature {
				continue
			}

			components := v.components
			var componentGIDs []gID
			for _, componentU := range components {
				componentGlyph, hasComponent := ft.face.NominalGlyph(componentU)
				if !hasComponent {
					break
				}
				componentGIDs = append(componentGIDs, gID(componentGlyph))
			}

			if len(components) != len(componentGIDs) {
				continue
			}

			ligatureSet.Ligatures = append(ligatureSet.Ligatures, tables.Ligature{
				LigatureGlyph:     gID(ligatureGlyph),
				ComponentGlyphIDs: componentGIDs, // ligatures are 2-component
			})
		}
		out.LigatureSets = append(out.LigatureSets, ligatureSet)
	}

	return &lookupGSUB{
		LookupOptions: font.LookupOptions{Flag: lookupFlags},
		Subtables: []tables.GSUBLookup{
			out,
		},
	}
}

func arabicFallbackSynthesizeLookup(font *Font, featureIndex int) *lookupGSUB {
	switch featureIndex {
	case 0, 1, 2, 3:
		return arabicFallbackSynthesizeLookupSingle(font, featureIndex)
	case 4:
		return arabicFallbackSynthesizeLookupLigature(font, arabicLigature3Table[:], otIgnoreMarks)
	case 5:
		return arabicFallbackSynthesizeLookupLigature(font, arabicLigatureTable[:], otIgnoreMarks)
	case 6:
		return arabicFallbackSynthesizeLookupLigature(font, arabicLigatureMarkTable[:], 0)
	default:
		panic("unexpected arabic fallback feature index")
	}
}

const arabicFallbackMaxLookups = len(arabicFallbackFeatures)

type arabicFallbackPlan struct {
	accelArray [arabicFallbackMaxLookups]otLayoutLookupAccelerator
	numLookups int
	maskArray  [arabicFallbackMaxLookups]GlyphMask
}

func (fbPlan *arabicFallbackPlan) initWin1256(plan *otShapePlan, font *Font) bool {
	// does this font look like it's Windows-1256-encoded?
	g1, _ := font.face.NominalGlyph(0x0627) /* ALEF */
	g2, _ := font.face.NominalGlyph(0x0644) /* LAM */
	g3, _ := font.face.NominalGlyph(0x0649) /* ALEF MAKSURA */
	g4, _ := font.face.NominalGlyph(0x064A) /* YEH */
	g5, _ := font.face.NominalGlyph(0x0652) /* SUKUN */
	if !(g1 == 199 && g2 == 225 && g3 == 236 && g4 == 237 && g5 == 250) {
		return false
	}

	var j int
	for _, man := range arabicWin1256GsubLookups {
		fbPlan.maskArray[j] = plan.otMap.getMask1(man.tag)
		if fbPlan.maskArray[j] != 0 {
			if man.lookup != nil {
				fbPlan.accelArray[j].init(*man.lookup)
				j++
			}
		}
	}

	fbPlan.numLookups = j

	return j > 0
}

func (fbPlan *arabicFallbackPlan) initUnicode(plan *otShapePlan, font *Font) bool {
	var j int
	for i, feat := range arabicFallbackFeatures {
		fbPlan.maskArray[j] = plan.otMap.getMask1(feat)
		if fbPlan.maskArray[j] != 0 {
			lk := arabicFallbackSynthesizeLookup(font, i)
			if lk != nil {
				fbPlan.accelArray[j].init(*lk)
				j++
			}
		}
	}

	fbPlan.numLookups = j

	return j > 0
}

func newArabicFallbackPlan(plan *otShapePlan, font *Font) *arabicFallbackPlan {
	var fbPlan arabicFallbackPlan

	/* Try synthesizing GSUB table using Unicode Arabic Presentation Forms,
	* in case the font has cmap entries for the presentation-forms characters. */
	if fbPlan.initUnicode(plan, font) {
		return &fbPlan
	}

	/* See if this looks like a Windows-1256-encoded font. If it does, use a
	* hand-coded GSUB table. */
	if fbPlan.initWin1256(plan, font) {
		return &fbPlan
	}

	return &arabicFallbackPlan{}
}

func (fbPlan *arabicFallbackPlan) shape(font *Font, buffer *Buffer) {
	var c otApplyContext
	c.reset(0, font, buffer)
	for i := 0; i < fbPlan.numLookups; i++ {
		if fbPlan.accelArray[i].lookup != nil {
			c.setLookupMask(fbPlan.maskArray[i])
			c.substituteLookup(&fbPlan.accelArray[i])
		}
	}
}
