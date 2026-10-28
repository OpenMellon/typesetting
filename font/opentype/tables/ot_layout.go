// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package tables

import (
	"errors"
	"fmt"
)

// The following are types shared by GSUB and GPOS tables

// Coverage specifies all the glyphs affected by a substitution or
// positioning operation described in a subtable.
// Conceptually is it a []GlyphIndex, with an Index method,
// but it may be implemented for efficiently.
// See https://learn.microsoft.com/typography/opentype/spec/chapter2#lookup-table
type Coverage interface {
	isCov()

	// Index returns the index of the provided glyph, or
	// `false` if the glyph is not covered by this lookup.
	// Note: this method is injective: two distincts, covered glyphs are mapped
	// to distincts indices.
	Index(GlyphID) (int, bool)

	// Len return the number of glyphs covered.
	// It is 0 for empty coverages.
	// For non empty Coverages, it is also 1 + (maximum index returned)
	Len() int
}

func (Coverage1) isCov() {}
func (Coverage2) isCov() {}

type Coverage1 struct {
	format uint16    `unionTag:"1"`
	Glyphs []GlyphID `arrayCount:"FirstUint16"`
}

type Coverage2 struct {
	format uint16        `unionTag:"2"`
	Ranges []RangeRecord `arrayCount:"FirstUint16"`
}

type RangeRecord struct {
	StartGlyphID       GlyphID // First glyph ID in the range
	EndGlyphID         GlyphID // Last glyph ID in the range
	StartCoverageIndex uint16  // Coverage Index of first glyph ID in range
}

// ClassDef stores a value for a set of GlyphIDs.
// Conceptually it is a map[GlyphID]uint16, but it may
// be implemented more efficiently.
type ClassDef interface {
	isClassDef()
	Class(gi GlyphID) (uint16, bool)

	// Extent returns the maximum class ID + 1. This is the length
	// required for an array to be indexed by the class values.
	Extent() int
}

func (ClassDef1) isClassDef() {}
func (ClassDef2) isClassDef() {}

type ClassDef1 struct {
	format          uint16   `unionTag:"1"`
	StartGlyphID    GlyphID  // First glyph ID of the classValueArray
	ClassValueArray []uint16 `arrayCount:"FirstUint16"` //[glyphCount]	Array of Class Values — one per glyph ID
}

type ClassDef2 struct {
	format            uint16             `unionTag:"2"`
	ClassRangeRecords []ClassRangeRecord `arrayCount:"FirstUint16"` //[glyphCount]	Array of Class Values — one per glyph ID
}

type ClassRangeRecord struct {
	StartGlyphID GlyphID // First glyph ID in the range
	EndGlyphID   GlyphID // Last glyph ID in the range
	Class        uint16  // Applied to all glyphs in the range
}

// Lookups

type SequenceLookupRecord struct {
	SequenceIndex   uint16 // Index (zero-based) into the input glyph sequence
	LookupListIndex uint16 // Index (zero-based) into the LookupList
}

type SequenceContextFormat1 struct {
	format     uint16            `unionTag:"1"`
	coverage   Coverage          `offsetSize:"Offset16"`                             // Offset to Coverage table, from beginning of SequenceContextFormat1 table
	SeqRuleSet []SequenceRuleSet `arrayCount:"FirstUint16"  offsetsArray:"Offset16"` //[seqRuleSetCount]	Array of offsets to SequenceRuleSet tables, from beginning of SequenceContextFormat1 table (offsets may be NULL)
}

func (sc *SequenceContextFormat1) sanitize(lookupCount uint16) error {
	for _, set := range sc.SeqRuleSet {
		for _, rule := range set.SeqRule {
			if err := rule.sanitize(lookupCount); err != nil {
				return err
			}
		}
	}
	return nil
}

type SequenceRuleSet struct {
	SeqRule []SequenceRule `arrayCount:"FirstUint16" offsetsArray:"Offset16"` // Array of offsets to SequenceRule tables, from beginning of the SequenceRuleSet table
}

type SequenceRule struct {
	glyphCount       uint16                 // Number of glyphs in the input glyph sequence
	seqLookupCount   uint16                 // Number of SequenceLookupRecords
	InputSequence    []GlyphID              `arrayCount:"ComputedField-glyphCount-1"`   //[glyphCount - 1]	Array of input glyph IDs—starting with the second glyph
	SeqLookupRecords []SequenceLookupRecord `arrayCount:"ComputedField-seqLookupCount"` //[seqLookupCount]	Array of Sequence lookup records
}

