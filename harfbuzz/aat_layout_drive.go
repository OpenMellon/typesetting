package harfbuzz

import (
	"fmt"

	"github.com/OpenMellon/typesetting/font"
	"github.com/OpenMellon/typesetting/font/opentype/tables"
)

func (s stateTableDriver) drive(c driverContext, ac *aatApplyContext) {
	const (
		stateStartOfText = uint16(0)
		classEndOfText   = uint16(0)
		DontAdvance      = 0x4000
	)
	buffer := ac.buffer
	if !c.inPlace() {
		buffer.clearOutput()
	}

	state := stateStartOfText
	// If there's only one range, we already checked the flag.
	var lastRange int = -1 // index in ac.rangeFlags, or -1
	if len(ac.rangeFlags) > 1 {
		lastRange = 0
	}
	startStateSafeToBreakEot := !c.isActionable(s.machine.GetEntry(stateStartOfText, classEndOfText))

	for buffer.idx = 0; ; {
		class := classEndOfText
		if buffer.idx < len(buffer.Info) {
			class = s.getClass(buffer.Info[buffer.idx].Glyph, &ac.machineClassCache)
		}
	resume:
		if debugMode {
			fmt.Printf("\t\tState machine - class %d at index %d\n", class, buffer.idx)
		}
		entry := s.machine.GetEntry(state, class)
		nextState := entry.NewState // we only supported extended table

		isNotEpsilonTransition := entry.Flags&DontAdvance == 0
		isNotActionable := !c.isActionable(entry)

		if lastRange != -1 {
			// This block is copied in NoncontextualSubtable::apply. Keep in sync.
			range_ := lastRange
			if buffer.idx < len(buffer.Info) {
				cluster := buffer.cur(0).Cluster
				for cluster < ac.rangeFlags[range_].clusterFirst {
					range_--
				}
				for cluster > ac.rangeFlags[range_].clusterLast {
					range_++
				}

				lastRange = range_
			}
			if !(ac.rangeFlags[range_].flags&ac.subtableFlags != 0) {
				if buffer.idx == len(buffer.Info) {
					break
				}

				state = stateStartOfText
				buffer.nextGlyph()
				continue
			}
		} else {
			// Fast path for when transitioning from start-state to start-state with
			// no action and advancing. Do so as long as the class remains the same.
			// This is common with runs of non-actionable glyphs.

			isNullTransition := state == stateStartOfText &&
				nextState == stateStartOfText &&
				startStateSafeToBreakEot &&
				isNotActionable &&
				isNotEpsilonTransition &&
				lastRange == -1

			if isNullTransition {
				oldKlass := class
				for class == oldKlass {
					c.transition(buffer, s, entry)

					if buffer.idx == len(buffer.Info) {
						break
					}

					buffer.nextGlyph()

					class = classEndOfText
					if buffer.idx < len(buffer.Info) {
						class = s.getClass(buffer.Info[buffer.idx].Glyph, &ac.machineClassCache)
					}
				}

				if buffer.idx == len(buffer.Info) {
					break
				}

				goto resume
			}
		}

		/* Conditions under which it's guaranteed safe-to-break before current glyph:
		 *
		 * 1. There was no action in this transition; and
		 *
		 * 2. If we break before current glyph, the results will be the same. That
		 *    is guaranteed if:
		 *
		 *    2a. We were already in start-of-text state; or
		 *
		 *    2b. We are epsilon-transitioning to start-of-text state; or
		 *
		 *    2c. Starting from start-of-text state seeing current glyph:
		 *
		 *        2c'. There won't be any actions; and
		 *
		 *        2c". We would end up in the same state that we were going to end up
		 *             in now, including whether epsilon-transitioning.
		 *
		 *    and
		 *
		 * 3. If we break before current glyph, there won't be any end-of-text action
		 *    after previous glyph.
		 *
		 * This triples the transitions we need to look up, but is worth returning
		 * granular unsafe-to-break results. See eg.:
		 *
		 *   https://github.com/harfbuzz/harfbuzz/issues/2860
		 */

		wouldbeEntry := s.machine.GetEntry(stateStartOfText, class)
		isSafeToBreak := /* 1. */ !c.isActionable(entry) &&
			/* 2. */
			(
			/* 2a. */
			state == stateStartOfText ||
				/* 2b. */
				((entry.Flags&DontAdvance != 0) && nextState == stateStartOfText) ||
				/* 2c. */
				(
				/* 2c'. */
				!c.isActionable(wouldbeEntry) &&
					/* 2c". */
					(nextState == wouldbeEntry.NewState) &&
					(entry.Flags&DontAdvance) == (wouldbeEntry.Flags&DontAdvance))) &&
			/* 3. */
			!c.isActionable(s.machine.GetEntry(state, classEndOfText))

		if !isSafeToBreak && buffer.backtrackLen() != 0 && buffer.idx < len(buffer.Info) {
			buffer.unsafeToBreakFromOutbuffer(buffer.backtrackLen()-1, buffer.idx+1)
		}

		c.transition(buffer, s, entry)

		state = nextState

		if debugMode {
			fmt.Printf("\t\tState machine - new state %d\n", state)
		}

		if buffer.idx == len(buffer.Info) {
			break
		}

		if isNotEpsilonTransition {
			buffer.nextGlyph()
		} else {
			if buffer.maxOps <= 0 {
				buffer.maxOps--
				buffer.nextGlyph()
			} else {
				buffer.maxOps--
			}
		}
	}

	if !c.inPlace() {
		buffer.swapBuffers()
	}
}

