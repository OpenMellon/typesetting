package harfbuzz

import (
	"fmt"
	"math"

	"github.com/OpenMellon/typesetting/font"
	"github.com/OpenMellon/typesetting/font/opentype/tables"
)

// ported from harfbuzz/src/hb-ot-layout-gsubgpos.hh Copyright © 2007,2008,2009,2010  Red Hat, Inc. 2010,2012  Google, Inc.  Behdad Esfahbod

// GSUB or GPOS lookup
type layoutLookup interface {
	// accumulate the subtables coverage into the diggest
	collectCoverage(*setDigest)
	// walk the subtables to add them to the context
	dispatchSubtables(*getSubtablesContext)

	// walk the subtables and apply the sub/pos
	dispatchApply(ctx *otApplyContext) bool

	Props() uint32
	isReverse() bool
}

/*
 * GSUB/GPOS Common
 */

const ignoreFlags = otIgnoreBaseGlyphs | otIgnoreLigatures | otIgnoreMarks

// use a digest to speedup match
type otLayoutLookupAccelerator struct {
	lookup    layoutLookup
	subtables getSubtablesContext
	digest    setDigest
}

func (ac *otLayoutLookupAccelerator) init(lookup layoutLookup) {
	ac.lookup = lookup
	ac.digest = setDigest{}
	lookup.collectCoverage(&ac.digest)
	ac.subtables = nil
	lookup.dispatchSubtables(&ac.subtables)
}

// apply the subtables and stops at the first success.
func (ac *otLayoutLookupAccelerator) apply(c *otApplyContext) bool {
	for _, table := range ac.subtables {
		if table.apply(c) {
			return true
		}
	}
	return false
}

// represents one layout subtable, with its own coverage
type applicable struct {
	objApply func(c *otApplyContext) bool

	digest setDigest
}

func newGSUBApplicable(table tables.GSUBLookup) applicable {
	ap := applicable{objApply: func(c *otApplyContext) bool { return c.applyGSUB(table) }}
	ap.digest.collectCoverage(table.Cov())
	return ap
}

func newGPOSApplicable(table tables.GPOSLookup) applicable {
	ap := applicable{objApply: func(c *otApplyContext) bool { return c.applyGPOS(table) }}
	ap.digest.collectCoverage(table.Cov())
	return ap
}

func (ap applicable) apply(c *otApplyContext) bool {
	return ap.digest.mayHave(gID(c.buffer.cur(0).Glyph)) && ap.objApply(c)
}

type getSubtablesContext []applicable

// one for GSUB, one for GPOS (known at compile time)
type otProxyMeta struct {
	recurseFunc recurseFunc
	tableIndex  uint8 // 0 for GSUB, 1 for GPOS
	inplace     bool
}

var (
	proxyGSUB = otProxyMeta{tableIndex: 0, inplace: false, recurseFunc: applyRecurseGSUB}
	proxyGPOS = otProxyMeta{tableIndex: 1, inplace: true, recurseFunc: applyRecurseGPOS}
)

type otProxy struct {
	otProxyMeta
	accels []otLayoutLookupAccelerator
}

type wouldApplyContext struct {
	glyphs      []GID
	indices     []uint16 // see get1N
	zeroContext bool
}

// `value` interpretation is dictated by the context
type matcherFunc = func(gid gID, value uint16) bool

// interprets `value` as a Glyph
func matchGlyph(gid gID, value uint16) bool { return gid == gID(value) }

// interprets `value` as a Class
func matchClass(class tables.ClassDef) matcherFunc {
	return func(gid gID, value uint16) bool {
		c, _ := class.Class(gid)
		return uint16(c) == value
	}
}

// interprets `value` as an index in coverage array
func matchCoverage(covs []tables.Coverage) matcherFunc {
	return func(gid gID, value uint16) bool {
		_, covered := covs[value].Index(gid)
		return covered
	}
}

const (
	no uint8 = iota
	yes
	maybe
)

type otApplyContextMatcher struct {
	matchFunc    matcherFunc
	lookupProps  uint32
	mask         GlyphMask
	ignoreZWNJ   bool
	ignoreZWJ    bool
	ignoreHidden bool
	perSyllable  bool
	syllable     uint8
}

