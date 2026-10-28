package harfbuzz

import (
	"fmt"

	"github.com/OpenMellon/typesetting/font/opentype/tables"
	ucd "github.com/OpenMellon/typesetting/internal/unicodedata"
)

// `input` starts with second glyph (`inputCount` = len(input)+1)
// only the input lookupsContext is needed
func (c *wouldApplyContext) wouldApplyChainLookup(backtrack, input, lookahead []uint16, inputLookupContext matcherFunc) bool {
	contextOk := true
	if c.zeroContext {
		contextOk = len(backtrack) == 0 && len(lookahead) == 0
	}
	return contextOk && c.wouldMatchInput(input, inputLookupContext)
}

// `input` starts with second glyph (`count` = len(input)+1)
func (c *wouldApplyContext) wouldMatchInput(input []uint16, matchFunc matcherFunc) bool {
	if len(c.glyphs) != len(input)+1 {
		return false
	}

	for i, glyph := range input {
		if !matchFunc(gID(c.glyphs[i+1]), glyph) {
			return false
		}
	}

	return true
}

// `input` starts with second glyph (`inputCount` = len(input)+1)
func (c *otApplyContext) matchInput(input []uint16, matchFunc matcherFunc) (_ bool, endPosition int, totalComponentCount uint8) {
	buffer := c.buffer
	count := len(input) + 1
	if count == 1 {
		endPosition = buffer.idx + 1
		c.matchPositions[0] = buffer.idx
		totalComponentCount = buffer.cur(0).getLigNumComps()
		return true, endPosition, totalComponentCount
	}

	if count > maxContextLength {
		return false, 0, 0
	}
	skippyIter := &c.iterInput
	skippyIter.reset(buffer.idx)
	skippyIter.setMatchFunc(matchFunc, input)

	/*
	* This is perhaps the trickiest part of OpenType...  Remarks:
	*
	* - If all components of the ligature were marks, we call this a mark ligature.
	*
	* - If there is no GDEF, and the ligature is NOT a mark ligature, we categorize
	*   it as a ligature glyph.
	*
	* - Ligatures cannot be formed across glyphs attached to different components
	*   of previous ligatures.  Eg. the sequence is LAM,SHADDA,LAM,FATHA,HEH, and
	*   LAM,LAM,HEH form a ligature, leaving SHADDA,FATHA next to eachother.
	*   However, it would be wrong to ligate that SHADDA,FATHA sequence.
	*   There are a couple of exceptions to this:
	*
	*   o If a ligature tries ligating with marks that belong to it itself, go ahead,
	*     assuming that the font designer knows what they are doing (otherwise it can
	*     break Indic stuff when a matra wants to ligate with a conjunct,
	*
	*   o If two marks want to ligate and they belong to different components of the
	*     same ligature glyph, and said ligature glyph is to be ignored according to
	*     mark-filtering rules, then allow.
	*     https://github.com/harfbuzz/harfbuzz/issues/545
	 */

	firstLigID := buffer.cur(0).getLigID()
	firstLigComp := buffer.cur(0).getLigComp()

	const (
		ligbaseNotChecked = iota
		ligbaseMayNotSkip
		ligbaseMaySkip
	)
	ligbase := ligbaseNotChecked
	for i := 1; i < count; i++ {
		if ok, unsafeTo := skippyIter.next(); !ok {
			return false, unsafeTo, 0
		}

		ensureLen(&c.matchPositions, i+1)

		c.matchPositions[i] = skippyIter.idx

		thisLigID := buffer.Info[skippyIter.idx].getLigID()
		thisLigComp := buffer.Info[skippyIter.idx].getLigComp()
		if firstLigID != 0 && firstLigComp != 0 {
			/* If first component was attached to a previous ligature component,
			* all subsequent components should be attached to the same ligature
			* component, otherwise we shouldn't ligate them... */
			if firstLigID != thisLigID || firstLigComp != thisLigComp {
				/* ...unless, we are attached to a base ligature and that base
				 * ligature is ignorable. */
				if ligbase == ligbaseNotChecked {
					found := false
					out := buffer.outInfo
					j := len(out)
					for j != 0 && out[j-1].getLigID() == firstLigID {
						if out[j-1].getLigComp() == 0 {
							j--
							found = true
							break
						}
						j--
					}

					if found && skippyIter.maySkip(&out[j]) == yes {
						ligbase = ligbaseMaySkip
					} else {
						ligbase = ligbaseMayNotSkip
					}
				}

				if ligbase == ligbaseMayNotSkip {
					return false, 0, 0
				}
			}
		} else {
			/* If first component was NOT attached to a previous ligature component,
			* all subsequent components should also NOT be attached to any ligature
			* component, unless they are attached to the first component itself! */
			if thisLigID != 0 && thisLigComp != 0 && (thisLigID != firstLigID) {
				return false, 0, 0
			}
		}

		totalComponentCount += buffer.Info[skippyIter.idx].getLigNumComps()
	}

	endPosition = skippyIter.idx + 1
	totalComponentCount += buffer.cur(0).getLigNumComps()
	c.matchPositions[0] = buffer.idx

	return true, endPosition, totalComponentCount
}

