package unicodedata

import (
	"unicode"
)

// used as reference in benchmark
func lookupCombiningClass(ch rune) uint8 {
	for i, t := range combiningClasses {
		if t == nil {
			continue
		}
		if unicode.Is(t, ch) {
			return uint8(i)
		}
	}
	return 0
}

// used as reference in benchmark
func lookupMirrorChar(ch rune) rune {
	m, ok := mirroring[ch]
	if !ok {
		m = ch
	}
	return m
}

// used as reference in benchmark
func decompose(ab rune) (a, b rune, ok bool) {
	if a, b, ok = decomposeHangul(ab); ok {
		return a, b, true
	}

	// Check if it's a single-character decomposition.
	if m1, ok := decompose1[ab]; ok {
		return m1, 0, true
	}
	if m2, ok := decompose2[ab]; ok {
		return m2[0], m2[1], true
	}
	return ab, 0, false
}

// used as reference in benchmark
func compose_(a, b rune) (rune, bool) {
	// Hangul is handled algorithmically.
	if ab, ok := composeHangul(a, b); ok {
		return ab, true
	}
	u := compose[[2]rune{a, b}]
	return u, u != 0
}

// used as reference in benchmark
func lookupIndicConjunctBreak(r rune) IndicConjunctBreak {
	if unicode.Is(indicCBLinker, r) {
		return ICBLinker
	} else if unicode.Is(indicCBConsonant, r) {
		return ICBConsonant
	} else if unicode.Is(indicCBExtend, r) {
		return ICBExtend
	} else {
		return 0
	}
}

// used as reference in benchmark
func lookupGraphemeBreak(ch rune) *unicode.RangeTable {
	// a lot of runes do not have a grapheme break property :
	// avoid testing all the graphemeBreaks classes for them
	if !unicode.Is(graphemeBreakAll, ch) {
		return nil
	}
	for _, class := range graphemeBreaks {
		if unicode.Is(class, ch) {
			return class
		}
	}
	return nil
}

// used as reference in benchmark
func lookupWordBreak(ch rune) *unicode.RangeTable {
	// a lot of runes do not have a word break property :
	// avoid testing all the wordBreaks classes for them
	if !unicode.Is(wordBreakAll, ch) {
		return nil
	}
	for _, class := range wordBreaks {
		if unicode.Is(class, ch) {
			return class
		}
	}
	return nil
}

// used as reference in benchmark
func lookupLineBreak(ch rune) *unicode.RangeTable {
	for _, class := range lineBreaks {
		if unicode.Is(class, ch) {
			return class
		}
	}
	return nil
}