func (m *otApplyContextMatcher) init(c *otApplyContext, contextMatch bool) {
	m.matchFunc = nil
	m.lookupProps = c.lookupProps
	/* Ignore ZWNJ if we are matching GPOS, or matching GSUB context and asked to. */
	m.ignoreZWNJ = c.tableIndex == 1 || (contextMatch && c.autoZWNJ)
	/* Ignore ZWJ if we are matching context, or asked to. */
	m.ignoreZWJ = contextMatch || c.autoZWJ
	/* Ignore hidden glyphs (like CGJ) during GPOS. */
	m.ignoreHidden = c.tableIndex == 1
	if contextMatch {
		m.mask = math.MaxUint32
	} else {
		m.mask = c.lookupMask
	}
	// Per syllable matching is only for GSUB.
	m.perSyllable = c.tableIndex == 0 && c.perSyllable
	m.syllable = 0
}

func (m otApplyContextMatcher) mayMatch(info *GlyphInfo, glyphData []uint16) uint8 {
	if info.Mask&m.mask == 0 || (m.perSyllable && m.syllable != 0 && m.syllable != info.syllable) {
		return no
	}

	if m.matchFunc != nil {
		if m.matchFunc(gID(info.Glyph), glyphData[0]) {
			return yes
		}
		return no
	}

	return maybe
}

func (m otApplyContextMatcher) maySkip(c *otApplyContext, info *GlyphInfo) uint8 {
	if !c.checkGlyphProperty(info, m.lookupProps) {
		return yes
	}

	if info.isDefaultIgnorable() &&
		(m.ignoreZWNJ || !info.isZwnj()) &&
		(m.ignoreZWJ || !info.isZwj()) &&
		(m.ignoreHidden || !info.isHidden()) {
		return maybe
	}

	return no
}

type skippingIterator struct {
	c       *otApplyContext
	matcher otApplyContextMatcher

	matchGlyphDataArray []uint16
	matchGlyphDataStart int // start as index in matchGlyphDataArray

	idx int
	end int
}

func (it *skippingIterator) init(c *otApplyContext, contextMatch bool) {
	it.c = c
	it.end = len(c.buffer.Info)
	it.matchGlyphDataArray = it.matchGlyphDataArray[:0]
	it.matchGlyphDataStart = 0
	it.matcher.init(c, contextMatch)
}

func (it *skippingIterator) setMatchFunc(matchFunc matcherFunc, glyphData []uint16) {
	it.matcher.matchFunc = matchFunc
	it.matchGlyphDataArray = glyphData
	it.matchGlyphDataStart = 0
}

func (it *skippingIterator) reset(startIndex int) {
	// For GSUB forward iterator
	it.idx = startIndex
	it.end = len(it.c.buffer.Info)
	it.matcher.syllable = it.c.buffer.cur(0).syllable
}

func (it *skippingIterator) resetBack(startIndex int) {
	// For GSUB backward iterator
	it.idx = startIndex
	it.matcher.syllable = it.c.buffer.cur(0).syllable
}

func (it *skippingIterator) resetFast(startIndex int) {
	// Doesn't set end or syllable. Used by GPOS which doesn't care / change.
	it.idx = startIndex
}

func (it *skippingIterator) maySkip(info *GlyphInfo) uint8 { return it.matcher.maySkip(it.c, info) }

type matchRes uint8

const (
	match matchRes = iota
	notMatch
	skip
)

func (it *skippingIterator) match(info *GlyphInfo) matchRes {
	skipR := it.matcher.maySkip(it.c, info)
	if skipR == yes {
		return skip
	}

	matchR := it.matcher.mayMatch(info, it.matchGlyphDataArray[it.matchGlyphDataStart:])
	if matchR == yes || (matchR == maybe && skipR == no) {
		return match
	}

	if skipR == no {
		return notMatch
	}

	return skip
}

func (it *skippingIterator) next() (_ bool, unsafeTo int) {
	stop := it.end - 1
	for it.idx < stop {
		it.idx++
		info := &it.c.buffer.Info[it.idx]
		switch it.match(info) {
		case match:
			if len(it.matchGlyphDataArray) != 0 {
				it.matchGlyphDataStart++
			}
			return true, 0
		case notMatch:
			return false, it.idx + 1
		case skip:
			continue
		}
	}
	return false, it.end
}