func (sr *SequenceRule) sanitize(lookupCount uint16) error {
	for _, rec := range sr.SeqLookupRecords {
		if rec.SequenceIndex >= sr.glyphCount {
			return fmt.Errorf("invalid sequence lookup table (input index %d >= %d)", rec.SequenceIndex, sr.glyphCount)
		}
		if rec.LookupListIndex >= lookupCount {
			return fmt.Errorf("invalid sequence lookup table (lookup index %d >= %d)", rec.LookupListIndex, lookupCount)
		}
	}
	return nil
}

type SequenceContextFormat2 struct {
	format          uint16                 `unionTag:"2"`
	coverage        Coverage               `offsetSize:"Offset16"`                            //	Offset to Coverage table, from beginning of SequenceContextFormat2 table
	ClassDef        ClassDef               `offsetSize:"Offset16"`                            //	Offset to ClassDef table, from beginning of SequenceContextFormat2 table
	ClassSeqRuleSet []ClassSequenceRuleSet `arrayCount:"FirstUint16" offsetsArray:"Offset16"` //[classSeqRuleSetCount]	Array of offsets to ClassSequenceRuleSet tables, from beginning of SequenceContextFormat2 table (may be NULL)
}

// ClassSequenceRuleSet has the same binary format as SequenceRuleSet,
// and using the same type simplifies later processing.
type ClassSequenceRuleSet = SequenceRuleSet

type SequenceContextFormat3 struct {
	format           uint16                 `unionTag:"3"`
	glyphCount       uint16                 // Number of glyphs in the input sequence
	seqLookupCount   uint16                 // Number of SequenceLookupRecords
	Coverages        []Coverage             `arrayCount:"ComputedField-glyphCount" offsetsArray:"Offset16"` //[glyphCount]	Array of offsets to Coverage tables, from beginning of SequenceContextFormat3 subtable
	SeqLookupRecords []SequenceLookupRecord `arrayCount:"ComputedField-seqLookupCount"`                     //[seqLookupCount]	Array of SequenceLookupRecords
}

type ChainedSequenceContextFormat1 struct {
	format            uint16                   `unionTag:"1"`
	coverage          Coverage                 `offsetSize:"Offset16"`                             //	Offset to Coverage table, from beginning of ChainSequenceContextFormat1 table
	ChainedSeqRuleSet []ChainedSequenceRuleSet `arrayCount:"FirstUint16"  offsetsArray:"Offset16"` //[chainedSeqRuleSetCount]	Array of offsets to ChainedSeqRuleSet tables, from beginning of ChainedSequenceContextFormat1 table (may be NULL)
}

type ChainedSequenceRuleSet struct {
	ChainedSeqRules []ChainedSequenceRule `arrayCount:"FirstUint16" offsetsArray:"Offset16"` // Array of offsets to SequenceRule tables, from beginning of the SequenceRuleSet table
}

type ChainedSequenceRule struct {
	BacktrackSequence []GlyphID              `arrayCount:"FirstUint16"` //[backtrackGlyphCount]	Array of backtrack glyph IDs
	inputGlyphCount   uint16                 //	Number of glyphs in the input sequence
	InputSequence     []GlyphID              `arrayCount:"ComputedField-inputGlyphCount-1"` //[inputGlyphCount - 1]	Array of input glyph IDs—start with second glyph
	LookaheadSequence []GlyphID              `arrayCount:"FirstUint16"`                     //[lookaheadGlyphCount]	Array of lookahead glyph IDs
	SeqLookupRecords  []SequenceLookupRecord `arrayCount:"FirstUint16"`                     //[seqLookupCount]	Array of SequenceLookupRecords
}

type ChainedSequenceContextFormat2 struct {
	format                 uint16                        `unionTag:"2"`
	coverage               Coverage                      `offsetSize:"Offset16"`                            // Offset to Coverage table, from beginning of ChainedSequenceContextFormat2 table
	BacktrackClassDef      ClassDef                      `offsetSize:"Offset16"`                            // Offset to ClassDef table containing backtrack sequence context, from beginning of ChainedSequenceContextFormat2 table
	InputClassDef          ClassDef                      `offsetSize:"Offset16"`                            // Offset to ClassDef table containing input sequence context, from beginning of ChainedSequenceContextFormat2 table
	LookaheadClassDef      ClassDef                      `offsetSize:"Offset16"`                            // Offset to ClassDef table containing lookahead sequence context, from beginning of ChainedSequenceContextFormat2 table
	ChainedClassSeqRuleSet []ChainedClassSequenceRuleSet `arrayCount:"FirstUint16" offsetsArray:"Offset16"` //[chainedClassSeqRuleSetCount]	Array of offsets to ChainedClassSequenceRuleSet tables, from beginning of ChainedSequenceContextFormat2 table (may be NULL)
}

// ChainedClassSequenceRuleSet has the same binary format as ChainedSequenceRuleSet,
// and using the same type simplifies later processing.
type ChainedClassSequenceRuleSet = ChainedSequenceRuleSet

