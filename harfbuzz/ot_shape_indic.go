package harfbuzz

import (
	ot "github.com/OpenMellon/typesetting/font/opentype"
	"github.com/OpenMellon/typesetting/font/opentype/tables"
	"github.com/OpenMellon/typesetting/language"
)

// ported from harfbuzz/src/hb-ot-shape-complex-indic.cc, .hh Copyright © 2011,2012  Google, Inc.  Behdad Esfahbod

// Keep in sync with the code generator.
const (
	posStart = iota

	posRaToBecomeReph
	posPreM
	posPreC

	posBaseC
	posAfterMain

	posAboveC

	posBeforeSub
	posBelowC
	posAfterSub

	posBeforePost
	posPostC
	posAfterPost

	posSmvd

	posEnd
)

var _ otComplexShaper = (*complexShaperIndic)(nil)

// Indic shaper.
type complexShaperIndic struct {
	complexShaperNil

	plan indicShapePlan
}

/* Note:
 *
 * We treat Vowels and placeholders as if they were consonants.  This is safe because Vowels
 * cannot happen in a consonant syllable.  The plus side however is, we can call the
 * consonant syllable logic from the vowel syllable function and get it all right! */
const (
	consonantFlags = 1<<indSM_ex_C | 1<<indSM_ex_CS |
		1<<indSM_ex_Ra | 1<<indSM_ex_CM | 1<<indSM_ex_V |
		1<<indSM_ex_PLACEHOLDER | 1<<indSM_ex_DOTTEDCIRCLE
	joinerFlags = 1<<indSM_ex_ZWJ | 1<<indSM_ex_ZWNJ
)

func isOneOf(info *GlyphInfo, flags uint32) bool {
	/* If it ligated, all bets are off. */
	if info.ligated() {
		return false
	}
	return 1<<info.complexCategory&flags != 0
}

func isJoiner(info *GlyphInfo) bool {
	return isOneOf(info, joinerFlags)
}

func isConsonant(info *GlyphInfo) bool {
	return isOneOf(info, consonantFlags)
}

func isHalant(info *GlyphInfo) bool {
	return isOneOf(info, 1<<indSM_ex_H)
}

func (info *GlyphInfo) setIndicProperties() {
	u := info.codepoint
	type_ := indicGetCategories(u)
	info.complexCategory, info.complexAux = uint8(type_&0xFF), uint8(type_>>8)
}

type indicWouldSubstituteFeature struct {
	lookups []lookupMap
	//   count int
	zeroContext bool
}

func newIndicWouldSubstituteFeature(map_ *otMap, featureTag tables.Tag, zeroContext bool) indicWouldSubstituteFeature {
	var out indicWouldSubstituteFeature
	out.zeroContext = zeroContext
	out.lookups = map_.getStageLookups(0 /*GSUB*/, map_.getFeatureStage(0 /*GSUB*/, featureTag))
	return out
}

func (ws indicWouldSubstituteFeature) wouldSubstitute(glyphs []GID, font *Font) bool {
	for _, lk := range ws.lookups {
		if otLayoutLookupWouldSubstitute(font, lk.index, glyphs, ws.zeroContext) {
			return true
		}
	}
	return false
}

/*
 * Indic configurations.  Note that we do not want to keep every single script-specific
 * behavior in these tables necessarily.  This should mainly be used for per-script
 * properties that are cheaper keeping here, than in the code.  Ie. if, say, one and
 * only one script has an exception, that one script can be if'ed directly in the code,
 * instead of adding a new flag in these structs.
 */

// reph_position_t
const (
	rephPosAfterMain  = posAfterMain
	rephPosBeforeSub  = posBeforeSub
	rephPosAfterSub   = posAfterSub
	rephPosBeforePost = posBeforePost
	rephPosAfterPost  = posAfterPost
)

// reph_mode_t
const (
	rephModeImplicit = iota /* Reph formed out of initial Ra,H sequence. */
	rephModeExplicit        /* Reph formed out of initial Ra,H,ZWJ sequence. */
	rephModeLogRepha        /* Encoded Repha character, needs reordering. */
)

