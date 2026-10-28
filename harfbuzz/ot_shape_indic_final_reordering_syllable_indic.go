package harfbuzz

import (
	"fmt"

	ucd "github.com/OpenMellon/typesetting/internal/unicodedata"
	"github.com/OpenMellon/typesetting/language"
)

func (indicPlan *indicShapePlan) finalReorderingSyllableIndic(buffer *Buffer, start, end int) {
	info := buffer.Info

	/* This function relies heavily on halant glyphs.  Lots of ligation
	* and possibly multiple substitutions happened prior to this
	* phase, and that might have messed up our properties.  Recover
	* from a particular case of that where we're fairly sure that a
	* class of otH is desired but has been lost. */
	/* We don't call loadViramaGlyph(), since we know it's already
	* loaded. */
	viramaGlyph := indicPlan.viramaGlyph
	if viramaGlyph != 0 {
		for i := start; i < end; i++ {
			if info[i].Glyph == viramaGlyph &&
				info[i].ligated() && info[i].multiplied() {
				/* This will make sure that this glyph passes isHalant() test. */
				info[i].complexCategory = indSM_ex_H
				info[i].clearLigatedAndMultiplied()
			}
		}
	}

	/* 4. Final reordering:
	*
	* After the localized forms and basic shaping forms GSUB features have been
	* applied (see below), the shaping engine performs some final glyph
	* reordering before applying all the remaining font features to the entire
	* syllable.
	 */

	tryPref := indicPlan.maskArray[indicPref] != 0

	/* Find base again */
	var base int
	for base = start; base < end; base++ {
		if info[base].complexAux >= posBaseC {
			if tryPref && base+1 < end {
				for i := base + 1; i < end; i++ {
					if (info[i].Mask & indicPlan.maskArray[indicPref]) != 0 {
						if !(info[i].substituted() && info[i].ligatedAndDidntMultiply()) {
							/* Ok, this was a 'pref' candidate but didn't form any.
							* Base is around here... */
							base = i
							for base < end && isHalant(&info[base]) {
								base++
							}
							if base < end {
								info[base].complexAux = posBaseC
							}

							tryPref = false
						}
						break
					}
					if base == end {
						break
					}
				}
			}
			/* For Malayalam, skip over unformed below- (but NOT post-) forms. */
			if buffer.Props.Script == language.Malayalam {
				for i := base + 1; i < end; i++ {
					for i < end && isJoiner(&info[i]) {
						i++
					}
					if i == end || !isHalant(&info[i]) {
						break
					}
					i++ /* Skip halant. */
					for i < end && isJoiner(&info[i]) {
						i++
					}
					if i < end && isConsonant(&info[i]) && info[i].complexAux == posBelowC {
						base = i
						info[base].complexAux = posBaseC
					}
				}
			}

			if start < base && info[base].complexAux > posBaseC {
				base--
			}
			break
		}
	}
	if base == end && start < base && isOneOf(&info[base-1], 1<<indSM_ex_ZWJ) {
		base--
	}
	if base < end {
		for start < base && isOneOf(&info[base], 1<<indSM_ex_N|1<<indSM_ex_H) {
			base--
		}
	}

	/*   o Reorder matras:
	*
	*     If a pre-base matra character had been reordered before applying basic
	*     features, the glyph can be moved closer to the main consonant based on
	*     whether half-forms had been formed. Actual position for the matra is
	*     defined as “after last standalone halant glyph, after initial matra
	*     position and before the main consonant”. If ZWJ or ZWNJ follow this
	*     halant, position is moved after it.
	*
	* IMPLEMENTATION NOTES:
	*
	* It looks like the last sentence is wrong.  Testing, with Windows 7 Uniscribe
	* and Devanagari shows that the behavior is best described as:
	*
	* "If ZWJ follows this halant, matra is NOT repositioned after this halant.
	*  If ZWNJ follows this halant, position is moved after it."
	*
	* Test case, with Adobe Devanagari or Nirmala UI:
	*
	*   U+091F,U+094D,U+200C,U+092F,U+093F
	*   (Matra moves to the middle, after ZWNJ.)
	*
	*   U+091F,U+094D,U+200D,U+092F,U+093F
	*   (Matra does NOT move, stays to the left.)
	*
	* https://github.com/harfbuzz/harfbuzz/issues/1070
	 */

	if start+1 < end && start < base /* Otherwise there can't be any pre-base matra characters. */ {
		/* If we lost track of base, alas, position before last thingy. */
		newPos := base - 1
		if base == end {
			newPos = base - 2
		}

		/* Malayalam / Tamil do not have "half" forms or explicit virama forms.
		 * The glyphs formed by 'half' are Chillus or ligated explicit viramas.
		 * We want to position matra after them.
		 */
		if buffer.Props.Script != language.Malayalam && buffer.Props.Script != language.Tamil {
		search:
			for newPos > start && !isOneOf(&info[newPos], 1<<indSM_ex_M|1<<indSM_ex_MPst|1<<indSM_ex_H) {
				newPos--
			}
			/* If we found no Halant we are done.
			* Otherwise only proceed if the Halant does
			* not belong to the Matra itself! */
			if isHalant(&info[newPos]) && info[newPos].complexAux != posPreM {
				if newPos+1 < end {
					/* . If ZWJ follows this halant, matra is NOT repositioned after this halant. */
					if info[newPos+1].complexCategory == indSM_ex_ZWJ {
						/* Keep searching. */
						if newPos > start {
							newPos--
							goto search
						}
					}
					/* . If ZWNJ follows this halant, position is moved after it.
					*
					* IMPLEMENTATION NOTES:
					*
					* This is taken care of by the state-machine. A Halant,ZWNJ is a terminating
					* sequence for a consonant syllable; any pre-base matras occurring after it
					* will belong to the subsequent syllable.
					 */
				}
			} else {
				newPos = start /* No move. */
			}
		}

		if start < newPos && info[newPos].complexAux != posPreM {
			/* Now go see if there's actually any matras... */
			for i := newPos; i > start; i-- {
				if info[i-1].complexAux == posPreM {
					oldPos := i - 1
					if oldPos < base && base <= newPos { /* Shouldn't actually happen. */
						base--
					}

					if debugMode {
						fmt.Printf("INDIC - matras: switching glyph %d to %d (and shifting between)", oldPos, newPos)
					}

					tmp := info[oldPos]
					copy(info[oldPos:newPos], info[oldPos+1:])
					info[newPos] = tmp

					/* Note: this mergeClusters() is intentionally *after* the reordering.
					* Indic matra reordering is special and tricky... */
					buffer.mergeClusters(newPos, min(end, base+1))

					newPos--
				}
			}
		} else {
			for i := start; i < base; i++ {
				if info[i].complexAux == posPreM {
					buffer.mergeClusters(i, min(end, base+1))
					break
				}
			}
		}
	}

	/*   o Reorder reph:
	*
	*     Reph’s original position is always at the beginning of the syllable,
	*     (i.e. it is not reordered at the character reordering stage). However,
	*     it will be reordered according to the basic-forms shaping results.
	*     Possible positions for reph, depending on the script, are; after main,
	*     before post-base consonant forms, and after post-base consonant forms.
	 */

	/* Two cases:
	*
	* - If repha is encoded as a sequence of characters (Ra,H or Ra,H,ZWJ), then
	*   we should only move it if the sequence ligated to the repha form.
	*
	* - If repha is encoded separately and in the logical position, we should only
	*   move it if it did NOT ligate.  If it ligated, it's probably the font trying
	*   to make it work without the reordering.
	 */
	if start+1 < end && info[start].complexAux == posRaToBecomeReph &&
		(info[start].complexCategory == indSM_ex_Repha) != info[start].ligatedAndDidntMultiply() {
		var newRephPos int
		rephPos := indicPlan.config.rephPos

		/*       1. If reph should be positioned after post-base consonant forms,
		 *          proceed to step 5.
		 */
		if rephPos == rephPosAfterPost {
			goto reph_step_5
		}

		/*       2. If the reph repositioning class is not after post-base: target
		 *          position is after the first explicit halant glyph between the
		 *          first post-reph consonant and last main consonant. If ZWJ or ZWNJ
		 *          are following this halant, position is moved after it. If such
		 *          position is found, this is the target position. Otherwise,
		 *          proceed to the next step.
		 *
		 *          Note: in old-implementation fonts, where classifications were
		 *          fixed in shaping engine, there was no case where reph position
		 *          will be found on this step.
		 */
		{
			newRephPos = start + 1
			for newRephPos < base && !isHalant(&info[newRephPos]) {
				newRephPos++
			}

			if newRephPos < base && isHalant(&info[newRephPos]) {
				/* .If ZWJ or ZWNJ are following this halant, position is moved after it. */
				if newRephPos+1 < base && isJoiner(&info[newRephPos+1]) {
					newRephPos++
				}
				goto reph_move
			}
		}

		/*       3. If reph should be repositioned after the main consonant: find the
		 *          first consonant not ligated with main, or find the first
		 *          consonant that is not a potential pre-base-reordering Ra.
		 */
		if rephPos == rephPosAfterMain {
			newRephPos = base
			for newRephPos+1 < end && info[newRephPos+1].complexAux <= posAfterMain {
				newRephPos++
			}
			if newRephPos < end {
				goto reph_move
			}
		}

		/*       4. If reph should be positioned before post-base consonant, find
		 *          first post-base classified consonant not ligated with main. If no
		 *          consonant is found, the target position should be before the
		 *          first matra, syllable modifier sign or vedic sign.
		 */
		/* This is our take on what step 4 is trying to say (and failing, BADLY). */
		if rephPos == rephPosAfterSub {
			newRephPos = base
			for newRephPos+1 < end &&
				(1<<info[newRephPos+1].complexAux)&(1<<posPostC|1<<posAfterPost|1<<posSmvd) == 0 {
				newRephPos++
			}
			if newRephPos < end {
				goto reph_move
			}
		}

		/*       5. If no consonant is found in steps 3 or 4, move reph to a position
		 *          immediately before the first post-base matra, syllable modifier
		 *          sign or vedic sign that has a reordering class after the intended
		 *          reph position. For example, if the reordering position for reph
		 *          is post-main, it will skip above-base matras that also have a
		 *          post-main position.
		 */
	reph_step_5:
		{
			/* Copied from step 2. */
			newRephPos = start + 1
			for newRephPos < base && !isHalant(&info[newRephPos]) {
				newRephPos++
			}

			if newRephPos < base && isHalant(&info[newRephPos]) {
				/* .If ZWJ or ZWNJ are following this halant, position is moved after it. */
				if newRephPos+1 < base && isJoiner(&info[newRephPos+1]) {
					newRephPos++
				}
				goto reph_move
			}
		}
		/* See https://github.com/harfbuzz/harfbuzz/issues/2298#issuecomment-615318654 */

		/*       6. Otherwise, reorder reph to the end of the syllable.
		 */
		{
			newRephPos = end - 1
			for newRephPos > start && info[newRephPos].complexAux == posSmvd {
				newRephPos--
			}

			/*
			* If the Reph is to be ending up after a Matra,Halant sequence,
			* position it before that Halant so it can interact with the Matra.
			* However, if it's a plain Consonant,Halant we shouldn't do that.
			* Uniscribe doesn't do this.
			* TEST: U+0930,U+094D,U+0915,U+094B,U+094D
			 */
			if isHalant(&info[newRephPos]) {
				for i := base + 1; i < newRephPos; i++ {
					if ic := info[i].complexCategory; ic == indSM_ex_M || ic == indSM_ex_MPst {
						/* Ok, got it. */
						newRephPos--
					}
				}
			}

			goto reph_move
		}

	reph_move:
		{

			if debugMode {
				fmt.Printf("INDIC - reph: switching glyph %d to %d (and shifting between)", start, newRephPos)
			}

			/* Move */
			buffer.mergeClusters(start, newRephPos+1)
			reph := info[start]
			copy(info[start:newRephPos], info[start+1:])
			info[newRephPos] = reph

			if start < base && base <= newRephPos {
				base--
			}
		}
	}

	/*   o Reorder pre-base-reordering consonants:
	*
	*     If a pre-base-reordering consonant is found, reorder it according to
	*     the following rules:
	 */

	if tryPref && base+1 < end /* Otherwise there can't be any pre-base-reordering Ra. */ {
		for i := base + 1; i < end; i++ {
			if (info[i].Mask & indicPlan.maskArray[indicPref]) != 0 {
				/*       1. Only reorder a glyph produced by substitution during application
				 *          of the <pref> feature. (Note that a font may shape a Ra consonant with
				 *          the feature generally but block it in certain contexts.)
				 */
				/* Note: We just check that something got substituted.  We don't check that
				 * the <pref> feature actually did it...
				 *
				 * Reorder pref only if it ligated. */
				if info[i].ligatedAndDidntMultiply() {
					/*
					*       2. Try to find a target position the same way as for pre-base matra.
					*          If it is found, reorder pre-base consonant glyph.
					*
					*       3. If position is not found, reorder immediately before main
					*          consonant.
					 */

					newPos := base
					/* Malayalam / Tamil do not have "half" forms or explicit virama forms.
					* The glyphs formed by 'half' are Chillus or ligated explicit viramas.
					* We want to position matra after them.
					 */
					if buffer.Props.Script != language.Malayalam && buffer.Props.Script != language.Tamil {
						for newPos > start && !isOneOf(&info[newPos-1], 1<<indSM_ex_M|1<<indSM_ex_MPst|1<<indSM_ex_H) {
							newPos--
						}
					}

					if newPos > start && isHalant(&info[newPos-1]) {
						/* . If ZWJ or ZWNJ follow this halant, position is moved after it. */
						if newPos < end && isJoiner(&info[newPos]) {
							newPos++
						}
					}

					{

						oldPos := i
						buffer.mergeClusters(newPos, oldPos+1)

						if debugMode {
							fmt.Printf("INDIC - pre-base: switching glyph %d to %d (and shifting between)", oldPos, newPos)
						}

						tmp := info[oldPos]
						copy(info[newPos+1:], info[newPos:oldPos])
						info[newPos] = tmp

						if newPos <= base && base < oldPos {
							base++
						}
					}
				}

				break
			}
		}
	}

	/* Apply 'init' to the Left Matra if it's a word start. */
	if info[start].complexAux == posPreM {
		const flagRange = 1<<ucd.Unassigned |
			1<<ucd.Co |
			1<<ucd.Cs |
			1<<ucd.Ll |
			1<<ucd.Lm |
			1<<ucd.Lo |
			1<<ucd.Lt |
			1<<ucd.Lu |
			1<<ucd.Mc |
			1<<ucd.Me |
			1<<ucd.Mn
		if start == 0 || 1<<info[start-1].unicode.generalCategory()&flagRange == 0 {
			info[start].Mask |= indicPlan.maskArray[indicInit]
		} else {
			buffer.unsafeToBreak(start-1, start+1)
		}
	}
}