type ChainedSequenceContextFormat3 struct {
	format             uint16                 `unionTag:"3"`
	BacktrackCoverages []Coverage             `arrayCount:"FirstUint16" offsetsArray:"Offset16"` //[backtrackGlyphCount]	Array of offsets to coverage tables for the backtrack sequence
	InputCoverages     []Coverage             `arrayCount:"FirstUint16" offsetsArray:"Offset16"` //[inputGlyphCount]	Array of offsets to coverage tables for the input sequence
	LookaheadCoverages []Coverage             `arrayCount:"FirstUint16" offsetsArray:"Offset16"` //[lookaheadGlyphCount]	Array of offsets to coverage tables for the lookahead sequence
	SeqLookupRecords   []SequenceLookupRecord `arrayCount:"FirstUint16"`                         //[seqLookupCount]	Array of SequenceLookupRecords
}

type Extension struct {
	substFormat         uint16   //	Format identifier. Set to 1.
	ExtensionLookupType uint16   //	Lookup type of subtable referenced by extensionOffset (that is, the extension subtable).
	ExtensionOffset     Offset32 //	Offset to the extension subtable, of lookup type extensionLookupType, relative to the start of the ExtensionSubstFormat1 subtable.
	RawData             []byte   `subsliceStart:"AtStart" arrayCount:"ToEnd"`
}

// GSUB is the Glyph Substitution (GSUB) table.
// It provides data for substition of glyphs for appropriate rendering of scripts,
// such as cursively-connecting forms in Arabic script,
// or for advanced typographic effects, such as ligatures.
// See https://learn.microsoft.com/fr-fr/typography/opentype/spec/gsub
type GSUB Layout

// GSUBLookup is one lookup subtable data
type GSUBLookup interface {
	isGSUBLookup()

	// Coverage returns the coverage of the lookup subtable.
	// For ContextualSubs3 and ChainedContextualSubs3, its the coverage of the first input.
	Cov() Coverage
}

func (SingleSubs) isGSUBLookup()             {}
func (MultipleSubs) isGSUBLookup()           {}
func (AlternateSubs) isGSUBLookup()          {}
func (LigatureSubs) isGSUBLookup()           {}
func (ContextualSubs) isGSUBLookup()         {}
func (ChainedContextualSubs) isGSUBLookup()  {}
func (ExtensionSubs) isGSUBLookup()          {}
func (ReverseChainSingleSubs) isGSUBLookup() {}

func (ms MultipleSubs) Sanitize() error {
	if exp, got := ms.Coverage.Len(), len(ms.Sequences); exp != got {
		return fmt.Errorf("GSUB: invalid MultipleSubs sequences count (%d != %d)", exp, got)
	}
	return nil
}

func (ls LigatureSubs) Sanitize() error {
	if exp, got := ls.Coverage.Len(), len(ls.LigatureSets); exp != got {
		return fmt.Errorf("GSUB: invalid LigatureSubs sets count (%d != %d)", exp, got)
	}
	return nil
}

func (cs ContextualSubs) Sanitize(lookupCount uint16) error {
	if f1, isFormat1 := cs.Data.(ContextualSubs1); isFormat1 {
		return (*SequenceContextFormat1)(&f1).sanitize(lookupCount)
	}
	return nil
}

func (rs ReverseChainSingleSubs) Sanitize() error {
	if exp, got := rs.coverage.Len(), len(rs.SubstituteGlyphIDs); exp != got {
		return fmt.Errorf("GSUB: invalid ReverseChainSingleSubs glyphs count (%d != %d)", exp, got)
	}
	return nil
}

func (ext ExtensionSubs) Resolve() (GSUBLookup, error) {
	if L, E := len(ext.RawData), int(ext.ExtensionOffset); L < E {
		return nil, fmt.Errorf("EOF: expected length: %d, got %d", E, L)
	}
	lk, err := parseGSUBLookup(ext.RawData[ext.ExtensionOffset:], ext.ExtensionLookupType)
	if err != nil {
		return nil, err
	}
	if _, isExt := lk.(ExtensionSubs); isExt {
		return nil, errors.New("invalid extension substitution table")
	}
	return lk, nil
}