// blwf_mode_t
const (
	blwfModePreAndPost = iota /* Below-forms feature applied to pre-base and post-base. */
	blwfModePostOnly          /* Below-forms feature applied to post-base only. */
)

type indicConfig struct {
	script     language.Script
	hasOldSpec bool
	virama     rune
	rephPos    uint8
	rephMode   uint8
	blwfMode   uint8
}

var indicConfigs = [...]indicConfig{
	/* Default.  Should be first. */
	{0, false, 0, rephPosBeforePost, rephModeImplicit, blwfModePreAndPost},
	{language.Devanagari, true, 0x094D, rephPosBeforePost, rephModeImplicit, blwfModePreAndPost},
	{language.Bengali, true, 0x09CD, rephPosAfterSub, rephModeImplicit, blwfModePreAndPost},
	{language.Gurmukhi, true, 0x0A4D, rephPosBeforeSub, rephModeImplicit, blwfModePreAndPost},
	{language.Gujarati, true, 0x0ACD, rephPosBeforePost, rephModeImplicit, blwfModePreAndPost},
	{language.Oriya, true, 0x0B4D, rephPosAfterMain, rephModeImplicit, blwfModePreAndPost},
	{language.Tamil, true, 0x0BCD, rephPosAfterPost, rephModeImplicit, blwfModePreAndPost},
	{language.Telugu, true, 0x0C4D, rephPosAfterPost, rephModeExplicit, blwfModePostOnly},
	{language.Kannada, true, 0x0CCD, rephPosAfterPost, rephModeImplicit, blwfModePostOnly},
	{language.Malayalam, true, 0x0D4D, rephPosAfterMain, rephModeLogRepha, blwfModePreAndPost},
}

var indicFeatures = [...]otMapFeature{
	/*
	* Basic features.
	* These features are applied in order, one at a time, after initial_reordering.
	 */
	{ot.NewTag('n', 'u', 'k', 't'), ffGlobalManualJoiners | ffPerSyllable},
	{ot.NewTag('a', 'k', 'h', 'n'), ffGlobalManualJoiners | ffPerSyllable},
	{ot.NewTag('r', 'p', 'h', 'f'), ffManualJoiners | ffPerSyllable},
	{ot.NewTag('r', 'k', 'r', 'f'), ffGlobalManualJoiners | ffPerSyllable},
	{ot.NewTag('p', 'r', 'e', 'f'), ffManualJoiners | ffPerSyllable},
	{ot.NewTag('b', 'l', 'w', 'f'), ffManualJoiners | ffPerSyllable},
	{ot.NewTag('a', 'b', 'v', 'f'), ffManualJoiners | ffPerSyllable},
	{ot.NewTag('h', 'a', 'l', 'f'), ffManualJoiners | ffPerSyllable},
	{ot.NewTag('p', 's', 't', 'f'), ffManualJoiners | ffPerSyllable},
	{ot.NewTag('v', 'a', 't', 'u'), ffGlobalManualJoiners | ffPerSyllable},
	{ot.NewTag('c', 'j', 'c', 't'), ffGlobalManualJoiners | ffPerSyllable},
	/*
	* Other features.
	* These features are applied all at once, after final_reordering
	* but before clearing syllables.
	* Default Bengali font in Windows for example has intermixed
	* lookups for init,pres,abvs,blws features.
	 */
	{ot.NewTag('i', 'n', 'i', 't'), ffManualJoiners | ffPerSyllable},
	{ot.NewTag('p', 'r', 'e', 's'), ffGlobalManualJoiners | ffPerSyllable},
	{ot.NewTag('a', 'b', 'v', 's'), ffGlobalManualJoiners | ffPerSyllable},
	{ot.NewTag('b', 'l', 'w', 's'), ffGlobalManualJoiners | ffPerSyllable},
	{ot.NewTag('p', 's', 't', 's'), ffGlobalManualJoiners | ffPerSyllable},
	{ot.NewTag('h', 'a', 'l', 'n'), ffGlobalManualJoiners | ffPerSyllable},
}

