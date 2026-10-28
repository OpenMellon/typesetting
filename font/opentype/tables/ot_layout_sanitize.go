// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package tables

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/bits"
)

func (cs *ContextualPos) Sanitize(lookupCount uint16) error {
	if f1, isFormat1 := cs.Data.(ContextualPos1); isFormat1 {
		return (*SequenceContextFormat1)(&f1).sanitize(lookupCount)
	}
	return nil
}

func (ext ExtensionPos) Resolve() (GPOSLookup, error) {
	if L, E := len(ext.RawData), int(ext.ExtensionOffset); L < E {
		return nil, fmt.Errorf("EOF: expected length: %d, got %d", E, L)
	}
	lk, err := parseGPOSLookup(ext.RawData[ext.ExtensionOffset:], ext.ExtensionLookupType)
	if err != nil {
		return nil, err
	}
	if _, isExt := lk.(ExtensionPos); isExt {
		return nil, errors.New("invalid extension positioning table")
	}
	return lk, nil
}

func parseGPOSLookup(src []byte, lookupType uint16) (out GPOSLookup, err error) {
	switch lookupType {
	case 1: // Single adjustment	Adjust position of a single glyph
		out, _, err = ParseSinglePos(src)
	case 2: // Pair adjustment	Adjust position of a pair of glyphs
		out, _, err = ParsePairPos(src)
	case 3: // Cursive attachment	Attach cursive glyphs
		out, _, err = ParseCursivePos(src)
	case 4: // MarkToBase attachment	Attach a combining mark to a base glyph
		out, _, err = ParseMarkBasePos(src)
	case 5: // MarkToLigature attachment	Attach a combining mark to a ligature
		out, _, err = ParseMarkLigPos(src)
	case 6: // MarkToMark attachment	Attach a combining mark to another mark
		out, _, err = ParseMarkMarkPos(src)
	case 7: // Context positioning	Position one or more glyphs in context
		out, _, err = ParseContextualPos(src)
	case 8: // Chained Context positioning	Position one or more glyphs in chained context
		out, _, err = ParseChainedContextualPos(src)
	case 9: // Extension positioning	Extension mechanism for other positionings
		out, _, err = ParseExtensionPos(src)
	default:
		err = fmt.Errorf("invalid GPOS Loopkup type %d", lookupType)
	}
	return out, err
}

