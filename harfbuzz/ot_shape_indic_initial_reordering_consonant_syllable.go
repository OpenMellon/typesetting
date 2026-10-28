package harfbuzz

import (
	"fmt"
	"sort"

	"github.com/OpenMellon/typesetting/language"
)

/* Rules from:
 * https://docs.microsqoft.com/en-us/typography/script-development/devanagari */
func (indicPlan *indicShapePlan) initialReorderingConsonantSyllable(font *Font, buffer *Buffer, start, end int) {
	info := buffer.Info

	/* https://github.com/harfbuzz/harfbuzz/issues/435#issuecomment-335560167
	* For compatibility with legacy usage in Kannada,
	* Ra+h+ZWJ must behave like Ra+ZWJ+h... */
	if buffer.Props.Script == language.Kannada &&
		start+3 <= end &&
		isOneOf(&info[start], 1<<indSM_ex_Ra) &&
		isOneOf(&info[start+1], 1<<indSM_ex_H) &&
		isOneOf(&info[start+2], 1<<indSM_ex_ZWJ) {
		buffer.mergeClusters(start+1, start+3)
		info[start+1], info[start+2] = info[start+2], info[start+1]
	}

	/* 1. Find base consonant:
	*
	* The shaping engine finds the base consonant of the syllable, using the
	* following algorithm: starting from the end of the syllable, move backwards
	* until a consonant is found that does not have a below-base or post-base
	* form (post-base forms have to follow below-base forms), or that is not a
	* pre-base-reordering Ra, or arrive at the first consonant. The consonant
	* stopped at will be the base.
	*
	*   o If the syllable starts with Ra + Halant (in a script that has Reph)
	*     and has more than one consonant, Ra is excluded from candidates for
	*     base consonants.
	 */

	base := end
	hasReph := false

	{
		/* . If the syllable starts with Ra + Halant (in a script that has Reph)
		 *    and has more than one consonant, Ra is excluded from candidates for
		 *    base consonants. */
		limit := start
		if indicPlan.maskArray[indicRphf] != 0 && start+3 <= end &&
			((indicPlan.config.rephMode == rephModeImplicit && !isJoiner(&info[start+2])) ||
				(indicPlan.config.rephMode == rephModeExplicit && info[start+2].complexCategory == indSM_ex_ZWJ)) {
			/* See if it matches the 'rphf' feature. */
			glyphs := [3]GID{info[start].Glyph, info[start+1].Glyph, 0}
			if indicPlan.config.rephMode == rephModeExplicit {
				glyphs[2] = info[start+2].Glyph
			}
			if indicPlan.rphf.wouldSubstitute(glyphs[:2], font) ||
				(indicPlan.config.rephMode == rephModeExplicit &&
					indicPlan.rphf.wouldSubstitute(glyphs[:3], font)) {
				limit += 2
				for limit < end && isJoiner(&info[limit]) {
					limit++
				}
				base = start
				hasReph = true
			}
		} else if indicPlan.config.rephMode == rephModeLogRepha && info[start].complexCategory == indSM_ex_Repha {
			limit += 1
			for limit < end && isJoiner(&info[limit]) {
				limit++
			}
			base = start
			hasReph = true
		}

		{
			/* . starting from the end of the syllable, move backwards */
			i := end
			seenBelow := false
			for do := true; do; do = i > limit {
				i--
				/* . until a consonant is found */
				if isConsonant(&info[i]) {
					/* . that does not have a below-base or post-base form
					 * (post-base forms have to follow below-base forms), */
					if info[i].complexAux != posBelowC &&
						(info[i].complexAux != posPostC || seenBelow) {
						base = i
						break
					}
					if info[i].complexAux == posBelowC {
						seenBelow = true
					}

					/* . or that is not a pre-base-reordering Ra,
					 *
					 * IMPLEMENTATION NOTES:
					 *
					 * Our pre-base-reordering Ra's are marked posPostC, so will be skipped
					 * by the logic above already.
					 */

					/* . or arrive at the first consonant. The consonant stopped at will
					 * be the base. */
					base = i
				} else {
					/* A ZWJ after a Halant stops the base search, and requests an explicit
					 * half form.
					 * A ZWJ before a Halant, requests a subjoined form instead, and hence
					 * search continues.  This is particularly important for Bengali
					 * sequence Ra,H,Ya that should form Ya-Phalaa by subjoining Ya. */
					if start < i &&
						info[i].complexCategory == indSM_ex_ZWJ &&
						info[i-1].complexCategory == indSM_ex_H {
						break
					}
				}
			}
		}

		/* . If the syllable starts with Ra + Halant (in a script that has Reph)
		 *    and has more than one consonant, Ra is excluded from candidates for
		 *    base consonants.
		 *
		 *  Only do this for unforced Reph. (ie. not for Ra,H,ZWJ. */
		if hasReph && base == start && limit-base <= 2 {
			/* Have no other consonant, so Reph is not formed and Ra becomes base. */
			hasReph = false
		}
	}

	/* 2. Decompose and reorder Matras:
	*
	* Each matra and any syllable modifier sign in the syllable are moved to the
	* appropriate position relative to the consonant(s) in the syllable. The
	* shaping engine decomposes two- or three-part matras into their constituent
	* parts before any repositioning. Matra characters are classified by which
	* consonant in a conjunct they have affinity for and are reordered to the
	* following positions:
	*
	*   o Before first half form in the syllable
	*   o After subjoined consonants
	*   o After post-form consonant
	*   o After main consonant (for above marks)
	*
	* IMPLEMENTATION NOTES:
	*
	* The normalize() routine has already decomposed matras for us, so we don't
	* need to worry about that.
	 */

	/* 3.  Reorder marks to canonical order:
	*
	* Adjacent nukta and halant or nukta and vedic sign are always repositioned
	* if necessary, so that the nukta is first.
	*
	* IMPLEMENTATION NOTES:
	*
	* We don't need to do this: the normalize() routine already did this for us.
	 */

	/* Reorder characters */

	for i := start; i < base; i++ {
		info[i].complexAux = min8(posPreC, info[i].complexAux)
	}

	if base < end {
		info[base].complexAux = posBaseC
	}

	/* Handle beginning Ra */
	if hasReph {
		info[start].complexAux = posRaToBecomeReph
	}

	/* For old-style Indic script tags, move the first post-base Halant after
	* last consonant.
	*
	* Reports suggest that in some scripts Uniscribe does this only if there
	* is *not* a Halant after last consonant already.  We know that is the
	* case for Kannada, while it reorders unconditionally in other scripts,
	* eg. Malayalam, Bengali, and Devanagari.  We don't currently know about
	* other scripts, so we block Kannada.
	*
	* Kannada test case:
	* U+0C9A,U+0CCD,U+0C9A,U+0CCD
	* With some versions of Lohit Kannada.
	* https://bugs.freedesktop.org/show_bug.cgi?id=59118
	*
	* Malayalam test case:
	* U+0D38,U+0D4D,U+0D31,U+0D4D,U+0D31,U+0D4D
	* With lohit-ttf-20121122/Lohit-Malayalam.ttf
	*
	* Bengali test case:
	* U+0998,U+09CD,U+09AF,U+09CD
	* With Windows XP vrinda.ttf
	* https://github.com/harfbuzz/harfbuzz/issues/1073
	*
	* Devanagari test case:
	* U+091F,U+094D,U+0930,U+094D
	* With chandas.ttf
	* https://github.com/harfbuzz/harfbuzz/issues/1071
	 */
	if indicPlan.isOldSpec {
		disallowDoubleHalants := buffer.Props.Script == language.Kannada
		for i := base + 1; i < end; i++ {
			if info[i].complexCategory == indSM_ex_H {
				var j int
				for j = end - 1; j > i; j-- {
					if isConsonant(&info[j]) ||
						(disallowDoubleHalants && info[j].complexCategory == indSM_ex_H) {
						break
					}
				}
				if info[j].complexCategory != indSM_ex_H && j > i {
					/* Move Halant to after last consonant. */
					if debugMode {
						fmt.Printf("INDIC - halant: switching glyph %d to %d (and shifting between)", i, j)
					}
					t := info[i]
					copy(info[i:j], info[i+1:])
					info[j] = t
				}
				break
			}
		}
	}

	/* Attach misc marks to previous char to move with them. */
	{
		var lastPos uint8 = posStart
		for i := start; i < end; i++ {
			if 1<<info[i].complexCategory&(joinerFlags|1<<indSM_ex_N|1<<indSM_ex_RS|1<<indSM_ex_CM|1<<indSM_ex_H) != 0 {
				info[i].complexAux = lastPos
				if info[i].complexCategory == indSM_ex_H && info[i].complexAux == posPreM {
					/*
					* Uniscribe doesn't move the Halant with Left Matra.
					* TEST: U+092B,U+093F,U+094D
					* We follow.
					 */
					for j := i; j > start; j-- {
						if info[j-1].complexAux != posPreM {
							info[i].complexAux = info[j-1].complexAux
							break
						}
					}
				}
			} else if info[i].complexAux != posSmvd {
				if info[i].complexCategory == indSM_ex_MPst &&
					i > start && info[i-1].complexCategory == indSM_ex_SM {
					info[i-1].complexAux = info[i].complexAux
				}

				lastPos = info[i].complexAux
			}
		}
	}

	/* For post-base consonants let them own anything before them
	* since the last consonant or matra. */
	{
		last := base
		for i := base + 1; i < end; i++ {
			if isConsonant(&info[i]) {
				for j := last + 1; j < i; j++ {
					if info[j].complexAux < posSmvd {
						info[j].complexAux = info[i].complexAux
					}
				}
				last = i
			} else if ic := info[i].complexCategory; ic == indSM_ex_M || ic == indSM_ex_MPst {
				last = i
			}
		}
	}

	{
		/* Use Syllable for sort accounting temporarily. */
		syllable := info[start].syllable
		for i := start; i < end; i++ {
			info[i].syllable = uint8(i - start)
		}

		/* Sit tight, rock 'n roll! */

		if debugMode {
			fmt.Printf("INDIC - post-base: sorting between glyph %d and %d\n", start, end)
		}

		subSlice := info[start:end]
		sort.SliceStable(subSlice, func(i, j int) bool { return subSlice[i].complexAux < subSlice[j].complexAux })

		// Find base again; also flip left-matra sequence.
		firstLeftMatra := end
		lastLeftMatra := end
		base = end
		for i := start; i < end; i++ {
			if info[i].complexAux == posBaseC {
				base = i
				break
			} else if info[i].complexAux == posPreM {
				if firstLeftMatra == end {
					firstLeftMatra = i
				}
				lastLeftMatra = i
			}
		}

		/* https://github.com/harfbuzz/harfbuzz/issues/3863 */
		if firstLeftMatra < lastLeftMatra {
			/* No need to merge clusters, handled later. */
			buffer.reverseRange(firstLeftMatra, lastLeftMatra+1)
			/* Reverse back nuktas, etc. */
			i := firstLeftMatra
			for j := i; j <= lastLeftMatra; j++ {
				if ic := info[j].complexCategory; ic == indSM_ex_M || ic == indSM_ex_MPst {
					buffer.reverseRange(i, j+1)
					i = j + 1
				}
			}
		}

		// Things are out-of-control for post base positions, they may shuffle
		// around like crazy.  In old-spec mode, we move halants around, so in
		// that case merge all clusters after base.  Otherwise, check the sort
		// order and merge as needed.
		// For pre-base stuff, we handle cluster issues in final reordering.
		//
		// We could use buffer.sort() for this, if there was no special
		// reordering of pre-base stuff happening later...
		// We don't want to mergeClusters all of that, which buffer.sort()
		// would.  Here's a concrete example:
		//
		// Assume there's a pre-base consonant and explicit Halant before base,
		// followed by a prebase-reordering (left) Matra:
		//
		//   C,H,ZWNJ,B,M
		//
		// At this point in reordering we would have:
		//
		//   M,C,H,ZWNJ,B
		//
		// whereas in final reordering we will bring the Matra closer to Base:
		//
		//   C,H,ZWNJ,M,B
		//
		// That's why we don't want to merge-clusters anything before the Base
		// at this point.  But if something moved from after Base to before it,
		// we should merge clusters from base to them.  In final-reordering, we
		// only move things around before base, and merge-clusters up to base.
		// These two merge-clusters from the two sides of base will interlock
		// to merge things correctly. See:
		// https://github.com/harfbuzz/harfbuzz/issues/2272
		if indicPlan.isOldSpec || end-start > 127 {
			buffer.mergeClusters(base, end)
		} else {
			/* Note!  Syllable is a one-byte field. */
			for i := base; i < end; i++ {
				if info[i].syllable != 255 {
					ma, mi := i, i
					j := start + int(info[i].syllable)
					for j != i {
						mi = min(mi, j)
						ma = max(ma, j)
						next := start + int(info[j].syllable)
						info[j].syllable = 255 /* So we don't process j later again. */
						j = next
					}
					buffer.mergeClusters(max(base, mi), ma+1)
				}
			}
		}

		/* Put syllable back in. */
		for i := start; i < end; i++ {
			info[i].syllable = syllable
		}
	}

	/* Setup masks now */

	{
		var mask GlyphMask

		/* Reph */
		for i := start; i < end && info[i].complexAux == posRaToBecomeReph; i++ {
			info[i].Mask |= indicPlan.maskArray[indicRphf]
		}

		/* Pre-base */
		mask = indicPlan.maskArray[indicHalf]
		if !indicPlan.isOldSpec &&
			indicPlan.config.blwfMode == blwfModePreAndPost {
			mask |= indicPlan.maskArray[indicBlwf]
		}
		for i := start; i < base; i++ {
			info[i].Mask |= mask
		}
		/* Base */
		mask = 0
		if base < end {
			info[base].Mask |= mask
		}
		/* Post-base */
		mask = indicPlan.maskArray[indicBlwf] |
			indicPlan.maskArray[indicAbvf] |
			indicPlan.maskArray[indicPstf]
		for i := base + 1; i < end; i++ {
			info[i].Mask |= mask
		}
	}

	if indicPlan.isOldSpec &&
		buffer.Props.Script == language.Devanagari {
		/* Old-spec eye-lash Ra needs special handling.  From the
		 * spec:
		 *
		 * "The feature 'below-base form' is applied to consonants
		 * having below-base forms and following the base consonant.
		 * The exception is vattu, which may appear below half forms
		 * as well as below the base glyph. The feature 'below-base
		 * form' will be applied to all such occurrences of Ra as well."
		 *
		 * Test case: U+0924,U+094D,U+0930,U+094d,U+0915
		 * with Sanskrit 2003 font.
		 *
		 * However, note that Ra,Halant,ZWJ is the correct way to
		 * request eyelash form of Ra, so we wouldbn't inhibit it
		 * in that sequence.
		 *
		 * Test case: U+0924,U+094D,U+0930,U+094d,U+200D,U+0915
		 */
		for i := start; i+1 < base; i++ {
			if info[i].complexCategory == indSM_ex_Ra &&
				info[i+1].complexCategory == indSM_ex_H &&
				(i+2 == base ||
					info[i+2].complexCategory != indSM_ex_ZWJ) {
				info[i].Mask |= indicPlan.maskArray[indicBlwf]
				info[i+1].Mask |= indicPlan.maskArray[indicBlwf]
			}
		}
	}

	prefLen := 2
	if indicPlan.maskArray[indicPref] != 0 && base+prefLen < end {
		/* Find a Halant,Ra sequence and mark it for pre-base-reordering processing. */
		for i := base + 1; i+prefLen-1 < end; i++ {
			var glyphs [2]GID
			for j := 0; j < prefLen; j++ {
				glyphs[j] = info[i+j].Glyph
			}
			if indicPlan.pref.wouldSubstitute(glyphs[:prefLen], font) {
				for j := 0; j < prefLen; j++ {
					info[i].Mask |= indicPlan.maskArray[indicPref]
					i++
				}
				break
			}
		}
	}

	/* Apply ZWJ/ZWNJ effects */
	for i := start + 1; i < end; i++ {
		if isJoiner(&info[i]) {
			nonJoiner := info[i].complexCategory == indSM_ex_ZWNJ
			j := i

			for do := true; do; do = (j > start && !isConsonant(&info[j])) {
				j--

				/* ZWJ/ZWNJ should disable CJCT.  They do that by simply
				 * being there, since we don't skip them for the CJCT
				 * feature (ie. F_MANUAL_ZWJ) */

				/* A ZWNJ disables HALF. */
				if nonJoiner {
					info[j].Mask &= ^indicPlan.maskArray[indicHalf]
				}

			}
		}
	}
}