// in the same order as the indicFeatures array
const (
	indicNukt = iota
	indicAkhn
	indicRphf
	indicRkrf
	indicPref
	indicBlwf
	indicAbvf
	indicHalf
	indicPstf
	indicVatu
	indicCjct

	indicInit
	indicPres
	indicAbvs
	indicBlws
	indicPsts
	indicHaln

	indicNumFeatures
	indicBasicFeatures = indicInit /* Don't forget to update this! */
)

func (cs *complexShaperIndic) collectFeatures(plan *otShapePlanner) {
	map_ := &plan.otMap

	/* Do this before any lookups have been applied. */
	map_.addGSUBPause(setupSyllablesIndic)

	map_.enableFeatureExt(ot.NewTag('l', 'o', 'c', 'l'), ffPerSyllable, 1)
	/* The Indic specs do not require ccmp, but we apply it here since if
	* there is a use of it, it's typically at the beginning. */
	map_.enableFeatureExt(ot.NewTag('c', 'c', 'm', 'p'), ffPerSyllable, 1)

	i := 0
	map_.addGSUBPause(cs.initialReorderingIndic)

	for ; i < indicBasicFeatures; i++ {
		map_.addFeatureExt(indicFeatures[i].tag, indicFeatures[i].flags, 1)
		map_.addGSUBPause(nil)
	}

	map_.addGSUBPause(cs.plan.finalReorderingIndic)

	for ; i < indicNumFeatures; i++ {
		map_.addFeatureExt(indicFeatures[i].tag, indicFeatures[i].flags, 1)
	}
}

func (complexShaperIndic) overrideFeatures(plan *otShapePlanner) {
	plan.otMap.disableFeature(ot.NewTag('l', 'i', 'g', 'a'))
	plan.otMap.addGSUBPause(nil)
}

type indicShapePlan struct {
	blwf indicWouldSubstituteFeature
	pstf indicWouldSubstituteFeature
	vatu indicWouldSubstituteFeature
	rphf indicWouldSubstituteFeature
	pref indicWouldSubstituteFeature

	maskArray   [indicNumFeatures]GlyphMask
	config      indicConfig
	viramaGlyph GID // cached value

	isOldSpec bool
}

func (indicPlan *indicShapePlan) loadViramaGlyph(font *Font) GID {
	if indicPlan.viramaGlyph == ^GID(0) {
		glyph, ok := font.face.NominalGlyph(indicPlan.config.virama)
		if indicPlan.config.virama == 0 || !ok {
			glyph = 0
		}
		/* Technically speaking, the spec says we should apply 'locl' to virama too.
		* Maybe one day... */

		/* Our get_nominal_glyph() function needs a font, so we can't get the virama glyph
		* during shape planning...  Instead, overwrite it here. */
		indicPlan.viramaGlyph = glyph
	}

	return indicPlan.viramaGlyph
}

func (cs *complexShaperIndic) dataCreate(plan *otShapePlan) {
	var indicPlan indicShapePlan

	indicPlan.config = indicConfigs[0]
	for i := 1; i < len(indicConfigs); i++ {
		if plan.props.Script == indicConfigs[i].script {
			indicPlan.config = indicConfigs[i]
			break
		}
	}

	indicPlan.isOldSpec = indicPlan.config.hasOldSpec && ((plan.otMap.chosenScript[0] & 0x000000FF) != '2')
	indicPlan.viramaGlyph = ^GID(0)

	/* Use zero-context wouldSubstitute() matching for new-spec of the main
	* Indic scripts, and scripts with one spec only, but not for old-specs.
	* The new-spec for all dual-spec scripts says zero-context matching happens.
	*
	* However, testing with Malayalam shows that old and new spec both allow
	* context.  Testing with Bengali new-spec however shows that it doesn't.
	* So, the heuristic here is the way it is.  It should *only* be changed,
	* as we discover more cases of what Windows does.  DON'T TOUCH OTHERWISE. */
	zeroContext := !indicPlan.isOldSpec && plan.props.Script != language.Malayalam
	indicPlan.rphf = newIndicWouldSubstituteFeature(&plan.otMap, ot.NewTag('r', 'p', 'h', 'f'), zeroContext)
	indicPlan.pref = newIndicWouldSubstituteFeature(&plan.otMap, ot.NewTag('p', 'r', 'e', 'f'), zeroContext)
	indicPlan.blwf = newIndicWouldSubstituteFeature(&plan.otMap, ot.NewTag('b', 'l', 'w', 'f'), zeroContext)
	indicPlan.pstf = newIndicWouldSubstituteFeature(&plan.otMap, ot.NewTag('p', 's', 't', 'f'), zeroContext)
	indicPlan.vatu = newIndicWouldSubstituteFeature(&plan.otMap, ot.NewTag('v', 'a', 't', 'u'), zeroContext)

	for i := range indicPlan.maskArray {
		if indicFeatures[i].flags&ffGLOBAL != 0 {
			indicPlan.maskArray[i] = 0
		} else {
			indicPlan.maskArray[i] = plan.otMap.getMask1(indicFeatures[i].tag)
		}
	}

	cs.plan = indicPlan
}

