package harfbuzz

import (
	"github.com/OpenMellon/typesetting/font/opentype/tables"
)

func (c *otApplyContext) applyGPOSMarkToBase(data tables.MarkBasePos, markIndex int) bool {
	buffer := c.buffer

	// Now we search backwards for a non-mark glyph.
	// We don't use skippy_iter.prev() to avoid O(n^2) behavior.

	skippyIter := &c.iterInput
	skippyIter.matcher.lookupProps = uint32(otIgnoreMarks)

	if c.lastBaseUntil > buffer.idx {
		c.lastBaseUntil = 0
		c.lastBase = -1
	}

	for j := buffer.idx; j > c.lastBaseUntil; j-- {
		ma := skippyIter.match(&buffer.Info[j-1])
		if ma == match {
			// https://github.com/harfbuzz/harfbuzz/issues/4124

			// We only want to attach to the first of a MultipleSubst sequence.
			// https://github.com/harfbuzz/harfbuzz/issues/740
			// Reject others...
			// ...but stop if we find a mark in the MultipleSubst sequence:
			// https://github.com/harfbuzz/harfbuzz/issues/1020
			idx := j - 1
			accept := !buffer.Info[idx].multiplied() || buffer.Info[idx].getLigComp() == 0 ||
				idx == 0 || buffer.Info[idx-1].isMark() ||
				buffer.Info[idx].getLigID() != buffer.Info[idx-1].getLigID() ||
				buffer.Info[idx].getLigComp() != buffer.Info[idx-1].getLigComp()+1

			_, covered := data.BaseCoverage.Index(gID(buffer.Info[idx].Glyph))
			if !accept && !covered {
				ma = skip
			}
		}
		if ma == match {
			c.lastBase = j - 1
			break
		}
	}

	c.lastBaseUntil = buffer.idx
	if c.lastBase == -1 {
		buffer.unsafeToConcatFromOutbuffer(0, buffer.idx+1)
		return false
	}

	idx := c.lastBase
	baseIndex, ok := data.BaseCoverage.Index(gID(buffer.Info[idx].Glyph))
	if !ok {
		buffer.unsafeToConcatFromOutbuffer(idx, buffer.idx+1)
		return false
	}

	return c.applyGPOSMarks(data.MarkArray, markIndex, baseIndex, data.BaseArray.Anchors(), idx)
}

func (c *otApplyContext) applyGPOSMarkToLigature(data tables.MarkLigPos, markIndex int) bool {
	buffer := c.buffer

	// now we search backwards for a non-mark glyph
	skippyIter := &c.iterInput
	skippyIter.matcher.lookupProps = uint32(otIgnoreMarks)
	if c.lastBaseUntil > buffer.idx {
		c.lastBaseUntil = 0
		c.lastBase = -1
	}

	for j := buffer.idx; j > c.lastBaseUntil; j-- {
		ma := skippyIter.match(&buffer.Info[j-1])
		if ma == match {
			c.lastBase = j - 1
			break
		}
	}
	c.lastBaseUntil = buffer.idx
	if c.lastBase == -1 {
		c.buffer.unsafeToConcatFromOutbuffer(0, buffer.idx+1)
		return false
	}

	idx := c.lastBase
	ligIndex, ok := data.LigatureCoverage.Index(gID(buffer.Info[idx].Glyph))
	if !ok {
		c.buffer.unsafeToConcatFromOutbuffer(idx, c.buffer.idx+1)
		return false
	}

	ligAttach := data.LigatureArray.LigatureAttachs[ligIndex].Anchors()

	// Find component to attach to
	compCount := ligAttach.Len()
	if compCount == 0 {
		return false
	}

	// We must now check whether the ligature ID of the current mark glyph
	// is identical to the ligature ID of the found ligature.  If yes, we
	// can directly use the component index.  If not, we attach the mark
	// glyph to the last component of the ligature.
	ligID := buffer.Info[idx].getLigID()
	markID := buffer.cur(0).getLigID()
	markComp := buffer.cur(0).getLigComp()
	compIndex := compCount - 1
	if ligID != 0 && ligID == markID && markComp > 0 {
		compIndex = min(compCount, int(buffer.cur(0).getLigComp())) - 1
	}

	return c.applyGPOSMarks(data.MarkArray, markIndex, compIndex, ligAttach, idx)
}

func (c *otApplyContext) applyGPOSMarkToMark(data tables.MarkMarkPos, mark1Index int) bool {
	buffer := c.buffer

	// now we search backwards for a suitable mark glyph until a non-mark glyph
	skippyIter := &c.iterInput
	skippyIter.resetFast(buffer.idx)
	skippyIter.matcher.lookupProps = c.lookupProps &^ uint32(ignoreFlags)
	if ok, _ := skippyIter.prev(); !ok {
		return false
	}

	if !buffer.Info[skippyIter.idx].isMark() {
		return false
	}

	j := skippyIter.idx

	id1 := buffer.cur(0).getLigID()
	id2 := buffer.Info[j].getLigID()
	comp1 := buffer.cur(0).getLigComp()
	comp2 := buffer.Info[j].getLigComp()

	if id1 == id2 {
		if id1 == 0 { /* Marks belonging to the same base. */
			goto good
		} else if comp1 == comp2 { /* Marks belonging to the same ligature component. */
			goto good
		}
	} else {
		/* If ligature ids don't match, it may be the case that one of the marks
		* itself is a ligature.  In which case match. */
		if (id1 > 0 && comp1 == 0) || (id2 > 0 && comp2 == 0) {
			goto good
		}
	}

	/* Didn't match. */
	return false

good:
	mark2Index, ok := data.Mark2Coverage.Index(gID(buffer.Info[j].Glyph))
	if !ok {
		return false
	}

	return c.applyGPOSMarks(data.Mark1Array, mark1Index, mark2Index, data.Mark2Array.Anchors(), j)
}