func (it *skippingIterator) prev() (_ bool, unsafeFrom int) {
	stop := 0
	L := len(it.c.buffer.outInfo)
	for it.idx > stop {
		it.idx--
		var info *GlyphInfo
		if it.idx < L {
			info = &it.c.buffer.outInfo[it.idx]
		} else {
			// we are in "position mode" : outInfo is not used anymore
			// in the C implementation, outInfo and info now are sharing the same storage
			info = &it.c.buffer.Info[it.idx]
		}

		switch it.match(info) {
		case match:
			if len(it.matchGlyphDataArray) != 0 {
				it.matchGlyphDataStart++
			}
			return true, 0
		case notMatch:
			return false, max(1, it.idx) - 1
		case skip:
			continue
		}
	}
	return false, 0
}

type recurseFunc = func(c *otApplyContext, lookupIndex uint16) bool

type otApplyContext struct {
	font   *Font
	buffer *Buffer

	recurseFunc recurseFunc
	gdef        tables.GDEF
	varStore    tables.ItemVarStore
	indices     []uint16 // see get1N()

	iterContext skippingIterator
	iterInput   skippingIterator

	nestingLevelLeft int
	tableIndex       uint8 // 0 for GSUB, 1 for GPOS
	lookupMask       GlyphMask
	lookupProps      uint32
	randomState      uint32
	lookupIndex      uint16
	direction        Direction

	hasGlyphClasses bool
	autoZWNJ        bool
	autoZWJ         bool
	perSyllable     bool
	newSyllables    uint8 // 0xFF for undefined
	random          bool

	lastBase      int // GPOS uses
	lastBaseUntil int // GPOS uses

	matchPositions []int
}

func (c *otApplyContext) reset(tableIndex uint8, font *Font, buffer *Buffer) {
	c.font = font
	c.buffer = buffer

	c.recurseFunc = nil
	c.gdef = font.face.GDEF
	c.varStore = c.gdef.ItemVarStore
	c.indices = c.indices[:0]

	c.nestingLevelLeft = maxNestingLevel
	c.tableIndex = tableIndex
	c.lookupMask = 1
	c.lookupProps = 0
	c.randomState = 1
	c.lookupIndex = 0
	c.direction = buffer.Props.Direction

	c.hasGlyphClasses = c.gdef.GlyphClassDef != nil
	c.autoZWNJ = true
	c.autoZWJ = true
	c.perSyllable = false
	c.newSyllables = 0xFF
	c.random = false

	c.lastBase = -1
	c.lastBaseUntil = 0

	ensureLen(&c.matchPositions, 8)

	// iterContext
	// iterInput
	c.initIters()
}

func (c *otApplyContext) initIters() {
	c.iterInput.init(c, false)
	c.iterContext.init(c, true)
}

func (c *otApplyContext) setLookupMask(mask GlyphMask) {
	c.lookupMask = mask
	c.initIters()
}

func (c *otApplyContext) setLookupProps(lookupProps uint32) {
	c.lookupProps = lookupProps
	c.initIters()
}

func (c *otApplyContext) applyRecurseLookup(lookupIndex uint16, l layoutLookup) bool {
	savedLookupProps := c.lookupProps
	savedLookupIndex := c.lookupIndex

	c.lookupIndex = lookupIndex
	c.setLookupProps(l.Props())

	savedMatchPositions := c.matchPositions
	var stackMatchPositions [8]int
	c.matchPositions = stackMatchPositions[:]

	ret := l.dispatchApply(c)

	c.lookupIndex = savedLookupIndex
	c.setLookupProps(savedLookupProps)

	c.matchPositions = savedMatchPositions

	return ret
}

func (c *otApplyContext) substituteLookup(accel *otLayoutLookupAccelerator) {
	c.applyString(proxyGSUB, accel)
}

func (c *otApplyContext) checkGlyphProperty(info *GlyphInfo, matchProps uint32) bool {
	glyphProps := info.glyphProps

	/* Not covered, if, for example, glyph class is ligature and
	 * matchProps includes LookupFlags::IgnoreLigatures */
	if (glyphProps & uint16(matchProps) & ignoreFlags) != 0 {
		return false
	}

	if glyphProps&tables.GPMark != 0 {
		return c.matchPropertiesMark(info, glyphProps, matchProps)
	}

	return true
}