func (indicPlan *indicShapePlan) consonantPositionFromFace(consonant, virama GID, font *Font) uint8 {
	/* For old-spec, the order of glyphs is Consonant,Virama,
	* whereas for new-spec, it's Virama,Consonant.  However,
	* some broken fonts (like Free Sans) simply copied lookups
	* from old-spec to new-spec without modification.
	* And oddly enough, Uniscribe seems to respect those lookups.
	* Eg. in the sequence U+0924,U+094D,U+0930, Uniscribe finds
	* base at 0.  The font however, only has lookups matching
	* 930,94D in 'blwf', not the expected 94D,930 (with new-spec
	* table).  As such, we simply match both sequences.  Seems
	* to work.
	*
	* Vatu is done as well, for:
	* https://github.com/harfbuzz/harfbuzz/issues/1587
	 */
	glyphs := [3]GID{virama, consonant, virama}
	if indicPlan.blwf.wouldSubstitute(glyphs[0:2], font) ||
		indicPlan.blwf.wouldSubstitute(glyphs[1:3], font) ||
		indicPlan.vatu.wouldSubstitute(glyphs[0:2], font) ||
		indicPlan.vatu.wouldSubstitute(glyphs[1:3], font) {
		return posBelowC
	}
	if indicPlan.pstf.wouldSubstitute(glyphs[0:2], font) ||
		indicPlan.pstf.wouldSubstitute(glyphs[1:3], font) {
		return posPostC
	}
	if indicPlan.pref.wouldSubstitute(glyphs[0:2], font) ||
		indicPlan.pref.wouldSubstitute(glyphs[1:3], font) {
		return posPostC
	}
	return posBaseC
}

func (cs *complexShaperIndic) setupMasks(plan *otShapePlan, buffer *Buffer, _ *Font) {
	/* We cannot setup masks here.  We save information about characters
	* and setup masks later on in a pause-callback. */

	info := buffer.Info
	for i := range info {
		info[i].setIndicProperties()
	}
}

func setupSyllablesIndic(_ *otShapePlan, _ *Font, buffer *Buffer) bool {
	findSyllablesIndic(buffer)
	iter, count := buffer.syllableIterator()
	for start, end := iter.next(); start < count; start, end = iter.next() {
		buffer.unsafeToBreak(start, end)
	}
	return false
}

func foundSyllableIndic(syllableType uint8, ts, te int, info []GlyphInfo, syllableSerial *uint8) {
	for i := ts; i < te; i++ {
		info[i].syllable = (*syllableSerial << 4) | syllableType
	}
	*syllableSerial++
	if *syllableSerial == 16 {
		*syllableSerial = 1
	}
}

func (indicPlan *indicShapePlan) updateConsonantPositionsIndic(font *Font, buffer *Buffer) {
	virama := indicPlan.loadViramaGlyph(font)
	if virama != 0 {
		info := buffer.Info
		for i := range info {
			if info[i].complexAux == posBaseC {
				consonant := info[i].Glyph
				info[i].complexAux = indicPlan.consonantPositionFromFace(consonant, virama, font)
			}
		}
	}
}