///////

type aatFeatureMapping struct {
	otFeatureTag      font.Tag
	aatFeatureType    aatLayoutFeatureType
	selectorToEnable  aatLayoutFeatureSelector
	selectorToDisable aatLayoutFeatureSelector
}

// FaatLayoutFindFeatureMapping fetches the AAT feature-and-selector combination that corresponds
// to a given OpenType feature tag, or `nil` if not found.
func aatLayoutFindFeatureMapping(tag font.Tag) *aatFeatureMapping {
	low, high := 0, len(featureMappings)
	for low < high {
		mid := low + (high-low)/2 // avoid overflow when computing mid
		p := featureMappings[mid].otFeatureTag
		if tag < p {
			high = mid
		} else if tag > p {
			low = mid + 1
		} else {
			return &featureMappings[mid]
		}
	}
	return nil
}

func (sp *otShapePlan) aatLayoutSubstitute(font *Font, buffer *Buffer, features []Feature) {
	var map_ aatMap
	if len(features) != 0 {
		builder := newAatMapBuilder(font.face.Font, sp.props)
		for _, feature := range features {
			builder.addFeature(feature)
		}
		builder.compile(&map_)
	} else {
		map_ = sp.aatMap
	}

	morx := font.face.Morx
	c := newAatApplyContext(sp, font, buffer)
	c.buffer.unsafeToConcat(0, maxInt)
	c.setupBufferGlyphSet()
	for i, chain := range morx {
		accel := font.morxAccels[i]
		c.rangeFlags = map_.chainFlags[i]
		c.applyMorx(chain, accel)
	}
	// NOTE: we dont support obsolete 'mort' table
}

const bsfAatHasDeleted = bsfShaper0

func aatLayoutRemoveDeletedGlyphs(buffer *Buffer) {
	if buffer.scratchFlags&bsfAatHasDeleted != 0 {
		buffer.deleteGlyphsInplace(func(info *GlyphInfo) bool { return info.isAatDeleted() })
	}
}

func (sp *otShapePlan) aatLayoutPosition(font *Font, buffer *Buffer) {
	kerx := font.face.Kerx

	c := newAatApplyContext(sp, font, buffer)
	c.ankrTable = font.face.Ankr
	c.setupBufferGlyphSet()
	c.applyKernx(kerx, font.kerxAccels)
}

func collectLookupGlyphs(dst *intSet, src tables.AatLookupMixed) {
	for _, r := range src.Coverage() {
		dst.addRange(uint32(r[0]), uint32(r[1]))
	}
}

func (sp *otShapePlan) aatLayoutTrack(font *Font, buffer *Buffer) {
	c := newAatApplyContext(sp, font, buffer)
	c.applyTrak(0)
}