func (c *otApplyContext) matchPropertiesMark(info *GlyphInfo, glyphProps uint16, matchProps uint32) bool {
	/* If using mark filtering sets, the high uint16 of
	 * matchProps has the set index. */
	if uint16(matchProps)&font.UseMarkFilteringSet != 0 {
		_, has := c.gdef.MarkGlyphSetsDef.Coverages[matchProps>>16].Index(gID(info.Glyph))
		return has
	}

	/* The second byte of matchProps has the meaning
	 * "ignore marks of attachment type different than
	 * the attachment type specified." */
	if uint16(matchProps)&otMarkAttachmentType != 0 {
		return uint16(matchProps)&otMarkAttachmentType == (glyphProps & otMarkAttachmentType)
	}

	return true
}

func (c *otApplyContext) setGlyphClass(glyphIndex GID) {
	c.setGlyphClassExt(glyphIndex, 0, false, false)
}

func (c *otApplyContext) setGlyphClassExt(glyphIndex_ GID, classGuess uint16, ligature, component bool) {
	glyphIndex := gID(glyphIndex_)

	c.buffer.digest.add(glyphIndex)

	if c.newSyllables != 0xFF {
		c.buffer.cur(0).syllable = c.newSyllables
	}

	props := c.buffer.cur(0).glyphProps | substituted
	if ligature {
		props |= ligated
		// In the only place that the MULTIPLIED bit is used, Uniscribe
		// seems to only care about the "last" transformation between
		// Ligature and Multiple substitutions.  Ie. if you ligate, expand,
		// and ligate again, it forgives the multiplication and acts as
		// if only ligation happened.  As such, clear MULTIPLIED bit.
		props &= ^multiplied
	}
	if component {
		props |= multiplied
	}
	if c.hasGlyphClasses {
		props &= preserve
		c.buffer.cur(0).glyphProps = props | c.gdef.GlyphProps(glyphIndex)
	} else if classGuess != 0 {
		props &= preserve
		c.buffer.cur(0).glyphProps = props | classGuess
	} else {
		c.buffer.cur(0).glyphProps = props
	}
}

func (c *otApplyContext) replaceGlyph(glyphIndex GID) {
	c.setGlyphClass(glyphIndex)
	c.buffer.replaceGlyphIndex(glyphIndex)
}

func (c *otApplyContext) randomNumber() uint32 {
	/* http://www.cplusplus.com/reference/random/minstd_rand/ */
	c.randomState = c.randomState * 48271 % 2147483647
	return c.randomState
}

func (c *otApplyContext) applyRuleSet(ruleSet tables.SequenceRuleSet, match matcherFunc) bool {
	for _, rule := range ruleSet.SeqRule {
		// the first which match is applied
		applied := c.contextApplyLookup(rule.InputSequence, rule.SeqLookupRecords, match)
		if applied {
			return true
		}
	}
	return false
}

func (c *otApplyContext) applyChainRuleSet(ruleSet tables.ChainedClassSequenceRuleSet, match [3]matcherFunc) bool {
	for i, rule := range ruleSet.ChainedSeqRules {

		if debugMode {
			fmt.Println("APPLY - chain rule number", i)
		}

		b := c.chainContextApplyLookup(rule.BacktrackSequence, rule.InputSequence, rule.LookaheadSequence, rule.SeqLookupRecords, match)
		if b { // stop at the first application
			return true
		}
	}
	return false
}

// `input` starts with second glyph (`inputCount` = len(input)+1)
func (c *otApplyContext) contextApplyLookup(input []uint16, lookupRecord []tables.SequenceLookupRecord, lookupContext matcherFunc) bool {
	if len(input)+1 > maxContextLength {
		return false
	}
	hasMatch, matchEnd, _ := c.matchInput(input, lookupContext)
	if hasMatch {
		c.buffer.unsafeToBreak(c.buffer.idx, matchEnd)
		c.applyLookup(len(input)+1, lookupRecord, matchEnd)
		return true
	} else {
		c.buffer.unsafeToConcat(c.buffer.idx, matchEnd)
		return false
	}
}

