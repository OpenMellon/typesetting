package harfbuzz

import (
	"fmt"

	"github.com/OpenMellon/typesetting/font"
	"github.com/OpenMellon/typesetting/font/opentype/tables"
)

// port of hb_ot_substitute_post
func (c *otContext) substituteAfterPosition() {
	if c.plan.applyMorx && !c.plan.applyGpos {
		aatLayoutRemoveDeletedGlyphs(c.buffer)
	}

	dealWithVariationSelectors(c.buffer)
	hideDefaultIgnorables(c.buffer, c.font)

	if debugMode {
		fmt.Println("POSTPROCESS glyphs start")
	}
	c.plan.shaper.postprocessGlyphs(c.plan, c.buffer, c.font)
	if debugMode {
		fmt.Println("POSTPROCESS glyphs end ")
	}
}

/*
 * Position
 */

func zeroMarkWidthsByGdef(buffer *Buffer, adjustOffsets bool) {
	for i, inf := range buffer.Info {
		if inf.isMark() {
			pos := &buffer.Pos[i]
			if adjustOffsets { // adjustMarkOffsets
				pos.XOffset -= pos.XAdvance
				pos.YOffset -= pos.YAdvance
			}
			// zeroMarkWidth
			pos.XAdvance = 0
			pos.YAdvance = 0
		}
	}
}

// override Pos array with default values
func (c *otContext) positionDefault() {
	direction := c.buffer.Props.Direction
	info := c.buffer.Info
	pos := c.buffer.Pos

	if direction.isHorizontal() {
		for i, inf := range info {
			pos[i].XAdvance, pos[i].YAdvance = c.font.GlyphHAdvance(inf.Glyph), 0
			pos[i].XOffset, pos[i].YOffset = c.font.subtractGlyphHOrigin(inf.Glyph, 0, 0)
		}
	} else {
		for i, inf := range info {
			pos[i].XAdvance, pos[i].YAdvance = 0, c.font.getGlyphVAdvance(inf.Glyph)
			// v_origin defaults to non-zero; apply even if only fallback is there.
			pos[i].XOffset, pos[i].YOffset = c.font.subtractGlyphVOrigin(inf.Glyph, 0, 0)
		}
	}
	if c.buffer.scratchFlags&bsfHasSpaceFallback != 0 {
		fallbackSpaces(c.font, c.buffer)
	}
}

// hb_ot_position_plan
func (c *otContext) positionComplex() {
	info := c.buffer.Info
	pos := c.buffer.Pos

	/* If the font has no GPOS and direction is forward, then when
	* zeroing mark widths, we shift the mark with it, such that the
	* mark is positioned hanging over the previous glyph.  When
	* direction is backward we don't shift and it will end up
	* hanging over the next glyph after the final reordering.
	*
	* Note: If fallback positioning happens, we don't care about
	* this as it will be overridden. */
	adjustOffsetsWhenZeroing := c.plan.adjustMarkPositioningWhenZeroing && c.buffer.Props.Direction.isForward()

	// we change glyph origin to what GPOS expects (horizontal), apply GPOS, change it back.

	for i, inf := range info {
		pos[i].XOffset, pos[i].YOffset = c.font.addGlyphHOrigin(inf.Glyph, pos[i].XOffset, pos[i].YOffset)
	}

	otLayoutPositionStart(c.font, c.buffer)
	markBehavior, _ := c.plan.shaper.marksBehavior()

	if c.plan.zeroMarks {
		if markBehavior == zeroWidthMarksByGdefEarly {
			zeroMarkWidthsByGdef(c.buffer, adjustOffsetsWhenZeroing)
		}
	}

	c.plan.position(c.font, c.buffer) // apply GPOS, AAT

	if c.plan.zeroMarks {
		if markBehavior == zeroWidthMarksByGdefLate {
			zeroMarkWidthsByGdef(c.buffer, adjustOffsetsWhenZeroing)
		}
	}

	// finish off. Has to follow a certain order.
	zeroWidthDefaultIgnorables(c.buffer)
	otLayoutPositionFinishOffsets(c.font, c.buffer)

	for i, inf := range info {
		pos[i].XOffset, pos[i].YOffset = c.font.subtractGlyphHOrigin(inf.Glyph, pos[i].XOffset, pos[i].YOffset)
	}

	if c.plan.fallbackMarkPositioning {
		fallbackMarkPosition(c.plan, c.font, c.buffer, adjustOffsetsWhenZeroing)
	}
}

func (c *otContext) position() {
	c.buffer.clearPositions()

	c.positionDefault()

	if debugMode {
		fmt.Println("AFTER DEFAULT POSITION", c.buffer.Pos)
	}

	c.positionComplex()

	if c.buffer.Props.Direction.isBackward() {
		c.buffer.Reverse()
	}
}

