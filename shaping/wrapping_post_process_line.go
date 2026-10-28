package shaping

import (
	"math"

	"github.com/OpenMellon/typesetting/di"
	"golang.org/x/image/math/fixed"
)

func (l *LineWrapper) postProcessLine(finalLine Line, done bool) (WrappedLine, bool) {
	var trimmed fixed.Int26_6
	if len(finalLine) > 0 {
		computeBidiOrdering(l.config.Direction, finalLine)
		if !l.config.DisableTrailingWhitespaceTrim {
			// Here we find the last visual run in the line.
			goalIdx := len(finalLine) - 1
			if l.config.Direction.Progression() == di.TowardTopLeft {
				goalIdx = 0
			}
			for logicalIdx, run := range finalLine {
				if run.VisualIndex == int32(goalIdx) {
					goalIdx = logicalIdx
					break
				}
			}
			finalVisualRun := &finalLine[goalIdx]

			// This next block locates the first/last visual glyph on the line and
			// zeroes its advance if it is whitespace.
			if L := len(finalVisualRun.Glyphs); L > 0 {
				var finalVisualGlyph *Glyph
				if l.config.Direction.Progression() == di.FromTopLeft {
					finalVisualGlyph = &finalVisualRun.Glyphs[L-1]
				} else {
					finalVisualGlyph = &finalVisualRun.Glyphs[0]
				}

				if finalVisualRun.Direction.IsVertical() {
					if finalVisualGlyph.Height == 0 {
						finalVisualGlyph.YAdvance = 0
						finalVisualGlyph.Advance = 0
					}
				} else { // horizontal
					if finalVisualGlyph.Width == 0 {
						finalVisualGlyph.XAdvance = 0
						finalVisualGlyph.Advance = 0
					}
				}
				beforeTrim := finalVisualRun.Advance
				finalVisualRun.RecomputeAdvance()
				trimmed = beforeTrim - finalVisualRun.Advance
			}
		}

		finalLogicalRun := finalLine[len(finalLine)-1]
		// Update the start position of the next line.
		l.lineStartRune = finalLogicalRun.Runes.Count + finalLogicalRun.Runes.Offset
	}

	// Check whether we've exhausted the text.
	done = done || l.lineStartRune >= l.breaker.totalRunes

	// Implement truncation if needed.
	truncated := 0
	if l.truncating {
		l.config.TruncateAfterLines--
		insertTruncator := false
		if l.config.TruncateAfterLines == 0 {
			done = true
			truncated = l.breaker.totalRunes - l.lineStartRune
			insertTruncator = truncated > 0 || l.config.TextContinues
		}
		if insertTruncator {
			truncator := l.config.Truncator
			truncator.Runes.Count = truncated
			truncator.Runes.Offset = l.lineStartRune
			finalLine = append(finalLine, truncator)
			// We've just modified the line, we need to recompute the bidi ordering.
			computeBidiOrdering(l.config.Direction, finalLine)
		}
	}

	// Mark the paragraph as complete if needed.
	if done {
		l.more = false
	}

	return WrappedLine{finalLine, truncated, l.lineStartRune, trimmed}, done
}

// WrapNextLine wraps the shaped glyphs of a paragraph to a particular max width.
// It is meant to be called iteratively to wrap each line, allowing lines to
// be wrapped to different widths within the same paragraph. When done is true,
// subsequent calls to WrapNextLine (without calling Prepare) will return a nil line.
//
// The returned line is only valid until the next call to
// [*LineWrapper.Prepare] or [*LineWrapper.WrapParagraph].
//
// If the LineWrapper's [WrapConfig].TruncateAfterLines
// is non-zero, the final line of text returned by successive calls to WrapNextLine
// may be truncated. The quantity of runes truncated by wrapping the line is
// returned in [WrappedLine].Truncated. If this is non-zero AND the LineWrapper's
// [WrapConfig].Truncator is set, the final line will end with an extra [Output].
// [Output]s before the final [Output] will represent the input runes that are still
// visible before truncation, and the final [Output] will be a copy of the Truncator
// with its Runes.Count set to the quantity of runes truncated during line wrapping.
//
// See also [WrapNextLineF] which supports a decimal [maxWidth].
func (l *LineWrapper) WrapNextLine(maxWidth int) (out WrappedLine, done bool) {
	maxFixed := math.MaxInt32 >> 6
	if maxWidth > maxFixed {
		maxWidth = maxFixed
	}
	return l.WrapNextLineF(fixed.I(maxWidth))
}