func (indicPlan *indicShapePlan) initialReorderingStandaloneCluster(font *Font, buffer *Buffer, start, end int) {
	/* We treat placeholder/dotted-circle as if they are consonants, so we
	 * should just chain... */
	indicPlan.initialReorderingConsonantSyllable(font, buffer, start, end)
}

func (indicPlan *indicShapePlan) initialReorderingSyllableIndic(font *Font, buffer *Buffer, start, end int) {
	syllableType := (buffer.Info[start].syllable & 0x0F)
	switch syllableType {
	case indicVowelSyllable, indicConsonantSyllable: /* We made the vowels look like consonants.  So let's call the consonant logic! */
		indicPlan.initialReorderingConsonantSyllable(font, buffer, start, end)
	case indicBrokenCluster, indicStandaloneCluster: /* We already inserted dotted-circles, so just call the standalone_cluster. */
		indicPlan.initialReorderingStandaloneCluster(font, buffer, start, end)
	}
}

func (cs *complexShaperIndic) initialReorderingIndic(_ *otShapePlan, font *Font, buffer *Buffer) bool {
	if debugMode {
		fmt.Println("INDIC - start reordering indic initial")
	}

	cs.plan.updateConsonantPositionsIndic(font, buffer)
	ret := syllabicInsertDottedCircles(font, buffer, indicBrokenCluster,
		indSM_ex_DOTTEDCIRCLE, indSM_ex_Repha, posEnd)

	iter, count := buffer.syllableIterator()
	for start, end := iter.next(); start < count; start, end = iter.next() {
		cs.plan.initialReorderingSyllableIndic(font, buffer, start, end)
	}

	if debugMode {
		fmt.Println("INDIC - end reordering indic initial")
	}

	return ret
}