// AsGPOSLookups returns the GPOS lookup subtables
func (lk Lookup) AsGPOSLookups() ([]GPOSLookup, error) {
	var err error
	out := make([]GPOSLookup, len(lk.subtableOffsets))
	for i, offset := range lk.subtableOffsets {
		if L := len(lk.rawData); L < int(offset) {
			return nil, fmt.Errorf("EOF: expected length: %d, got %d", offset, L)
		}
		out[i], err = parseGPOSLookup(lk.rawData[offset:], lk.lookupType)
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

// ValueFormat is a mask indicating which field
// are set in a GPOS [ValueRecord].
// It is often shared between many records.
type ValueFormat uint16

// number of fields present
func (f ValueFormat) size() int { return bits.OnesCount16(uint16(f)) }

const (
	XPlacement ValueFormat = 1 << iota // Includes horizontal adjustment for placement
	YPlacement                         // Includes vertical adjustment for placement
	XAdvance                           // Includes horizontal adjustment for advance
	YAdvance                           // Includes vertical adjustment for advance
	XPlaDevice                         // Includes horizontal Device table for placement
	YPlaDevice                         // Includes vertical Device table for placement
	XAdvDevice                         // Includes horizontal Device table for advance
	YAdvDevice                         // Includes vertical Device table for advance

	//  Mask for having any Device table
	Devices = XPlaDevice | YPlaDevice | XAdvDevice | YAdvDevice
)

// ValueRecord has optional fields
type ValueRecord struct {
	XPlacement int16       // Horizontal adjustment for placement, in design units.
	YPlacement int16       // Vertical adjustment for placement, in design units.
	XAdvance   int16       // Horizontal adjustment for advance, in design units — only used for horizontal layout.
	YAdvance   int16       // Vertical adjustment for advance, in design units — only used for vertical layout.
	XPlaDevice DeviceTable // Offset to Device table (non-variable font) / VariationIndex table (variable font) for horizontal placement, from beginning of the immediate parent table (SinglePos or PairPosFormat2 lookup subtable, PairSet table within a PairPosFormat1 lookup subtable) — may be NULL.
	YPlaDevice DeviceTable // Offset to Device table (non-variable font) / VariationIndex table (variable font) for vertical placement, from beginning of the immediate parent table (SinglePos or PairPosFormat2 lookup subtable, PairSet table within a PairPosFormat1 lookup subtable) — may be NULL.
	XAdvDevice DeviceTable // Offset to Device table (non-variable font) / VariationIndex table (variable font) for horizontal advance, from beginning of the immediate parent table (SinglePos or PairPosFormat2 lookup subtable, PairSet table within a PairPosFormat1 lookup subtable) — may be NULL.
	YAdvDevice DeviceTable // Offset to Device table (non-variable font) / VariationIndex table (variable font) for vertical advance, from beginning of the immediate parent table (SinglePos or PairPosFormat2 lookup subtable, PairSet table within a PairPosFormat1 lookup su
}

// [data] must start at the immediate parent table, [offset] indicating
// the start of the record in it.
// Returns [offset] + the number of bytes read from [offset]
// Note that a [format] with value 0, is supported, resulting in a no-op
func parseValueRecord(format ValueFormat, data []byte, offset int) (out ValueRecord, _ int, err error) {
	if L := len(data); L < offset {
		return out, 0, fmt.Errorf("EOF: expected length: %d, got %d", offset, L)
	}

	size := format.size() // number of fields present
	if size == 0 {        // return early
		return out, offset, nil
	}
	// start by parsing the list of values
	values, err := ParseUint16s(data[offset:], size)
	if err != nil {
		return out, 0, fmt.Errorf("invalid value record: %s", err)
	}
	// follow the order
	cursor := 0
	if format&XPlacement != 0 {
		out.XPlacement = int16(values[cursor])
		cursor++
	}
	if format&YPlacement != 0 {
		out.YPlacement = int16(values[cursor])
		cursor++
	}
	if format&XAdvance != 0 {
		out.XAdvance = int16(values[cursor])
		cursor++
	}
	if format&YAdvance != 0 {
		out.YAdvance = int16(values[cursor])
		cursor++
	}
	if format&XPlaDevice != 0 {
		if devOffset := values[cursor]; devOffset != 0 {
			out.XPlaDevice, err = parseDeviceTable(data, devOffset)
			if err != nil {
				return out, 0, err
			}
		}
		cursor++
	}
	if format&YPlaDevice != 0 {
		if devOffset := values[cursor]; devOffset != 0 {
			out.YPlaDevice, err = parseDeviceTable(data, devOffset)
			if err != nil {
				return out, 0, err
			}
		}
		cursor++
	}
	if format&XAdvDevice != 0 {
		if devOffset := values[cursor]; devOffset != 0 {
			out.XAdvDevice, err = parseDeviceTable(data, devOffset)
			if err != nil {
				return out, 0, err
			}
		}
		cursor++
	}
	if format&YAdvDevice != 0 {
		if devOffset := values[cursor]; devOffset != 0 {
			out.YAdvDevice, err = parseDeviceTable(data, devOffset)
			if err != nil {
				return out, 0, err
			}
		}
		cursor++ // useless actually
	}
	return out, offset + 2*size, err
}

type pairValueRecords struct {
	data       []byte // start with the item count
	fmt1, fmt2 ValueFormat
}

// panic if index is out of range
func (ps pairValueRecords) get(index int) (out PairValueRecord, err error) {
	recLen := 1 + ps.fmt1.size() + ps.fmt2.size()
	offset := 2 + 2*index*recLen

	out.SecondGlyph = GlyphID(binary.BigEndian.Uint16(ps.data[offset:]))
	v1, newOffset, err := parseValueRecord(ps.fmt1, ps.data, offset+2)
	if err != nil {
		return out, fmt.Errorf("invalid pair set table: %s", err)
	}
	v2, _, err := parseValueRecord(ps.fmt2, ps.data, newOffset)
	if err != nil {
		return out, fmt.Errorf("invalid pair set table: %s", err)
	}
	out.ValueRecord1 = v1
	out.ValueRecord2 = v2
	return out, nil
}

// DeviceTable is either an DeviceHinting for standard fonts,
// or a DeviceVariation for variable fonts.
type DeviceTable interface {
	isDevice()
}

func (DeviceHinting) isDevice()   {}
func (DeviceVariation) isDevice() {}

type DeviceHinting struct {
	// with length endSize - startSize + 1
	Values []int8
	// correction range, in ppem
	StartSize, EndSize uint16
}

type DeviceVariation VariationStoreIndex

func parseDeviceTable(src []byte, offset uint16) (DeviceTable, error) {
	if L := len(src); L < int(offset)+6 {
		return nil, fmt.Errorf("EOF: expected length: %d, got %d", offset+6, L)
	}
	var header DeviceTableHeader
	header.mustParse(src[offset:])

	switch format := header.deltaFormat; format {
	case 1, 2, 3:
		var out DeviceHinting

		out.StartSize, out.EndSize = header.first, header.second
		if out.EndSize < out.StartSize {
			return nil, errors.New("invalid positionning device subtable")
		}

		nbPerUint16 := 16 / (1 << format) // 8, 4 or 2
		outLength := int(out.EndSize - out.StartSize + 1)
		var count int
		if outLength%nbPerUint16 == 0 {
			count = outLength / nbPerUint16
		} else {
			// add padding
			count = outLength/nbPerUint16 + 1
		}
		uint16s, err := ParseUint16s(src[offset+6:], count)
		if err != nil {
			return nil, err
		}
		out.Values = make([]int8, count*nbPerUint16) // handle rounding error by reslicing after
		switch format {
		case 1:
			for i, u := range uint16s {
				uint16As2Bits(out.Values[i*8:], u)
			}
		case 2:
			for i, u := range uint16s {
				uint16As4Bits(out.Values[i*4:], u)
			}
		case 3:
			for i, u := range uint16s {
				uint16As8Bits(out.Values[i*2:], u)
			}
		}
		out.Values = out.Values[:outLength]
		return out, nil
	case 0x8000:
		return DeviceVariation{DeltaSetOuter: header.first, DeltaSetInner: header.second}, nil
	default:
		return nil, fmt.Errorf("unsupported positionning device subtable: %d", format)
	}
}

type PairValueRecord struct {
	SecondGlyph  GlyphID     //	Glyph ID of second glyph in the pair (first glyph is listed in the Coverage table).
	ValueRecord1 ValueRecord //	Positioning data for the first glyph in the pair.
	ValueRecord2 ValueRecord //	Positioning data for the second glyph in the pair.
}

type Class1Record []Class2Record //[class2Count]	Array of Class2 records, ordered by classes in classDef2.

type Class2Record struct {
	ValueRecord1 ValueRecord //	Positioning for first glyph — empty if valueFormat1 = 0.
	ValueRecord2 ValueRecord //	Positioning for second glyph — empty if valueFormat2 = 0.
}

func (AnchorFormat1) isAnchor() {}
func (AnchorFormat2) isAnchor() {}
func (AnchorFormat3) isAnchor() {}

type AnchorFormat1 struct {
	anchorFormat uint16 `unionTag:"1"`
	XCoordinate  int16  // Horizontal value, in design units
	YCoordinate  int16  // Vertical value, in design units
}

type AnchorFormat2 struct {
	anchorFormat uint16 `unionTag:"2"`
	XCoordinate  int16  // Horizontal value, in design units
	YCoordinate  int16  // Vertical value, in design units
	AnchorPoint  uint16 // Index to glyph contour point
}

type AnchorFormat3 struct {
	anchorFormat  uint16      `unionTag:"3"`
	XCoordinate   int16       // Horizontal value, in design units
	YCoordinate   int16       // Vertical value, in design units
	xDeviceOffset Offset16    // Offset to Device table (non-variable font) / VariationIndex table (variable font) for X coordinate, from beginning of Anchor table (may be NULL)
	yDeviceOffset Offset16    // Offset to Device table (non-variable font) / VariationIndex table (variable font) for Y coordinate, from beginning of Anchor table (may be NULL)
	XDevice       DeviceTable `isOpaque:""` // Offset to Device table (non-variable font) / VariationIndex table (variable font) for X coordinate, from beginning of Anchor table (may be NULL)
	YDevice       DeviceTable `isOpaque:""` // Offset to Device table (non-variable font) / VariationIndex table (variable font) for Y coordinate, from beginning of Anchor table (may be NULL)
}

func (af *AnchorFormat3) parseXDevice(src []byte) error {
	if af.xDeviceOffset == 0 {
		return nil
	}
	var err error
	af.XDevice, err = parseDeviceTable(src, uint16(af.xDeviceOffset))
	return err
}

func (af *AnchorFormat3) parseYDevice(src []byte) error {
	if af.yDeviceOffset == 0 {
		return nil
	}
	var err error
	af.YDevice, err = parseDeviceTable(src, uint16(af.yDeviceOffset))
	return err
}

// AnchorMatrix is a compact representation of a [][]Anchor
type AnchorMatrix struct {
	records []anchorOffsets
	data    []byte
}

func (am AnchorMatrix) Len() int { return len(am.records) }

func (am AnchorMatrix) sanitizeOffsets() error {
	for _, list := range am.records {
		for _, offset := range list.offsets {
			if offset == 0 {
				continue
			}
			if L := len(am.data); L < int(offset) {
				return fmt.Errorf("EOF: expected length: %d, got %d", offset, L)
			}
		}
	}
	return nil
}

func (am AnchorMatrix) Anchor(index, class int) Anchor {
	if len(am.records) < index {
		return nil
	}
	offsets := am.records[index].offsets
	if len(offsets) < class {
		return nil
	}
	offset := offsets[class]
	if offset == 0 {
		return nil
	}
	anchor, _, _ := ParseAnchor(am.data[offset:]) // offset is sanitized
	return anchor
}

type MarkArray struct {
	MarkRecords []MarkRecord `arrayCount:"FirstUint16"` //[markCount]	Array of MarkRecords, ordered by corresponding glyphs in the associated mark Coverage table.
	MarkAnchors []Anchor     `isOpaque:""`              // with same length as MarkRecords
}

func (ma *MarkArray) parseMarkAnchors(src []byte) error {
	ma.MarkAnchors = make([]Anchor, len(ma.MarkRecords))
	var err error
	for i, rec := range ma.MarkRecords {
		if L := len(src); L < int(rec.markAnchorOffset) {
			return fmt.Errorf("EOF: expected length: %d, got %d", rec.markAnchorOffset, L)
		}
		ma.MarkAnchors[i], _, err = ParseAnchor(src[rec.markAnchorOffset:])
		if err != nil {
			return err
		}
	}
	return nil
}

type MarkRecord struct {
	MarkClass        uint16   // Class defined for the associated mark.
	markAnchorOffset Offset16 // Offset to Anchor table, from beginning of MarkArray table.
}

// ------------------------------ parsing helpers ------------------------------

// write 8 elements
func uint16As2Bits(dst []int8, u uint16) {
	const mask = 0xFE // 11111110
	dst[0] = int8((0-uint8(u>>15&1))&mask | uint8(u>>14&1))
	dst[1] = int8((0-uint8(u>>13&1))&mask | uint8(u>>12&1))
	dst[2] = int8((0-uint8(u>>11&1))&mask | uint8(u>>10&1))
	dst[3] = int8((0-uint8(u>>9&1))&mask | uint8(u>>8&1))
	dst[4] = int8((0-uint8(u>>7&1))&mask | uint8(u>>6&1))
	dst[5] = int8((0-uint8(u>>5&1))&mask | uint8(u>>4&1))
	dst[6] = int8((0-uint8(u>>3&1))&mask | uint8(u>>2&1))
	dst[7] = int8((0-uint8(u>>1&1))&mask | uint8(u>>0&1))
}

// write 4 elements
func uint16As4Bits(dst []int8, u uint16) {
	const mask = 0xF8 // 11111000

	dst[0] = int8((0-uint8(u>>15&1))&mask | uint8(u>>12&0x07))
	dst[1] = int8((0-uint8(u>>11&1))&mask | uint8(u>>8&0x07))
	dst[2] = int8((0-uint8(u>>7&1))&mask | uint8(u>>4&0x07))
	dst[3] = int8((0-uint8(u>>3&1))&mask | uint8(u>>0&0x07))
}

// write 2 elements
func uint16As8Bits(dst []int8, u uint16) {
	dst[0] = int8(u >> 8)
	dst[1] = int8(u)
}

// ParseUint16s interprets data as a (big endian) uint16 slice.
// It returns an error if [data] is not long enough for the given [count].
func ParseUint16s(src []byte, count int) ([]uint16, error) {
	if L := len(src); L < 2*count {
		return nil, fmt.Errorf("EOF: expected length: %d, got %d", 2*count, L)
	}
	out := make([]uint16, count)
	for i := range out {
		out[i] = binary.BigEndian.Uint16(src[2*i:])
	}
	return out, nil
}
