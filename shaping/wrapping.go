package shaping

import (
	"sort"

	"github.com/OpenMellon/typesetting/di"
	"github.com/OpenMellon/typesetting/segmenter"
)

// glyphIndex is the index in a Glyph slice
type glyphIndex = int

// mapRunesToClusterIndices
// returns a slice that maps rune indicies in the text to the index of the
// first glyph in the glyph cluster containing that rune in the shaped text.
// The indicies are relative to the region of runes covered by the input run.
// To translate an absolute rune index in text into a rune index into the returned
// mapping, subtract run.Runes.Offset first. If the provided buf is large enough to
// hold the return value, it will be used instead of allocating a new slice.
func mapRunesToClusterIndices(dir di.Direction, runes Range, glyphs []Glyph, buf []glyphIndex) []glyphIndex {
	if runes.Count <= 0 {
		return nil
	}
	var mapping []glyphIndex
	if cap(buf) >= runes.Count {
		mapping = buf[:runes.Count]
	} else {
		mapping = make([]glyphIndex, runes.Count)
	}
	glyphCursor := 0
	rtl := dir.Progression() == di.TowardTopLeft
	if rtl {
		glyphCursor = len(glyphs) - 1
	}
	// off tracks the offset position of the glyphs from the first rune of the
	// shaped text. This must be subtracted from all cluster indicies in order to
	// normalize them into the range [0,runes.Count).
	off := runes.Offset
	for i := 0; i < runes.Count; i++ {
		for glyphCursor >= 0 && glyphCursor < len(glyphs) &&
			((rtl && glyphs[glyphCursor].ClusterIndex-off <= i) ||
				(!rtl && glyphs[glyphCursor].ClusterIndex-off < i)) {
			if rtl {
				glyphCursor--
			} else {
				glyphCursor++
			}
		}
		if rtl {
			glyphCursor++
		} else if (glyphCursor >= 0 && glyphCursor < len(glyphs) &&
			glyphs[glyphCursor].ClusterIndex-off > i) ||
			(glyphCursor == len(glyphs) && len(glyphs) > 1) {
			glyphCursor--
			targetClusterIndex := glyphs[glyphCursor].ClusterIndex - off
			for glyphCursor-1 >= 0 && glyphs[glyphCursor-1].ClusterIndex-off == targetClusterIndex {
				glyphCursor--
			}
		}
		if glyphCursor < 0 {
			glyphCursor = 0
		} else if glyphCursor >= len(glyphs) {
			glyphCursor = len(glyphs) - 1
		}
		mapping[i] = glyphCursor
	}
	return mapping
}

// mapRuneToClusterIndex finds the lowest-index glyph for the glyph cluster contiaining the rune
// at runeIdx in the source text. It uses a binary search of the glyphs in order to achieve this.
// It is equivalent to using mapRunesToClusterIndices on only a single rune index, and is thus
// more efficient for single lookups while being less efficient for runs which require many
// lookups anyway.
func mapRuneToClusterIndex(dir di.Direction, runes Range, glyphs []Glyph, runeIdx int) glyphIndex {
	var index int
	rtl := dir.Progression() == di.TowardTopLeft
	if !rtl {
		index = sort.Search(len(glyphs), func(index int) bool {
			return glyphs[index].ClusterIndex-runes.Offset > runeIdx
		})
	} else {
		index = sort.Search(len(glyphs), func(index int) bool {
			return glyphs[index].ClusterIndex-runes.Offset < runeIdx
		})
	}
	if index < 1 {
		return 0
	}
	cluster := glyphs[index-1].ClusterIndex
	if rtl && cluster-runes.Offset > runeIdx {
		return index
	}
	for index-1 >= 0 && glyphs[index-1].ClusterIndex == cluster {
		index--
	}
	return index
}