// WrapNextLineF is the same as [WrapNextLine], but accepts a non integer [maxWidth].
func (l *LineWrapper) WrapNextLineF(maxWidth fixed.Int26_6) (out WrappedLine, done bool) {
	// If we've already finished the paragraph, don't do any more work.
	if !l.more {
		return WrappedLine{NextLine: l.lineStartRune}, true
	}

	defer func() {
		out, done = l.postProcessLine(out.Line, done)
	}()

	// If the iterator is empty, return early.
	_, _, hasFirst := l.glyphRuns.Peek()
	if !hasFirst {
		return WrappedLine{}, true
	}
	l.scratch.startLine()

	config := lineConfig{
		truncating:        l.config.TruncateAfterLines == 1,
		maxWidth:          maxWidth,
		truncatedMaxWidth: maxWidth - l.config.Truncator.Advance,
	}
	done = l.wrapNextLine(config)
	finalLine := l.scratch.finalizeBest()
	return WrappedLine{Line: finalLine}, done
}

// checkpoint captures both the current candidate line and the corresponding run iteration
// state. These can be restored together by calling restore().
func (l *LineWrapper) checkpoint() {
	l.scratch.candidateSave()
	l.glyphRuns.Save()
}

// restore resets the current candidate line and corresponding run iteration state to the
// values at the last call to checkpoint().
func (l *LineWrapper) restore() {
	l.scratch.candidateRestore()
	l.glyphRuns.Restore()
}

// wrapNextLine iteratively processes line breaking candidates, building a line within the
// wrapper's scratch [WrapBuffer]. It returns whether the paragraph is finished once it has
// successfully built a line.
func (l *LineWrapper) wrapNextLine(config lineConfig) (done bool) {
	for {
		l.checkpoint()
		option, ok := l.breaker.nextWordBreak()
		if !ok {
			break
		}
		switch result, candidateRun := l.processBreakOption(option, config); result {
		case breakInvalid:
			l.restore()
			continue
		case fits:
			l.scratch.markCandidateBest(candidateRun)
			if option.required {
				return false
			}
			continue
		case endLine:
			// Found a valid line ending the text, append the candidateRun and use it.
			l.scratch.markCandidateBest(candidateRun)
			return true
		case truncated:
			// The candidateRun does not fit.
			if !l.scratch.hasBest() {
				l.scratch.markCandidateBest()
			}
			if l.config.BreakPolicy == Never {
				return true
			}
			// Fall through to try grapheme breaking.
		case newLineBeforeBreak:
			l.restore()
			// We found a valid line that didn't use this break, so mark that it can be
			// reused on the next iteration.
			l.breaker.markWordOptionUnused()
			if l.config.BreakPolicy == Never || (l.config.BreakPolicy == WhenNecessary && !config.truncating) {
				return false
			}
			// Fall through to try grapheme breaking.
		case cannotFit:
			if l.config.BreakPolicy == Never {
				if config.truncating {
					return true
				}
				l.scratch.markCandidateBest(candidateRun)
				return false
			}
			// Fall through to try grapheme breaking.
		}
		// Ensure that the grapheme breaking has access to
		// all runs we already tried in the iterator.
		l.restore()
		// segment using UAX#29 grapheme clustering here and try
		// breaking again using only those boundaries to find a viable break in cases
		// where no UAX#14 breaks were viable above.
		for {
			l.checkpoint()
			option, ok := l.breaker.nextGraphemeBreak()
			if !ok {
				break
			}
			switch result, candidateRun := l.processBreakOption(option, config); result {
			case breakInvalid:
				l.restore()
				continue
			case fits:
				// If we found at least one viable line candidate, we aren't using the word break option.
				l.scratch.markCandidateBest(candidateRun)
				l.breaker.markWordOptionUnused()
				continue
			case endLine:
				l.scratch.markCandidateBest(candidateRun)
				return true
			case truncated:
				if !l.scratch.hasBest() {
					l.scratch.markCandidateBest()
				}
				return true
			case newLineBeforeBreak:
				l.restore()
				// If we found at least one viable line candidate, we aren't using the word break option.
				l.breaker.markWordOptionUnused()
				l.breaker.markGraphemeOptionUnused()
				return false
			case cannotFit:
				if config.truncating {
					// Don't append the candidate grapheme to leave as much space as possible for the
					// truncator.
					return true
				}
				// If no graphemes fit, we should still use one so that the line contains something. Maybe
				// the next grapheme will fit on the next line.
				l.scratch.markCandidateBest(candidateRun)
				l.breaker.markWordOptionUnused()
				return false
			}
		}
		return false
	}
	return true
}