/* Propagate cluster-level glyph flags to be the same on all cluster glyphs.
 * Simplifies using them. */
func propagateFlags(buffer *Buffer) {
	andMask := glyphFlagDefined
	if buffer.Flags&ProduceUnsafeToConcat == 0 {
		andMask &= ^GlyphUnsafeToConcat
	}

	info := buffer.Info

	if buffer.Flags&ProduceSafeToInsertTatweel == 0 {
		iter, count := buffer.clusterIterator()
		for start, end := iter.next(); start < count; start, end = iter.next() {
			if end-start == 1 {
				info[start].Mask &= andMask
				continue
			}

			var mask GlyphMask
			for i := start; i < end; i++ {
				mask |= info[i].Mask
			}

			mask &= andMask

			for i := start; i < end; i++ {
				info[i].Mask = mask
			}
		}
		return
	}

	// If we are producing SAFE_TO_INSERT_TATWEEL, then do two things:
	//
	// - If the places that the Arabic shaper marked as SAFE_TO_INSERT_TATWEEL,
	//   are UNSAFE_TO_BREAK, then clear the SAFE_TO_INSERT_TATWEEL,
	// - Any place that is SAFE_TO_INSERT_TATWEEL, is also now UNSAFE_TO_BREAK.
	//
	// We couldn't make this interaction earlier. It has to be done this way.

	iter, count := buffer.clusterIterator()
	for start, end := iter.next(); start < count; start, end = iter.next() {
		var mask GlyphMask
		for i := start; i < end; i++ {
			mask |= info[i].Mask
		}

		if mask&GlyphUnsafeToBreak != 0 {
			mask &= ^GlyphSafeToInsertTatweel
		}
		if mask&GlyphSafeToInsertTatweel != 0 {
			mask |= GlyphUnsafeToBreak | GlyphUnsafeToConcat
		}

		mask &= andMask

		for i := start; i < end; i++ {
			info[i].Mask = mask
		}
	}
}

// shaperOpentype is the main shaper of this library.
// It handles complex language and Opentype layout features found in fonts.
type shaperOpentype struct {
	tables *font.Font
	plan   otShapePlan
	key    otShapePlanKey
}

type otShapePlanKey = [2]int // -1 for not found

func (sp *shaperOpentype) init(tables *font.Font, coords []tables.Coord) {
	sp.plan = otShapePlan{}
	sp.key = otShapePlanKey{
		0: tables.GSUB.FindVariationIndex(coords),
		1: tables.GPOS.FindVariationIndex(coords),
	}
	sp.tables = tables
}

func (sp *shaperOpentype) compile(props SegmentProperties, userFeatures []Feature) {
	sp.plan.init0(sp.tables, props, userFeatures, sp.key)
}

// pull it all together!
func (sp *shaperOpentype) shape(font *Font, buffer *Buffer, features []Feature) {
	c := otContext{plan: &sp.plan, font: font, buffer: buffer, userFeatures: features}
	c.buffer.scratchFlags = bsfDefault

	const maxLenFactor = 64
	const maxLenMin = 16384
	const maxOpsFactor = 1024
	const maxOpsMin = 16384
	c.buffer.maxOps = max(len(c.buffer.Info)*maxOpsFactor, maxOpsMin)
	c.buffer.maxLen = max(len(c.buffer.Info)*maxLenFactor, maxLenMin)

	// save the original direction, we use it later.
	c.targetDirection = c.buffer.Props.Direction

	c.initializeMasks()
	c.buffer.setUnicodeProps()
	c.buffer.insertDottedCircle(c.font)

	c.buffer.formClusters()

	if debugMode {
		fmt.Println("FORMING CLUSTER :", c.buffer.Info)
	}

	c.buffer.ensureNativeDirection()

	if debugMode {
		fmt.Printf("PREPROCESS text start\n")
	}
	c.plan.shaper.preprocessText(c.plan, c.buffer, c.font)
	if debugMode {
		fmt.Println("PREPROCESS text end:", c.buffer.Info)
	}

	c.substituteBeforePosition() // apply GSUB

	if debugMode {
		fmt.Println("AFTER SUBSTITUTE", c.buffer.Info)
	}

	c.position()

	if debugMode {
		fmt.Println("AFTER POSITION", c.buffer.Pos)
	}

	c.substituteAfterPosition()

	propagateFlags(c.buffer)

	c.buffer.Props.Direction = c.targetDirection

	c.buffer.maxOps = maxOpsDefault
}