// `count` and `c.matchPositions` include the first glyph
func (c *otApplyContext) ligateInput(count, matchEnd int, ligGlyph gID, totalComponentCount uint8) {
	buffer := c.buffer

	buffer.mergeClusters(buffer.idx, matchEnd)

	/* - If a base and one or more marks ligate, consider that as a base, NOT
	*   ligature, such that all following marks can still attach to it.
	*   https://github.com/harfbuzz/harfbuzz/issues/1109
	*
	* - If all components of the ligature were marks, we call this a mark ligature.
	*   If it *is* a mark ligature, we don't allocate a new ligature id, and leave
	*   the ligature to keep its old ligature id.  This will allow it to attach to
	*   a base ligature in GPOS.  Eg. if the sequence is: LAM,LAM,SHADDA,FATHA,HEH,
	*   and LAM,LAM,HEH for a ligature, they will leave SHADDA and FATHA with a
	*   ligature id and component value of 2.  Then if SHADDA,FATHA form a ligature
	*   later, we don't want them to lose their ligature id/component, otherwise
	*   GPOS will fail to correctly position the mark ligature on top of the
	*   LAM,LAM,HEH ligature.  See:
	*     https://bugzilla.gnome.org/show_bug.cgi?id=676343
	*
	* - If a ligature is formed of components that some of which are also ligatures
	*   themselves, and those ligature components had marks attached to *their*
	*   components, we have to attach the marks to the new ligature component
	*   positions!  Now *that*'s tricky!  And these marks may be following the
	*   last component of the whole sequence, so we should loop forward looking
	*   for them and update them.
	*
	*   Eg. the sequence is LAM,LAM,SHADDA,FATHA,HEH, and the font first forms a
	*   'calt' ligature of LAM,HEH, leaving the SHADDA and FATHA with a ligature
	*   id and component == 1.  Now, during 'liga', the LAM and the LAM-HEH ligature
	*   form a LAM-LAM-HEH ligature.  We need to reassign the SHADDA and FATHA to
	*   the new ligature with a component value of 2.
	*
	*   This in fact happened to a font...  See:
	*   https://bugzilla.gnome.org/show_bug.cgi?id=437633
	 */

	isBaseLigature := buffer.Info[c.matchPositions[0]].isBaseGlyph()
	isMarkLigature := buffer.Info[c.matchPositions[0]].isMark()
	for i := 1; i < count; i++ {
		if !buffer.Info[c.matchPositions[i]].isMark() {
			isBaseLigature = false
			isMarkLigature = false
			break
		}
	}
	isLigature := !isBaseLigature && !isMarkLigature

	klass, ligID := uint16(0), uint8(0)
	if isLigature {
		klass = tables.GPLigature
		ligID = buffer.allocateLigID()
	}
	lastLigID := buffer.cur(0).getLigID()
	lastNumComponents := buffer.cur(0).getLigNumComps()
	componentsSoFar := lastNumComponents

	if isLigature {
		buffer.cur(0).setLigPropsForLigature(ligID, totalComponentCount)
		if buffer.cur(0).unicode.generalCategory() == ucd.Mn {
			buffer.cur(0).setGeneralCategory(ucd.Lo)
		}
	}

	// ReplaceGlyph_with_ligature
	c.setGlyphClassExt(GID(ligGlyph), klass, true, false)
	buffer.replaceGlyphIndex(GID(ligGlyph))

	for i := 1; i < count; i++ {
		for buffer.idx < c.matchPositions[i] {
			if isLigature {
				thisComp := buffer.cur(0).getLigComp()
				if thisComp == 0 {
					thisComp = lastNumComponents
				}
				newLigComp := componentsSoFar - lastNumComponents +
					min8(thisComp, lastNumComponents)
				buffer.cur(0).setLigPropsForMark(ligID, newLigComp)
			}
			buffer.nextGlyph()
		}

		lastLigID = buffer.cur(0).getLigID()
		lastNumComponents = buffer.cur(0).getLigNumComps()
		componentsSoFar += lastNumComponents

		/* Skip the base glyph */
		buffer.skipGlyph()
	}

	if !isMarkLigature && lastLigID != 0 {
		/* Re-adjust components for any marks following. */
		for i := buffer.idx; i < len(buffer.Info); i++ {
			if lastLigID != buffer.Info[i].getLigID() {
				break
			}

			thisComp := buffer.Info[i].getLigComp()
			if thisComp == 0 {
				break
			}

			newLigComp := componentsSoFar - lastNumComponents +
				min8(thisComp, lastNumComponents)
			buffer.Info[i].setLigPropsForMark(ligID, newLigComp)
		}
	}
}