func parseGSUBLookup(src []byte, lookupType uint16) (out GSUBLookup, err error) {
	switch lookupType {
	case 1: // Single (format 1.1 1.2)	Replace one glyph with one glyph
		out, _, err = ParseSingleSubs(src)
	case 2: // Multiple (format 2.1)	Replace one glyph with more than one glyph
		out, _, err = ParseMultipleSubs(src)
	case 3: // Alternate (format 3.1)	Replace one glyph with one of many glyphs
		out, _, err = ParseAlternateSubs(src)
	case 4: // Ligature (format 4.1)	Replace multiple glyphs with one glyph
		out, _, err = ParseLigatureSubs(src)
	case 5: // Context (format 5.1 5.2 5.3)	Replace one or more glyphs in context
		out, _, err = ParseContextualSubs(src)
	case 6: // Chaining Context (format 6.1 6.2 6.3)	Replace one or more glyphs in chained context
		out, _, err = ParseChainedContextualSubs(src)
	case 7: // Extension Substitution (format 7.1) Extension mechanism for other substitutions
		out, _, err = ParseExtensionSubs(src)
	case 8: // Reverse chaining context single (format 8.1)
		out, _, err = ParseReverseChainSingleSubs(src)
	default:
		err = fmt.Errorf("invalid GSUB Loopkup type %d", lookupType)
	}
	return out, err
}

// AsGSUBLookups returns the GSUB lookup subtables.
func (lk Lookup) AsGSUBLookups() ([]GSUBLookup, error) {
	var err error
	out := make([]GSUBLookup, len(lk.subtableOffsets))
	for i, offset := range lk.subtableOffsets {
		if L := len(lk.rawData); L < int(offset) {
			return nil, fmt.Errorf("EOF: expected length: %d, got %d", offset, L)
		}
		out[i], err = parseGSUBLookup(lk.rawData[offset:], lk.lookupType)
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

// ------------------------ GPOS common data structures ------------------------

// GPOS is the Glyph Positioning (GPOS) table.
// It provides precise control over glyph placement
// for sophisticated text layout and rendering in each script
// and language system that a font supports.
// See https://learn.microsoft.com/fr-fr/typography/opentype/spec/gpos
type GPOS Layout

type GPOSLookup interface {
	isGPOSLookup()

	// Coverage returns the coverage of the lookup subtable.
	// For ContextualPos3 and ChainedContextualPos3, its the coverage of the first input.
	Cov() Coverage
}

func (SinglePos) isGPOSLookup()            {}
func (PairPos) isGPOSLookup()              {}
func (CursivePos) isGPOSLookup()           {}
func (MarkBasePos) isGPOSLookup()          {}
func (MarkLigPos) isGPOSLookup()           {}
func (MarkMarkPos) isGPOSLookup()          {}
func (ContextualPos) isGPOSLookup()        {}
func (ChainedContextualPos) isGPOSLookup() {}
func (ExtensionPos) isGPOSLookup()         {}

func (sp *SinglePos) Sanitize() error {
	if f2, isFormat2 := sp.Data.(SinglePosData2); isFormat2 {
		if exp, got := f2.coverage.Len(), len(f2.ValueRecords); exp != got {
			return fmt.Errorf("GPOS: invalid SinglePos values count (%d != %d)", exp, got)
		}
	}
	return nil
}

func (pp *PairPos) Sanitize() error {
	if f1, isFormat1 := pp.Data.(PairPosData1); isFormat1 {
		// there are fonts with to much PairSets : accept it
		if exp, got := f1.coverage.Len(), len(f1.PairSets); exp > got {
			return fmt.Errorf("GPOS: invalid PairPos1 sets count (%d > %d)", exp, got)
		}
	} else if f2, isFormat2 := pp.Data.(PairPosData2); isFormat2 {
		if exp, got := f2.ClassDef1.Extent(), int(f2.class1Count); exp != got {
			return fmt.Errorf("GPOS: invalid PairPos2 class1 count (%d != %d)", exp, got)
		}
		if exp, got := f2.ClassDef2.Extent(), int(f2.class2Count); exp != got {
			return fmt.Errorf("GPOS: invalid PairPos2 class2 count (%d != %d)", exp, got)
		}
	}
	return nil
}

func (mp *MarkBasePos) Sanitize() error {
	if exp, got := mp.markCoverage.Len(), len(mp.MarkArray.MarkRecords); exp != got {
		return fmt.Errorf("GPOS: invalid MarkBasePos marks count (%d != %d)", exp, got)
	}
	if exp, got := mp.BaseCoverage.Len(), len(mp.BaseArray.baseRecords); exp != got {
		return fmt.Errorf("GPOS: invalid MarkBasePos marks count (%d != %d)", exp, got)
	}
	if err := mp.BaseArray.Anchors().sanitizeOffsets(); err != nil {
		return err
	}

	return nil
}

func (mp *MarkLigPos) Sanitize() error {
	if exp, got := mp.MarkCoverage.Len(), len(mp.MarkArray.MarkAnchors); exp != got {
		return fmt.Errorf("GPOS: invalid MarkBasePos marks count (%d != %d)", exp, got)
	}
	if exp, got := mp.LigatureCoverage.Len(), len(mp.LigatureArray.LigatureAttachs); exp != got {
		return fmt.Errorf("GPOS: invalid MarkBasePos marks count (%d != %d)", exp, got)
	}

	return nil
}