func mapRunesToClusterIndices2(dir di.Direction, runes Range, glyphs []Glyph, buf []glyphIndex) []glyphIndex {
	if runes.Count <= 0 {
		return nil
	}
	var mapping []glyphIndex
	if cap(buf) >= runes.Count {
		mapping = buf[:runes.Count]
	} else {
		mapping = make([]glyphIndex, runes.Count)
	}

	rtl := dir.Progression() == di.TowardTopLeft
	if rtl {
		for gIdx := len(glyphs) - 1; gIdx >= 0; gIdx-- {
			cluster := glyphs[gIdx].ClusterIndex
			clusterEnd := gIdx
			for gIdx-1 >= 0 && glyphs[gIdx-1].ClusterIndex == cluster {
				gIdx--
				clusterEnd = gIdx
			}
			var nextCluster int
			if gIdx-1 >= 0 {
				nextCluster = glyphs[gIdx-1].ClusterIndex
			} else {
				nextCluster = runes.Count + runes.Offset
			}
			runesInCluster := nextCluster - cluster
			clusterOffset := cluster - runes.Offset
			for i := clusterOffset; i <= runesInCluster+clusterOffset && i < len(mapping); i++ {
				mapping[i] = clusterEnd
			}
		}
	} else {
		for gIdx := 0; gIdx < len(glyphs); gIdx++ {
			cluster := glyphs[gIdx].ClusterIndex
			clusterStart := gIdx
			for gIdx+1 < len(glyphs) && glyphs[gIdx+1].ClusterIndex == cluster {
				gIdx++
			}
			var nextCluster int
			if gIdx+1 < len(glyphs) {
				nextCluster = glyphs[gIdx+1].ClusterIndex
			} else {
				nextCluster = runes.Count + runes.Offset
			}
			runesInCluster := nextCluster - cluster
			clusterOffset := cluster - runes.Offset
			for i := clusterOffset; i <= runesInCluster+clusterOffset && i < len(mapping); i++ {
				mapping[i] = clusterStart
			}
		}
	}
	return mapping
}

func mapRunesToClusterIndices3(dir di.Direction, runes Range, glyphs []Glyph, buf []glyphIndex) []glyphIndex {
	if runes.Count <= 0 {
		return nil
	}
	var mapping []glyphIndex
	if cap(buf) >= runes.Count {
		mapping = buf[:runes.Count]
	} else {
		mapping = make([]glyphIndex, runes.Count)
	}

	rtl := dir.Progression() == di.TowardTopLeft
	if rtl {
		for gIdx := len(glyphs) - 1; gIdx >= 0; {
			glyph := &glyphs[gIdx]
			// go to the start of the cluster
			gIdx -= (glyph.GlyphCount - 1)
			clusterStart := glyph.ClusterIndex - runes.Offset // map back to [0;runes.Count[
			clusterEnd := glyph.RuneCount + clusterStart
			for i := clusterStart; i <= clusterEnd && i < len(mapping); i++ {
				mapping[i] = gIdx
			}
			// go to the next cluster
			gIdx--
		}
	} else {
		for gIdx := 0; gIdx < len(glyphs); {
			glyph := &glyphs[gIdx]
			clusterStart := glyph.ClusterIndex - runes.Offset // map back to [0;runes.Count[
			clusterEnd := glyph.RuneCount + clusterStart
			for i := clusterStart; i <= clusterEnd && i < len(mapping); i++ {
				mapping[i] = gIdx
			}
			// go to the next cluster
			gIdx += glyph.GlyphCount
		}
	}
	return mapping
}

// inclusiveGlyphRange returns the inclusive range of runes and glyphs matching
// the provided start and breakAfter rune positions.
// runeToGlyph must be a valid mapping from the rune representation to the
// glyph reprsentation produced by mapRunesToClusterIndices.
// numGlyphs is the number of glyphs in the output representing the runes
// under consideration.
func inclusiveGlyphRange(dir di.Direction, start, breakAfter int, runeToGlyph []int, numGlyphs int) (glyphStart, glyphEnd glyphIndex) {
	rtl := dir.Progression() == di.TowardTopLeft
	if rtl {
		glyphStart = runeToGlyph[breakAfter]
		if start-1 >= 0 {
			glyphEnd = runeToGlyph[start-1] - 1
		} else {
			glyphEnd = numGlyphs - 1
		}
	} else {
		glyphStart = runeToGlyph[start]
		if breakAfter+1 < len(runeToGlyph) {
			glyphEnd = runeToGlyph[breakAfter+1] - 1
		} else {
			glyphEnd = numGlyphs - 1
		}
	}
	return
}

// cutRun returns the sub-run of run containing glyphs corresponding to the provided
// _inclusive_ rune range.
// if [trimStart] is true, the leading letter spacing is removed
func cutRun(run Output, mapping []glyphIndex, startRune, endRune int, trimStart bool) Output {
	// Convert the rune range of interest into an inclusive range within the
	// current run's runes.
	runeStart := startRune - run.Runes.Offset
	runeEnd := endRune - run.Runes.Offset
	if runeStart < 0 {
		// If the start location is prior to the run of shaped text under consideration,
		// just work from the beginning of this run.
		runeStart = 0
	}
	if runeEnd >= len(mapping) {
		// If the break location is after the entire run of shaped text,
		// keep through the end of the run.
		runeEnd = len(mapping) - 1
	}
	glyphStart, glyphEnd := inclusiveGlyphRange(run.Direction, runeStart, runeEnd, mapping, len(run.Glyphs))

	// Construct a run out of the inclusive glyph range.
	run.Glyphs = run.Glyphs[glyphStart : glyphEnd+1]
	run.Runes.Count = runeEnd - runeStart + 1
	run.Runes.Offset = run.Runes.Offset + runeStart
	if trimStart {
		run.trimStartLetterSpacing()
	}
	run.RecomputeAdvance()
	return run
}