func (c *otApplyContext) recurse(subLookupIndex uint16) bool {
	if c.nestingLevelLeft == 0 || c.recurseFunc == nil || c.buffer.maxOps <= 0 {
		if c.buffer.maxOps <= 0 {
			c.buffer.maxOps--
			return false
		}
		c.buffer.maxOps--
	}

	c.nestingLevelLeft--
	ret := c.recurseFunc(c, subLookupIndex)
	c.nestingLevelLeft++
	return ret
}

func (c *otApplyContext) matchBacktrack(backtrack []uint16, matchFunc matcherFunc) (_ bool, matchStart int) {
	if len(backtrack) == 0 {
		return true, c.buffer.backtrackLen()
	}

	skippyIter := &c.iterContext
	skippyIter.resetBack(c.buffer.backtrackLen())
	skippyIter.setMatchFunc(matchFunc, backtrack)

	for i := 0; i < len(backtrack); i++ {
		if ok, unsafeFrom := skippyIter.prev(); !ok {
			return false, unsafeFrom
		}
	}

	return true, skippyIter.idx
}

func (c *otApplyContext) matchLookahead(lookahead []uint16, matchFunc matcherFunc, startIndex int) (_ bool, endIndex int) {
	if len(lookahead) == 0 {
		return true, startIndex
	}

	skippyIter := &c.iterContext
	skippyIter.reset(startIndex - 1)
	skippyIter.setMatchFunc(matchFunc, lookahead)

	for i := 0; i < len(lookahead); i++ {
		if ok, unsafeTo := skippyIter.next(); !ok {
			return false, unsafeTo
		}
	}

	return true, skippyIter.idx + 1
}

// `count` and `matchPositions` include the first glyph
// `lookupRecord` is in design order
func (c *otApplyContext) applyLookup(count int,
	lookupRecord []tables.SequenceLookupRecord, matchLength int,
) {
	buffer := c.buffer
	var end int

	/* All positions are distance from beginning of *output* buffer.
	* Adjust. */
	{
		bl := buffer.backtrackLen()
		end = bl + matchLength - buffer.idx

		delta := bl - buffer.idx
		/* Convert positions to new indexing. */
		for j := 0; j < count; j++ {
			c.matchPositions[j] += delta
		}
	}

	for _, lk := range lookupRecord {
		idx := int(lk.SequenceIndex)
		if idx >= count { // invalid, ignored
			continue
		}

		origLen := buffer.backtrackLen() + buffer.lookaheadLen()

		// This can happen if earlier recursed lookups deleted many entries.
		if c.matchPositions[idx] >= origLen {
			continue
		}

		buffer.moveTo(c.matchPositions[idx])

		if buffer.maxOps <= 0 {
			break
		}

		if debugMode {
			fmt.Printf("\t\tAPPLY nested lookup %d\n", lk.LookupListIndex)
		}

		if !c.recurse(lk.LookupListIndex) {
			continue
		}

		newLen := buffer.backtrackLen() + buffer.lookaheadLen()
		delta := newLen - origLen

		if delta == 0 {
			continue
		}

		// Recursed lookup changed buffer len. Adjust.
		//
		// TODO:
		//
		// Right now, if buffer length increased by n, we assume n new glyphs
		// were added right after the current position, and if buffer length
		// was decreased by n, we assume n match positions after the current
		// one where removed.  The former (buffer length increased) case is
		// fine, but the decrease case can be improved in at least two ways,
		// both of which are significant:
		//
		//   - If recursed-to lookup is MultipleSubst and buffer length
		//     decreased, then it's current match position that was deleted,
		//     NOT the one after it.
		//
		//   - If buffer length was decreased by n, it does not necessarily
		//     mean that n match positions where removed, as there recursed-to
		//     lookup might had a different LookupFlag.  Here's a constructed
		//     case of that:
		//     https://github.com/harfbuzz/harfbuzz/discussions/3538
		//
		// It should be possible to construct tests for both of these cases.
		//
		end += delta
		if end < int(c.matchPositions[idx]) {
			// End might end up being smaller than match_positions[idx] if the recursed
			// lookup ended up removing many items.
			// Just never rewind end beyond start of current position, since that is
			// not possible in the recursed lookup.  Also adjust delta as such.
			//
			// https://bugs.chromium.org/p/chromium/issues/detail?id=659496
			// https://github.com/harfbuzz/harfbuzz/issues/1611
			//
			delta += c.matchPositions[idx] - end
			end = c.matchPositions[idx]
		}

		next := idx + 1 // next now is the position after the recursed lookup.

		if delta > 0 {
			if delta+count > maxContextLength {
				break
			}
			ensureLen(&c.matchPositions, delta+count)
		} else {
			/* NOTE: delta is negative. */
			delta = max(delta, int(next)-int(count))
			next -= delta
		}

		/* Shift! */
		copy(c.matchPositions[next+delta:], c.matchPositions[next:count])
		next += delta
		count += delta

		/* Fill in new entries. */
		for j := idx + 1; j < next; j++ {
			c.matchPositions[j] = c.matchPositions[j-1] + 1
		}

		/* And fixup the rest. */
		for ; next < count; next++ {
			c.matchPositions[next] += delta
		}

	}

	buffer.moveTo(end)
}