func (indicPlan *indicShapePlan) finalReorderingIndic(plan *otShapePlan, font *Font, buffer *Buffer) bool {
	if len(buffer.Info) == 0 {
		return false
	}

	if debugMode {
		fmt.Println("INDIC - start reordering indic final")
	}

	iter, count := buffer.syllableIterator()
	for start, end := iter.next(); start < count; start, end = iter.next() {
		indicPlan.finalReorderingSyllableIndic(buffer, start, end)
	}

	if debugMode {
		fmt.Println("INDIC - end reordering indic final")
	}

	return false
}

func (ci complexShaperIndic) preprocessText(_ *otShapePlan, buffer *Buffer, _ *Font) {
	preprocessTextVowelConstraints(buffer)
}

func (cs *complexShaperIndic) decompose(c *otNormalizeContext, ab rune) (rune, rune, bool) {
	switch ab {
	/* Don't decompose these. */
	case 0x0931, /* DEVANAGARI LETTER RRA */
		// https://github.com/harfbuzz/harfbuzz/issues/779
		0x09DC, /* BENGALI LETTER RRA */
		0x09DD, /* BENGALI LETTER RHA */
		0x0B94:
		return 0, 0, false /* TAMIL LETTER AU */

		/*
		 * Decompose split matras that don't have Unicode decompositions.
		 */
	}

	return ucd.Decompose(ab)
}

func (cs *complexShaperIndic) compose(c *otNormalizeContext, a, b rune) (rune, bool) {
	/* Avoid recomposing split matras. */
	if ucd.LookupGeneralCategory(a).IsMark() {
		return 0, false
	}

	/* Composition-exclusion exceptions that we want to recompose. */
	if a == 0x09AF && b == 0x09BC {
		return 0x09DF, true
	}

	return ucd.Compose(a, b)
}

func (complexShaperIndic) marksBehavior() (zeroWidthMarks, bool) {
	return zeroWidthMarksNone, false
}

func (complexShaperIndic) normalizationPreference() normalizationMode {
	return nmComposedDiacriticsNoShortCircuit
}