// breakOption represets a location within the rune slice at which
// it may be safe to break a line of text.
type breakOption struct {
	// breakAtRune is the index at which it is safe to break.
	breakAtRune int
	// required indicates that the break option is mandatory.
	required bool
}

// isValid returns whether a given option violates shaping rules (like breaking
// a shaped text cluster).
func (option breakOption) isValid(runeToGlyph []int, out Output) bool {
	breakAfter := option.breakAtRune - out.Runes.Offset
	nextRune := breakAfter + 1
	if nextRune < len(runeToGlyph) && breakAfter >= 0 {
		// Check if this break is valid.
		gIdx := runeToGlyph[breakAfter]
		g2Idx := runeToGlyph[nextRune]
		if gIdx >= len(out.Glyphs) || g2Idx >= len(out.Glyphs) {
			return false
		}
		cIdx := out.Glyphs[gIdx].ClusterIndex
		c2Idx := out.Glyphs[g2Idx].ClusterIndex
		if cIdx == c2Idx {
			// This break is within a harfbuzz cluster, and is
			// therefore invalid.
			return false
		}
	}
	return true
}

// breaker generates line breaking candidates for a text.
type breaker struct {
	wordSegmenter     *segmenter.LineIterator
	graphemeSegmenter *segmenter.GraphemeIterator
	totalRunes        int
	// unusedWordBreak is a break requested from the breaker in a previous iteration
	// but which was not chosen as the line ending. Subsequent invocations of
	// WrapLine should start with this break.
	unusedWordBreak breakOption
	// previousWordBreak tracks the previous line breaking candidate, if any. It is
	// used to identify the range of runes between the previous and current
	// candidate.
	previousWordBreak breakOption
	// isUnusedWord indicates that the unusedBreak field is valid.
	isUnusedWord        bool
	unusedGraphemeBreak breakOption
	isUnusedGrapheme    bool
}

// newBreaker returns a breaker initialized to break the provided text.
func newBreaker(seg *segmenter.Segmenter, text []rune) *breaker {
	seg.Init(text)
	br := &breaker{
		wordSegmenter:     seg.LineIterator(),
		graphemeSegmenter: seg.GraphemeIterator(),
		totalRunes:        len(text),
	}
	return br
}

// nextWordRaw returns a naive break candidate on a uax#14 boundary which may be invalid.
func (b *breaker) nextWordRaw() (option breakOption, ok bool) {
	if b.wordSegmenter.Next() {
		currentSegment := b.wordSegmenter.Line()
		// Note : we dont use penalties for Mandatory Breaks so far,
		// we could add it with currentSegment.IsMandatoryBreak
		breakAtRune := currentSegment.Offset + len(currentSegment.Text) - 1
		option := breakOption{
			breakAtRune: breakAtRune,
			// Don't treat the EOF line break as special. We implicitly always break after
			// the end of text input anyway.
			required: currentSegment.IsMandatoryBreak && breakAtRune != b.totalRunes-1,
		}
		return option, true
	}
	// Unicode rules impose to always break at the end
	return breakOption{}, false
}

// nextGraphemeRaw returns a naive break candidate on a uax#29 boundary which may be invalid.
func (b *breaker) nextGraphemeRaw() (option breakOption, ok bool) {
	if b.graphemeSegmenter.Next() {
		currentSegment := b.graphemeSegmenter.Grapheme()
		// Note : we dont use penalties for Mandatory Breaks so far,
		// we could add it with currentSegment.IsMandatoryBreak
		option := breakOption{
			breakAtRune: currentSegment.Offset + len(currentSegment.Text) - 1,
		}
		return option, true
	}
	// Unicode rules impose to always break at the end
	return breakOption{}, false
}

// nextWordBreak returns the next rune offset at which the line can be broken
// on a UAX#14 boundary if any. If it returns false, there are no more candidates.
func (l *breaker) nextWordBreak() (breakOption, bool) {
	var option breakOption
	if l.isUnusedWord {
		option = l.unusedWordBreak
		l.isUnusedWord = false
	} else {
		var breakOk bool
		option, breakOk = l.nextWordRaw()
		if !breakOk {
			return option, false
		}
		l.previousWordBreak = l.unusedWordBreak
		l.unusedWordBreak = option
	}
	return option, true
}