func (c *otApplyContext) applyLookupContext1(data tables.SequenceContextFormat1, index int) bool {
	if index >= len(data.SeqRuleSet) { // index is not sanitized in tt.Parse
		return false
	}
	ruleSet := data.SeqRuleSet[index]
	return c.applyRuleSet(ruleSet, matchGlyph)
}

func (c *otApplyContext) applyLookupContext2(data tables.SequenceContextFormat2, _ int, glyphID GID) bool {
	class, _ := data.ClassDef.Class(gID(glyphID))
	var ruleSet tables.SequenceRuleSet
	if int(class) < len(data.ClassSeqRuleSet) {
		ruleSet = data.ClassSeqRuleSet[class]
	}
	return c.applyRuleSet(ruleSet, matchClass(data.ClassDef))
}

// return a slice containing [start, start+1, ..., end-1],
// using `indices` as an internal buffer to avoid allocations
// these indices are used to refer to coverage
func get1N(indices *[]uint16, start, end int) []uint16 {
	if end > cap(*indices) {
		*indices = make([]uint16, end)
		for i := range *indices {
			(*indices)[i] = uint16(i)
		}
	}
	return (*indices)[start:end]
}

func (c *otApplyContext) applyLookupContext3(data tables.SequenceContextFormat3, _ int) bool {
	covIndices := get1N(&c.indices, 1, len(data.Coverages))
	return c.contextApplyLookup(covIndices, data.SeqLookupRecords, matchCoverage(data.Coverages))
}

func (c *otApplyContext) applyLookupChainedContext1(data tables.ChainedSequenceContextFormat1, index int) bool {
	if index >= len(data.ChainedSeqRuleSet) { // index is not sanitized in tt.Parse
		return false
	}
	ruleSet := data.ChainedSeqRuleSet[index]
	return c.applyChainRuleSet(ruleSet, [3]matcherFunc{matchGlyph, matchGlyph, matchGlyph})
}

func (c *otApplyContext) applyLookupChainedContext2(data tables.ChainedSequenceContextFormat2, _ int, glyphID GID) bool {
	class, _ := data.InputClassDef.Class(gID(glyphID))
	var ruleSet tables.ChainedClassSequenceRuleSet
	if int(class) < len(data.ChainedClassSeqRuleSet) {
		ruleSet = data.ChainedClassSeqRuleSet[class]
	}
	return c.applyChainRuleSet(ruleSet, [3]matcherFunc{
		matchClass(data.BacktrackClassDef), matchClass(data.InputClassDef), matchClass(data.LookaheadClassDef),
	})
}

func (c *otApplyContext) applyLookupChainedContext3(data tables.ChainedSequenceContextFormat3, _ int) bool {
	lB, lI, lL := len(data.BacktrackCoverages), len(data.InputCoverages), len(data.LookaheadCoverages)
	return c.chainContextApplyLookup(get1N(&c.indices, 0, lB), get1N(&c.indices, 1, lI), get1N(&c.indices, 0, lL),
		data.SeqLookupRecords, [3]matcherFunc{
			matchCoverage(data.BacktrackCoverages), matchCoverage(data.InputCoverages), matchCoverage(data.LookaheadCoverages),
		})
}

// grows the slice if required; and copy the current contents
func ensureLen(slice *[]int, L int) {
	if L > len(*slice) {
		if L <= cap(*slice) {
			*slice = (*slice)[:L]
		} else {
			newSlice := make([]int, L)
			copy(newSlice, *slice)
			*slice = newSlice
		}
	}
}
