// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package font

import (
	"fmt"

	"github.com/OpenMellon/typesetting/font/cff"
	ot "github.com/OpenMellon/typesetting/font/opentype"
	"github.com/OpenMellon/typesetting/font/opentype/tables"
)

// return nil if the table is missing or invalid
func loadCff(ld *ot.Loader, numGlyphs int) (*cff.CFF, error) {
	raw, err := ld.RawTable(ot.MustNewTag("CFF "))
	if err != nil {
		return nil, err
	}
	cff, err := cff.Parse(raw)
	if err != nil {
		return nil, err
	}

	if N := len(cff.Charstrings); N != numGlyphs {
		return nil, fmt.Errorf("invalid number of glyphs in CFF table (%d != %d)", N, numGlyphs)
	}
	return cff, nil
}

// return nil if the table is missing or invalid
func loadCff2(ld *ot.Loader, numGlyphs, axisCount int) (*cff.CFF2, error) {
	raw, err := ld.RawTable(ot.MustNewTag("CFF2"))
	if err != nil {
		return nil, err
	}
	cff2, err := cff.ParseCFF2(raw)
	if err != nil {
		return nil, err
	}

	if N := len(cff2.Charstrings); N != numGlyphs {
		return nil, fmt.Errorf("invalid number of glyphs in CFF table (%d != %d)", N, numGlyphs)
	}

	if got := cff2.VarStore.AxisCount(); got != -1 && got != axisCount {
		return nil, fmt.Errorf("invalid number of axis in CFF table (%d != %d)", got, axisCount)
	}
	return cff2, nil
}

func loadHVtmx(hheaRaw, htmxRaw []byte, numGlyphs int) (*tables.Hhea, tables.Hmtx, error) {
	hhea, _, err := tables.ParseHhea(hheaRaw)
	if err != nil {
		return nil, tables.Hmtx{}, err
	}

	hmtx, _, err := tables.ParseHmtx(htmxRaw, int(hhea.NumOfLongMetrics), numGlyphs-int(hhea.NumOfLongMetrics))
	if err != nil {
		return nil, tables.Hmtx{}, err
	}
	return &hhea, hmtx, nil
}

func loadHmtx(ld *ot.Loader, numGlyphs int) (*tables.Hhea, tables.Hmtx, error) {
	rawHead, err := ld.RawTable(ot.MustNewTag("hhea"))
	if err != nil {
		return nil, tables.Hmtx{}, err
	}

	rawMetrics, err := ld.RawTable(ot.MustNewTag("hmtx"))
	if err != nil {
		return nil, tables.Hmtx{}, err
	}

	return loadHVtmx(rawHead, rawMetrics, numGlyphs)
}

func loadVmtx(ld *ot.Loader, numGlyphs int) (*tables.Hhea, tables.Hmtx, error) {
	rawHead, err := ld.RawTable(ot.MustNewTag("vhea"))
	if err != nil {
		return nil, tables.Hmtx{}, err
	}

	rawMetrics, err := ld.RawTable(ot.MustNewTag("vmtx"))
	if err != nil {
		return nil, tables.Hmtx{}, err
	}

	return loadHVtmx(rawHead, rawMetrics, numGlyphs)
}

func loadGDEF(ld *ot.Loader, axisCount int, gsub, gpos []byte) (tables.GDEF, error) {
	raw, err := ld.RawTable(ot.MustNewTag("GDEF"))
	if err != nil {
		return tables.GDEF{}, err
	}

	// Nuke the GDEF tables of to avoid unwanted width-zeroing.
	if isGDEFBlocklisted(raw, gsub, gpos) {
		return tables.GDEF{}, nil
	}

	GDEF, _, err := tables.ParseGDEF(raw)
	if err != nil {
		return tables.GDEF{}, err
	}

	err = sanitizeGDEF(GDEF, axisCount)
	if err != nil {
		return tables.GDEF{}, err
	}
	return GDEF, nil
}

// Face is a font with user-provided settings.
// Contrary to the [*Font] objects, Faces are NOT safe for concurrent use.
// A Face caches glyph extents and rune to glyph mapping, and should be reused when possible.
//
// Also note that an empty [Face] is invalid : the [NewFace] constructor is required to properly init caches.
type Face struct {
	*Font

	extentsCache extentsCache
	cmapCache    cache21_19_8

	coords       []tables.Coord
	xPpem, yPpem uint16
}

// NewFace wraps [font] and initializes glyph caches.
func NewFace(font *Font) *Face {
	out := &Face{Font: font, extentsCache: make(extentsCache, font.nGlyphs)}
	out.cmapCache.clear()
	return out
}

// NominalGlyph returns the glyph used to represent the given rune,
// or false if not found.
// Note that it only looks into the cmap, without taking account substitutions
// nor variation selectors.
func (f *Face) NominalGlyph(ch rune) (GID, bool) {
	if g, ok := f.cmapCache.get(uint32(ch)); ok {
		return GID(g), ok
	}
	g, ok := f.Cmap.Lookup(ch)
	if ok {
		f.cmapCache.set(uint32(ch), uint32(g))
	}
	return g, ok
}

// Ppem returns the horizontal and vertical pixels-per-em (ppem), used to select bitmap sizes.
func (f *Face) Ppem() (x, y uint16) { return f.xPpem, f.yPpem }

// SetPpem applies horizontal and vertical pixels-per-em (ppem).
func (f *Face) SetPpem(x, y uint16) {
	f.xPpem, f.yPpem = x, y
	// invalid the cache
	f.extentsCache.reset()
}

// Coords return a read-only slice of the current variable coordinates, expressed in normalized units.
// It is empty for non variable fonts.
func (f *Face) Coords() []tables.Coord { return f.coords }

// SetCoords applies a list of variation coordinates, expressed in normalized units.
// Use [NormalizeVariations] to convert from design (user) space units.
func (f *Face) SetCoords(coords []tables.Coord) {
	f.coords = coords
	// invalid the cache
	f.extentsCache.reset()
}