type processBreakResult uint8

const (
	// breakInvalid indicates that the provided break candidate is incompatible with the shaped
	// text and should be skipped.
	breakInvalid processBreakResult = iota
	// endLine indicates that the candidate fits on the line and terminates the text.
	endLine
	// truncated indicates that the candidate does not fit on the line and that truncation needs to
	// occur.
	truncated
	// newLineBeforeBreak indicates that the candidate contains a valid line, but the latest break
	// option could not fit.
	newLineBeforeBreak
	// fits indicates that the text up to the break option fit within the line and that another break
	// option can be attempted.
	fits
	// cannotFit indicates that this is the first break option on the line, and that even this option cannot
	// fit within the space available. When cannotFit is returned, the scratch buffer's candidate will contain
	// the run that cannot fit, but it will not be committed as the best option. The choice of how to handle
	// this is left to higher-level logic.
	cannotFit
)

// processBreakOption evaluates whether the provided breakOption can fit onto the current line wrapping line.
func (l *LineWrapper) processBreakOption(option breakOption, config lineConfig) (processBreakResult, Output) {
	// Discard break options on previous lines.
	if option.breakAtRune < l.lineStartRune {
		return breakInvalid, Output{}
	}

	// Fill candidate line with runs until the run containing the break option.
	l.fillUntil(l.glyphRuns, option)

	currRunIndex, run, _ := l.glyphRuns.Peek()
	l.mapper.mapRun(currRunIndex, run)
	if !option.isValid(l.mapper.mapping, run) {
		// Reject invalid line break candidate and acquire a new one.
		return breakInvalid, Output{}
	}
	isFirstInLine := l.scratch.candidateLen() == 0
	candidateRun := cutRun(run, l.mapper.mapping, l.lineStartRune, option.breakAtRune, isFirstInLine)
	candidateLineWidth := candidateRun.advanceSpaceAware(l.config.Direction) + l.scratch.candidateAdvance()
	if candidateLineWidth > config.maxWidth {
		// The run doesn't fit on the line.
		if !l.scratch.hasBest() {
			return cannotFit, candidateRun
		} else {
			return newLineBeforeBreak, candidateRun
		}
	} else if config.truncating && candidateLineWidth > config.truncatedMaxWidth {
		// The run would not fit if truncated.
		finalRunRune := candidateRun.Runes.Count + candidateRun.Runes.Offset
		if finalRunRune == l.breaker.totalRunes && !l.config.TextContinues {
			// The run contains the entire end of the text, so no truncation is
			// necessary.
			return endLine, candidateRun
		}
		// We must truncate the line in order to show it.
		return truncated, candidateRun
	} else {
		// The run does fit on the line. Commit this line as the best known
		// line, but keep lineCandidate unmodified so that later break
		// options can be attempted to see if a more optimal solution is
		// available.
		return fits, candidateRun
	}
}