func (l *breaker) markWordOptionUnused() {
	l.isUnusedWord = true
}

// nextGraphemeBreak returns the next grapheme cluster boundary break between
// the previous and current word boundary, if any. If it returns false, there are no
// more candidates between the previous and current word boundaries.
func (l *breaker) nextGraphemeBreak() (breakOption, bool) {
	for {
		var (
			option  breakOption
			breakOk bool
		)
		if l.isUnusedGrapheme {
			l.isUnusedGrapheme = false
			option = l.unusedGraphemeBreak
			breakOk = true
		} else {
			option, breakOk = l.nextGraphemeRaw()
		}
		if !breakOk {
			return option, false
		}
		// We don't want to consider the previous word break position in general, as it has already
		// been tried. The one exception to this is when we're iterating the very first time, in
		// which case the previousWordBreak will have its zero value and we still want to consider
		// breaking after the first rune if there's a grapheme cluseter boundary there.
		if option.breakAtRune <= l.previousWordBreak.breakAtRune && l.previousWordBreak.breakAtRune > 0 {
			continue
		}
		l.unusedGraphemeBreak = option
		if option.breakAtRune > l.unusedWordBreak.breakAtRune {
			// We've walked the grapheme iterator past the end of the line wrapping
			// candidate, so mark that we may need to re-check this break option
			// when evaluating the next segment.
			l.isUnusedGrapheme = true
			return option, false
		}
		return option, true
	}
}

func (l *breaker) markGraphemeOptionUnused() {
	l.isUnusedGrapheme = true
}

// Range indicates the location of a sequence of elements within a longer slice.
type Range struct {
	Offset int
	Count  int
}

// Line holds runs of shaped text wrapped onto a single line. All the contained
// Output should be displayed sequentially on one line.
type Line []Output

// WrapConfig provides line-wrapper settings.
type WrapConfig struct {
	// Direction describes the text layout of the overall paragraph, rather than
	// individual runs of text. This is used to compute the correct visual order of
	// bidirectional text runs.
	Direction di.Direction
	// TruncateAfterLines is the number of lines of text to allow before truncating
	// the text. A value of zero means no limit.
	TruncateAfterLines int
	// Truncator, if provided, will be inserted at the end of a truncated line. This
	// feature is only active if TruncateAfterLines is nonzero. See the documentation
	// for [LineWrapper.WrapNextLine] for details about how this works.
	Truncator Output
	// TextContinues indicates that the paragraph wrapped by this config is not the
	// final paragraph in the text. This alters text truncation when filling the
	// final line permitted by TruncateAfterLines. If the text of this paragraph
	// does fit entirely on TruncateAfterLines, normally the truncator symbol would
	// not be inserted. However, if the overall body of text continues beyond this
	// paragraph (indicated by TextContinues), the truncator should still be inserted
	// to indicate that further paragraphs of text were truncated. This field has
	// no effect if TruncateAfterLines is zero.
	TextContinues bool
	// BreakPolicy determines under what circumstances the wrapper will consider
	// breaking in between UAX#14 line breaking candidates, or "within words" in
	// many scripts.
	BreakPolicy LineBreakPolicy
	// DisableTrailingWhitespaceTrim turns off a feature that automatically sets the
	// advance of trailing whitespace on a line to zero. In display contexts, you
	// usually want this feature enabled, but for text editors it is frequently
	// desirable to allow trailing whitespace to occupy space itself.
	DisableTrailingWhitespaceTrim bool
}

// LineBreakPolicy specifies when considering a line break within a "word" or UAX#14
// segment is allowed.
type LineBreakPolicy uint8

const (
	// WhenNecessary means that lines will only be broken within words when the word
	// cannot fit on the next line by itself or during truncation to preserve as much
	// of the final word as possible.
	WhenNecessary LineBreakPolicy = iota
	// Never means that words will never be broken internally, allowing them to exceed
	// the specified maxWidth.
	Never
	// Always means that lines will always choose to break within words if it means that
	// more text can fit on the line.
	Always
)

func (l LineBreakPolicy) String() string {
	switch l {
	case WhenNecessary:
		return "WhenNecessary"
	case Never:
		return "Never"
	case Always:
		return "Always"
	default:
		return "unknown"
	}
}

// WithTruncator returns a copy of WrapConfig with the Truncator field set to the
// result of shaping input with shaper.
func (w WrapConfig) WithTruncator(shaper Shaper, input Input) WrapConfig {
	w.Truncator = shaper.Shape(input)
	return w
}
