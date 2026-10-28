// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package cff

// code is adapted from golang.org/x/image/font/sfnt

import (
	"errors"
	"fmt"

	ps "github.com/OpenMellon/typesetting/font/cff/interpreter"
)

var topDictOperators = [2][]topDictOperator{
	// 1-byte operators.
	{
		0: {func(t *topDict, s *ps.Machine) error {
			t.version = s.ArgStack.Uint16()
			return nil
		}, +1 /*version*/},
		1: {func(t *topDict, s *ps.Machine) error {
			t.notice = s.ArgStack.Uint16()
			return nil
		}, +1 /*Notice*/},
		2: {func(t *topDict, s *ps.Machine) error {
			t.fullName = s.ArgStack.Uint16()
			return nil
		}, +1 /*FullName*/},
		3: {func(t *topDict, s *ps.Machine) error {
			t.familyName = s.ArgStack.Uint16()
			return nil
		}, +1 /*FamilyName*/},
		4: {func(t *topDict, s *ps.Machine) error {
			t.weight = s.ArgStack.Uint16()
			return nil
		}, +1 /*Weight*/},
		5:  {topDictNoOp, -1 /*FontBBox*/},
		13: {topDictNoOp, +1 /*UniqueID*/},
		14: {topDictNoOp, -1 /*XUID*/},
		15: {func(t *topDict, s *ps.Machine) error {
			t.charsetOffset = int32(s.ArgStack.Vals[s.ArgStack.Top-1])
			return nil
		}, +1 /*charset*/},
		16: {func(t *topDict, s *ps.Machine) error {
			t.encodingOffset = int32(s.ArgStack.Vals[s.ArgStack.Top-1])
			return nil
		}, +1 /*Encoding*/},
		17: {func(t *topDict, s *ps.Machine) error {
			t.charStringsOffset = int32(s.ArgStack.Vals[s.ArgStack.Top-1])
			return nil
		}, +1 /*CharStrings*/},
		18: {func(t *topDict, s *ps.Machine) error {
			t.privateDictLength = int32(s.ArgStack.Vals[s.ArgStack.Top-2])
			t.privateDictOffset = int32(s.ArgStack.Vals[s.ArgStack.Top-1])
			return nil
		}, +2 /*Private*/},
	},
	// 2-byte operators. The first byte is the escape byte.
	{
		0: {topDictNoOp, +1 /*Copyright*/},
		1: {func(t *topDict, s *ps.Machine) error {
			t.isFixedPitch = s.ArgStack.Vals[s.ArgStack.Top-1] == 1
			return nil
		}, +1 /*isFixedPitch*/},
		2: {func(t *topDict, s *ps.Machine) error {
			t.italicAngle = float32(s.ArgStack.Vals[s.ArgStack.Top-1])
			return nil
		}, +1 /*ItalicAngle*/},
		3: {func(t *topDict, s *ps.Machine) error {
			t.underlinePosition = float32(s.ArgStack.Vals[s.ArgStack.Top-1])
			return nil
		}, +1 /*UnderlinePosition*/},
		4: {func(t *topDict, s *ps.Machine) error {
			t.underlineThickness = float32(s.ArgStack.Vals[s.ArgStack.Top-1])
			return nil
		}, +1 /*UnderlineThickness*/},
		5: {topDictNoOp, +1 /*PaintType*/},
		6: {func(_ *topDict, i *ps.Machine) error {
			if version := int(i.ArgStack.Vals[i.ArgStack.Top-1]); version != 2 {
				return fmt.Errorf("charstring type %d not supported", version)
			}
			return nil
		}, +1 /*CharstringType*/},
		7:  {topDictNoOp, -1 /*FontMatrix*/},
		8:  {topDictNoOp, +1 /*StrokeWidth*/},
		20: {topDictNoOp, +1 /*SyntheticBase*/},
		21: {topDictNoOp, +1 /*PostScript*/},
		22: {topDictNoOp, +1 /*BaseFontName*/},
		23: {topDictNoOp, -2 /*BaseFontBlend*/},
		30: {func(t *topDict, _ *ps.Machine) error {
			t.isCIDFont = true
			return nil
		}, +3 /*ROS*/},
		31: {topDictNoOp, +1 /*CIDFontVersion*/},
		32: {topDictNoOp, +1 /*CIDFontRevision*/},
		33: {topDictNoOp, +1 /*CIDFontType*/},
		34: {topDictNoOp, +1 /*CIDCount*/},
		35: {topDictNoOp, +1 /*UIDBase*/},
		36: {func(t *topDict, s *ps.Machine) error {
			t.fdArray = int32(s.ArgStack.Vals[s.ArgStack.Top-1])
			return nil
		}, +1 /*FDArray*/},
		37: {func(t *topDict, s *ps.Machine) error {
			t.fdSelect = int32(s.ArgStack.Vals[s.ArgStack.Top-1])
			return nil
		}, +1 /*FDSelect*/},
		38: {func(t *topDict, s *ps.Machine) error {
			t.cidFontName = s.ArgStack.Uint16()
			return nil
		}, +1 /*FontName*/},
	},
}

// privateDict contains fields specific to the Private DICT context.
type privateDict struct {
	subrsOffset                  int32
	defaultWidthX, nominalWidthX float64
}

func (privateDict) Context() ps.Context { return ps.PrivateDict }

// The Private DICT operators are defined by 5176.CFF.pdf Table 23 "Private
// DICT Operators".
func (priv *privateDict) Apply(state *ps.Machine, op ps.Operator) error {
	if !op.IsEscaped { // 1-byte operators.
		switch op.Operator {
		case 6, 7, 8, 9: // "BlueValues" "OtherBlues" "FamilyBlues" "FamilyOtherBlues"
			return state.ArgStack.PopN(-2)
		case 10, 11: // "StdHW" "StdVW"
			return state.ArgStack.PopN(1)
		case 20: // "defaultWidthX"
			if state.ArgStack.Top < 1 {
				return errors.New("invalid stack size for 'defaultWidthX' in private Dict charstring")
			}
			priv.defaultWidthX = state.ArgStack.Vals[state.ArgStack.Top-1]
			return state.ArgStack.PopN(1)
		case 21: // "nominalWidthX"
			if state.ArgStack.Top < 1 {
				return errors.New("invalid stack size for 'nominalWidthX' in private Dict charstring")
			}
			priv.nominalWidthX = state.ArgStack.Vals[state.ArgStack.Top-1]
			return state.ArgStack.PopN(1)
		case 19: // "Subrs" pop 1
			if state.ArgStack.Top < 1 {
				return errors.New("invalid stack size for 'subrs' in private Dict charstring")
			}
			priv.subrsOffset = int32(state.ArgStack.Vals[state.ArgStack.Top-1])
			return state.ArgStack.PopN(1)
		}
	} else { // 2-byte operators. The first byte is the escape byte.
		switch op.Operator {
		case 9, 10, 11, 14, 17, 18, 19: // "BlueScale" "BlueShift" "BlueFuzz" "ForceBold" "LanguageGroup" "ExpansionFactor" "initialRandomSeed"
			return state.ArgStack.PopN(1)
		case 12, 13: //  "StemSnapH"  "StemSnapV"
			return state.ArgStack.PopN(-2)
		}
	}
	return errors.New("invalid operand in private Dict charstring")
}