//	`input` starts with second glyph (`inputCount` = len(input)+1)
//
// lookupsContexts : backtrack, input, lookahead
func (c *otApplyContext) chainContextApplyLookup(backtrack, input, lookahead []uint16,
	lookupRecord []tables.SequenceLookupRecord, lookupContexts [3]matcherFunc,
) bool {
	if len(input)+1 > maxContextLength {
		return false
	}

	hasMatch, matchEnd, _ := c.matchInput(input, lookupContexts[1])
	endIndex := matchEnd
	if !(hasMatch && endIndex != 0) {
		c.buffer.unsafeToConcat(c.buffer.idx, endIndex)
		return false
	}

	hasMatch, endIndex = c.matchLookahead(lookahead, lookupContexts[2], matchEnd)
	if !hasMatch {
		c.buffer.unsafeToConcat(c.buffer.idx, endIndex)
		return false
	}

	hasMatch, startIndex := c.matchBacktrack(backtrack, lookupContexts[0])
	if !hasMatch {
		c.buffer.unsafeToConcatFromOutbuffer(startIndex, endIndex)
		return false
	}

	c.buffer.unsafeToBreakFromOutbuffer(startIndex, endIndex)
	c.applyLookup(len(input)+1, lookupRecord, matchEnd)
	return true
}

func (c *wouldApplyContext) wouldApplyLookupContext1(data tables.SequenceContextFormat1, index int) bool {
	if index >= len(data.SeqRuleSet) { // index is not sanitized in tt.Parse
		return false
	}
	ruleSet := data.SeqRuleSet[index]
	return c.wouldApplyRuleSet(ruleSet, matchGlyph)
}

func (c *wouldApplyContext) wouldApplyLookupContext2(data tables.SequenceContextFormat2, _ int, glyphID GID) bool {
	class, _ := data.ClassDef.Class(gID(glyphID))
	ruleSet := data.ClassSeqRuleSet[class]
	return c.wouldApplyRuleSet(ruleSet, matchClass(data.ClassDef))
}

func (c *wouldApplyContext) wouldApplyLookupContext3(data tables.SequenceContextFormat3, _ int) bool {
	covIndices := get1N(&c.indices, 1, len(data.Coverages))
	return c.wouldMatchInput(covIndices, matchCoverage(data.Coverages))
}

func (c *wouldApplyContext) wouldApplyRuleSet(ruleSet tables.SequenceRuleSet, match matcherFunc) bool {
	for _, rule := range ruleSet.SeqRule {
		if c.wouldMatchInput(rule.InputSequence, match) {
			return true
		}
	}
	return false
}

func (c *wouldApplyContext) wouldApplyChainRuleSet(ruleSet tables.ChainedSequenceRuleSet, inputMatch matcherFunc) bool {
	for _, rule := range ruleSet.ChainedSeqRules {
		if c.wouldApplyChainLookup(rule.BacktrackSequence, rule.InputSequence, rule.LookaheadSequence, inputMatch) {
			return true
		}
	}
	return false
}

func (c *wouldApplyContext) wouldApplyLookupChainedContext1(data tables.ChainedSequenceContextFormat1, index int) bool {
	if index >= len(data.ChainedSeqRuleSet) { // index is not sanitized in tt.Parse
		return false
	}
	ruleSet := data.ChainedSeqRuleSet[index]
	return c.wouldApplyChainRuleSet(ruleSet, matchGlyph)
}

func (c *wouldApplyContext) wouldApplyLookupChainedContext2(data tables.ChainedSequenceContextFormat2, _ int, glyphID GID) bool {
	class, _ := data.InputClassDef.Class(gID(glyphID))
	ruleSet := data.ChainedClassSeqRuleSet[class]
	return c.wouldApplyChainRuleSet(ruleSet, matchClass(data.InputClassDef))
}

func (c *wouldApplyContext) wouldApplyLookupChainedContext3(data tables.ChainedSequenceContextFormat3, _ int) bool {
	lB, lI, lL := len(data.BacktrackCoverages), len(data.InputCoverages), len(data.LookaheadCoverages)
	return c.wouldApplyChainLookup(get1N(&c.indices, 0, lB), get1N(&c.indices, 1, lI), get1N(&c.indices, 0, lL),
		matchCoverage(data.InputCoverages))
}
