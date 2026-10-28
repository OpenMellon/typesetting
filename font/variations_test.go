// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package font

import (
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"

	ot "github.com/OpenMellon/typesetting/font/opentype"
	"github.com/OpenMellon/typesetting/font/opentype/tables"
	tu "github.com/OpenMellon/typesetting/testutils"
)

// ported from harfbuzz/test/api/test-var-coords.c Copyright © 2019 Ebrahim Byagowi

func TestVar(t *testing.T) {
	font := loadFont(t, "toys/CFF2-VF.otf")

	for _, test := range []struct {
		design     float32
		normalized VarCoord
	}{
		{200, -16384},
		{206, -16117},
		{400, 0},
		{450, 1529},
		{600, 6014},
	} {
		coords := font.NormalizeVariations([]float32{test.design})
		tu.AssertC(t, coords[0] == test.normalized, fmt.Sprintf("%d != %d", coords[0], test.normalized))
	}

	// test for crash
	for weight := float32(200); weight < 901; weight++ {
		font.NormalizeVariations([]float32{weight})
	}

	face := Face{Font: font}
	face.SetVariations([]Variation{{ot.MustNewTag("wght"), 206.}})
	tu.Assert(t, len(face.coords) == 1)
	tu.Assert(t, face.coords[0] == -16117)

	face.SetVariations(nil)
	tu.Assert(t, len(face.coords) == 0)

	font = loadFont(t, "common/NotoSansCJKjp-VF.otf")
	for _, test := range []struct {
		design     float32
		normalized VarCoord
	}{
		{200, 1310},
		{301, 2672},
		{400, 6390},
		{401.2, 6423},
	} {
		coords := font.NormalizeVariations([]float32{test.design})
		tu.AssertC(t, coords[0] == test.normalized, fmt.Sprintf("%d != %d", coords[0], test.normalized))
	}
}

func TestGlyphExtentsVar(t *testing.T) {
	font := loadFont(t, "common/SourceSans-VF-HVAR.ttf")

	for _, test := range []struct {
		coord    float32
		expected GlyphExtents
	}{
		{100, GlyphExtents{XBearing: 56, YBearing: 672, Width: 556, Height: -684}},
		{500, GlyphExtents{XBearing: 50, YBearing: 667, Width: 592, Height: -679}},
		{900, GlyphExtents{XBearing: 44, YBearing: 662, Width: 630, Height: -674}},
	} {
		coords := font.NormalizeVariations([]float32{test.coord})
		face := NewFace(font)
		face.coords = coords

		ext, _ := face.GlyphExtents(2)

		// the ref. values are extracted from harfbuzz, which round extents
		ext.XBearing = float32(math.Round(float64(ext.XBearing)))
		ext.YBearing = float32(math.Round(float64(ext.YBearing)))
		ext.Width = float32(math.Round(float64(ext.Width)))
		ext.Height = float32(math.Round(float64(ext.Height)))

		tu.Assert(t, ext == test.expected)
	}
}

func TestGetDefaultCoords(t *testing.T) {
	tf := fvar{
		{Tag: ot.MustNewTag("wght"), Minimum: 38, Default: 88, Maximum: 250},
		{Tag: ot.MustNewTag("wdth"), Minimum: 60, Default: 402, Maximum: 402},
		{Tag: ot.MustNewTag("opsz"), Minimum: 10, Default: 14, Maximum: 72},
	}

	vars := []Variation{
		{Tag: ot.MustNewTag("wdth"), Value: 60},
	}
	coords := tf.getDesignCoordsDefault(vars)
	tu.Assert(t, reflect.DeepEqual(coords, []float32{88, 60, 14}))
}

func TestNormalizeVar(t *testing.T) {
	tf := fvar{
		{Tag: ot.MustNewTag("wdth"), Minimum: 60, Default: 402, Maximum: 500},
	}

	vars := []Variation{
		{Tag: ot.MustNewTag("wdth"), Value: 60},
	}
	coords := tf.normalizeCoordinates(tf.getDesignCoordsDefault(vars))
	tu.Assert(t, reflect.DeepEqual(coords, []VarCoord{tables.NewCoord(-1)}))

	vars = []Variation{
		{Tag: ot.MustNewTag("wdth"), Value: 30},
	}
	coords = tf.normalizeCoordinates(tf.getDesignCoordsDefault(vars))
	tu.Assert(t, reflect.DeepEqual(coords, []VarCoord{tables.NewCoord(-1)}))

	vars = []Variation{
		{Tag: ot.MustNewTag("wdth"), Value: 700},
	}
	coords = tf.normalizeCoordinates(tf.getDesignCoordsDefault(vars))
	tu.Assert(t, reflect.DeepEqual(coords, []VarCoord{tables.NewCoord(1)}))
}

func TestAdvanceHVar(t *testing.T) {
	font := loadFont(t, "common/Commissioner-VF.ttf")
	coords := []VarCoord{-6553, 0, 13108, tables.NewCoord(1)}

	// 0 - 99 GIDs
	exps := [100]float32{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 8, 8, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6,
		1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 3.2, 3.2, 3.2, 0, 0,
	}
	tu.Assert(t, font.hvar != nil)
	for i, exp := range exps {
		got := font.hvar.AdvanceDelta(gID(i), coords)
		if math.Abs(float64(got-exp)) > 0.001 {
			t.Errorf("expected %f, got %f", exp, got)
		}
	}
}

func TestAdvanceNoHVar(t *testing.T) {
	font := loadFont(t, "toys/GVAR-no-HVAR.ttf")

	tu.Assert(t, len(font.fvar) == 2)

	vars := []Variation{
		{Tag: ot.MustNewTag("wght"), Value: 600},
		{Tag: ot.MustNewTag("wght"), Value: 80},
	}
	face := Face{Font: font}
	face.SetVariations(vars)

	// 0 - 14 GIDs
	exps := [15]float32{600, 1164, 1170, 813, 741, 1164, 1170, 813, 741, 270, 270, 0, 0, 0, 0}

	for i, exp := range exps {
		got := face.HorizontalAdvance(GID(i))
		tu.Assert(t, exp == got)
	}
}

func TestInvalidGVAR(t *testing.T) {
	// this file is build by subsetting the 'glyf' table
	// but keeping the variations tables
	f, err := os.Open("testdata/Selawik-VF-Subset.ttf")
	tu.AssertNoErr(t, err)
	defer f.Close()

	ld, err := ot.NewLoader(f)
	tu.AssertNoErr(t, err)

	head, _, _ := LoadHeadTable(ld, nil)
	raw, _ := ld.RawTable(ot.MustNewTag("maxp"))
	maxp, _, _ := tables.ParseMaxp(raw)
	raw, _ = ld.RawTable(ot.MustNewTag("glyf"))
	locaRaw, _ := ld.RawTable(ot.MustNewTag("loca"))
	loca, _ := tables.ParseLoca(locaRaw, int(maxp.NumGlyphs), head.IndexToLocFormat == 1)
	glyf, _ := tables.ParseGlyf(raw, loca)

	raw, err = ld.RawTable(ot.MustNewTag("gvar"))
	tu.AssertNoErr(t, err)
	gvar, _, err := tables.ParseGvar(raw)
	tu.AssertNoErr(t, err)
	// check that newGvar does not crash..
	_, err = newGvar(gvar, glyf)
	// ... and reports an error
	tu.Assert(t, err != nil)
}
