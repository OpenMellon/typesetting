package shaping

import (
	"math"

	"github.com/OpenMellon/typesetting/di"
	"github.com/OpenMellon/typesetting/segmenter"
	"golang.org/x/image/math/fixed"
)

// runMapper efficiently maps a run to glyph clusters.
type runMapper struct {
	// valid indicates that the mapping field is populated.
	valid bool
	// runIdx is the index of the mapped run within glyphRuns.
	runIdx int
	// mapping holds the rune->glyph mapping for the run at index mappedRun within
	// glyphRuns.
	mapping []glyphIndex
}

// mapRun updates the mapping field to be valid for the given run. It will skip the mapping
// operation if the provided runIdx is equal to the runIdx of the previous call, as the
// current mapping value is already correct.
func (r *runMapper) mapRun(runIdx int, run Output) {
	if r.runIdx != runIdx || !r.valid {
		r.mapping = mapRunesToClusterIndices3(run.Direction, run.Runes, run.Glyphs, r.mapping)
		r.runIdx = runIdx
		r.valid = true
	}
}

// RunIterator defines a type that can incrementally provide shaped text.
type RunIterator interface {
	// Next returns the next run in the iterator, if any. If there is a next run,
	// its index, content, and true will be returned, and the iterator will advance
	// to the following element. Otherwise Next returns an undefined index, an empty
	// Output, and false.
	Next() (index int, run Output, isValid bool)
	// Peek returns the same thing Next() would, but does not advance the iterator (so the
	// next call to Next() will return the same thing).
	Peek() (index int, run Output, isValid bool)
	// Save marks the current iterator position such that the iterator can return to it later
	// when Restore() is called. Only one position may be saved at a time, with subsequent
	// calls to Save() overriding the current value.
	Save()
	// Restore resets the iteration state to the most recently Save()-ed position.
	Restore()
}

// shapedRunSlice is a [RunIterator] built from already-shaped text.
type shapedRunSlice struct {
	runs     []Output
	idx      int
	savedIdx int
}

var _ RunIterator = (*shapedRunSlice)(nil)

// NewSliceIterator returns a [RunIterator] backed by an already-shaped slice of [Output]s.
func NewSliceIterator(outs []Output) RunIterator {
	return &shapedRunSlice{
		runs: outs,
	}
}

// Reset configures the runSlice for reuse with the given shaped text.
func (r *shapedRunSlice) Reset(outs []Output) {
	r.runs = outs
	r.idx = 0
	r.savedIdx = 0
}

// Next implements [RunIterator.Next].
func (r *shapedRunSlice) Next() (int, Output, bool) {
	idx, run, ok := r.Peek()
	if ok {
		r.idx++
	}
	return idx, run, ok
}

// Peek implements [RunIterator.Peek].
func (r *shapedRunSlice) Peek() (int, Output, bool) {
	if r.idx >= len(r.runs) {
		return r.idx, Output{}, false
	}
	next := r.runs[r.idx]
	return r.idx, next, true
}

// Save implements [RunIterator.Save].
func (r *shapedRunSlice) Save() {
	r.savedIdx = r.idx
}

// Restore implements [RunIterator.Restore].
func (r *shapedRunSlice) Restore() {
	r.idx = r.savedIdx
}

// wrapBuffer provides reusable buffers for line wrapping. When using a
// wrapBuffer, returned line wrapping results will use memory stored within
// the buffer. This means that the same buffer cannot be reused for another
// wrapping operation while the wrapped lines are still in use (unless they
// are deeply copied). If necessary, using multiple wrapBuffers can work
// around this restriction.
type wrapBuffer struct {
	// paragraph is a buffer holding paragraph allocated (primarily) from subregions
	// of the line field.
	paragraph []Line
	// line is a large buffer of Outputs that is used to build lines.
	line []Output
	// lineUsed is the index of the first unused element in line.
	lineUsed int
	// lineExhausted indicates whether the previous shaping used all of line.
	lineExhausted bool
	// alt is a smaller temporary buffer that holds candidate lines while
	// they are being built.
	alt []Output
	// altAdvance is the sum of the advances of each run in alt.
	altAdvance fixed.Int26_6
	// altSave is a slice into alt used to save and restore the state of the alt buffer.
	altSave []Output
	// altAdvanceSave is the altAdvance of the altSave field.
	altAdvanceSave fixed.Int26_6
	// best is a slice holding the best known line. When possible, it
	// is a subslice of line, but if line runs out of capacity it will
	// be heap allocated.
	best []Output
	// bestInLine tracks whether the current best is allocated from within line.
	bestInLine bool
}

func (w *wrapBuffer) reset() {
	if cap(w.paragraph) < 10 {
		w.paragraph = make([]Line, 0, 10)
	}
	w.paragraph = w.paragraph[:0]
	if cap(w.alt) < 10 {
		w.alt = make([]Output, 0, 10)
	}
	w.alt = w.alt[:0]
	w.altAdvance = 0
	w.altSave = w.alt[:0]
	w.altAdvanceSave = 0
	if cap(w.line) < 100 {
		w.line = make([]Output, 0, 100)
	}
	w.line = w.line[:0]
	w.lineUsed = 0
	w.best = nil
	w.bestInLine = false
	if w.lineExhausted {
		w.lineExhausted = false
		// Trigger the go slice growth heuristic by appending an element to
		// the capacity.
		w.line = append(w.line[:cap(w.line)], Output{})[:0]
	}
}

// singleRunParagraph is an optimized helper for quickly constructing
// a []Line containing only a single run.
func (w *wrapBuffer) singleRunParagraph(run Output) []Line {
	w.paragraph = w.paragraph[:0]
	s := w.line[w.lineUsed : w.lineUsed+1]
	s[0] = run
	w.paragraphAppend(s)
	return w.finalParagraph()
}

func (w *wrapBuffer) paragraphAppend(line []Output) {
	w.paragraph = append(w.paragraph, line)
}

func (w *wrapBuffer) finalParagraph() []Line {
	return w.paragraph
}

func (w *wrapBuffer) startLine() {
	w.alt = w.alt[:0]
	w.altAdvance = 0
	w.altSave = w.alt[:0]
	w.altAdvanceSave = 0
	w.best = nil
	w.bestInLine = false
}

// candidateLen returns the number of [Output]s in the current line wrapping candidate.
func (w *wrapBuffer) candidateLen() int { return len(w.alt) }

// candidateAppend adds the given run to the current line wrapping candidate.
func (w *wrapBuffer) candidateAppend(run Output) {
	w.alt = append(w.alt, run)
	w.altAdvance = w.altAdvance + run.Advance
}

// candidateSave captures the current state of the line candidate, enabling it to
// be restored by a call to candidateRestore(). Only one state can be saved at a
// time, and a subsequent call to candidateSave will override any current saved
// state.
func (w *wrapBuffer) candidateSave() {
	w.altSave = w.alt
	w.altAdvanceSave = w.altAdvance
}

// candidateRestore resets the state of the line candidate to the state at the time
// of the most recent call to candidateSave().
func (w *wrapBuffer) candidateRestore() {
	w.alt = w.altSave
	w.altAdvance = w.altAdvanceSave
}

func (w *wrapBuffer) candidateAdvance() fixed.Int26_6 {
	return w.altAdvance
}

// markCandidateBest marks that the current line wrapping candidate is the best
// known line wrapping candidate with the given suffixes. Providing suffixes does
// not modify the current candidate, but does ensure that the "best" candidate ends
// with them.
func (w *wrapBuffer) markCandidateBest(suffixes ...Output) {
	neededLen := len(w.alt) + len(suffixes)
	if len(w.line[w.lineUsed:cap(w.line)]) < neededLen {
		w.lineExhausted = true
		w.best = make([]Output, neededLen)
		w.bestInLine = false
	} else {
		w.best = w.line[w.lineUsed : w.lineUsed+neededLen]
		w.bestInLine = true
	}
	n := copy(w.best, w.alt)
	copy(w.best[n:], suffixes)
}

// hasBest returns whether there is currently a known valid line wrapping candidate
// for the line.
func (w *wrapBuffer) hasBest() bool {
	return len(w.best) > 0
}

// finalizeBest commits the storage for the current best line and returns it.
func (w *wrapBuffer) finalizeBest() []Output {
	if w.bestInLine {
		w.lineUsed += len(w.best)
	}
	return w.best
}

// LineWrapper holds reusable state for a line wrapping operation. Reusing
// LineWrappers for multiple paragraphs should improve performance.
type LineWrapper struct {
	// config holds the current line wrapping settings.
	config WrapConfig
	// truncating tracks whether the wrapper should be performing truncation.
	truncating bool
	// seg is an internal storage used to initiate the breaker iterator.
	seg segmenter.Segmenter

	// breaker provides line-breaking candidates.
	breaker *breaker

	// scratch holds wrapping algorithm storage buffers for reuse.
	scratch wrapBuffer

	// mapper tracks rune->glyphCluster mappings.
	mapper runMapper
	// glyphRuns holds the runs of shaped text being wrapped.
	glyphRuns RunIterator
	// lineStartRune is the rune index of the first rune on the next line to
	// be shaped.
	lineStartRune int
	// more indicates that the iteration API has more data to return.
	more bool
}

// Prepare initializes the LineWrapper for the given paragraph and shaped text.
// It must be called prior to invoking WrapNextLine. Prepare invalidates any
// lines previously returned by this wrapper.
func (l *LineWrapper) Prepare(config WrapConfig, paragraph []rune, runs RunIterator) {
	l.config = config
	l.truncating = l.config.TruncateAfterLines > 0
	l.breaker = newBreaker(&l.seg, paragraph)
	l.glyphRuns = runs
	l.lineStartRune = 0
	l.more = true
	l.mapper.valid = false
	l.scratch.reset()
}

// WrapParagraph wraps the paragraph's shaped glyphs to a constant maxWidth.
// It is equivalent to iteratively invoking WrapLine with a constant maxWidth.
// If the config has a non-zero TruncateAfterLines, WrapParagraph will return at most
// that many lines. The truncated return value is the count of runes truncated from
// the end of the text. The returned lines are only valid until the next call to
// [*LineWrapper.WrapParagraph] or [*LineWrapper.Prepare].
//
// See [(*LineWrapper).WrapNextLine] for a description of how [WrapConfig]'s truncation
// features impact the wrapped text output. This method returns the quantity of runes
// truncated by line wrapping in the [truncated] return value.
//
// See also [WrapParagraphF] which supports a decimal [maxWidth].
func (l *LineWrapper) WrapParagraph(config WrapConfig, maxWidth int, paragraph []rune, runs RunIterator) (_ []Line, truncated int) {
	maxFixed := math.MaxInt32 >> 6
	if maxWidth > maxFixed {
		maxWidth = maxFixed
	}
	return l.WrapParagraphF(config, fixed.I(maxWidth), paragraph, runs)
}

// WrapParagraphF is the same as [WrapParagraph], but accepts a non integer [maxWidth].
func (l *LineWrapper) WrapParagraphF(config WrapConfig, maxWidth fixed.Int26_6, paragraph []rune, runs RunIterator) (_ []Line, truncated int) {
	l.scratch.reset()
	// Check whether we can skip line wrapping altogether for the simple single-run-that-fits case.
	if !(config.TextContinues && config.TruncateAfterLines == 1) {
		runs.Save()
		// We can only skip wrapping if the text doesn't contain any forced line
		// breaks that need to be evaluated by the real algorithm, so we need to
		// quickly scan it for that.
		l.breaker = newBreaker(&l.seg, paragraph)
		hasMandatoryBreak := false
		for {
			option, ok := l.breaker.nextWordBreak()
			if !ok {
				break
			}
			if option.required {
				hasMandatoryBreak = true
				break
			}
		}
		if !hasMandatoryBreak {
			_, firstRun, hasFirst := runs.Next()
			_, _, hasSecond := runs.Peek()
			if hasFirst && !hasSecond {
				if firstRun.Advance <= maxWidth {
					return l.scratch.singleRunParagraph(firstRun), 0
				}
			}
		}
		runs.Restore()
	}

	l.Prepare(config, paragraph, runs)
	var (
		line WrappedLine
		done bool
	)
	for !done {
		line, done = l.WrapNextLineF(maxWidth)
		if line.Line != nil {
			l.scratch.paragraphAppend(line.Line)
		}
	}
	return l.scratch.finalParagraph(), line.Truncated
}

// fillUntil tries to fill the line candidate slice with runs until it reaches a run containing the
// provided break option.
func (l *LineWrapper) fillUntil(runs RunIterator, option breakOption) {
	currRunIndex, run, more := runs.Peek()
	for more && option.breakAtRune >= run.Runes.Count+run.Runes.Offset {
		if l.lineStartRune >= run.Runes.Offset+run.Runes.Count {
			// Consume the run we peeked (which we know is valid)
			_, _, _ = runs.Next()

			currRunIndex, run, more = runs.Peek()
			continue
		} else if l.lineStartRune > run.Runes.Offset {
			// If part of this run has already been used on a previous line, trim
			// the runes corresponding to those glyphs off.
			l.mapper.mapRun(currRunIndex, run)
			isFirstInLine := l.scratch.candidateLen() == 0
			run = cutRun(run, l.mapper.mapping, l.lineStartRune, run.Runes.Count+run.Runes.Offset, isFirstInLine)
		}
		// While the run being processed doesn't contain the current line breaking
		// candidate, just append it to the candidate line.
		l.scratch.candidateAppend(run)
		// Consume the run we peeked (which we know is valid)
		_, _, _ = runs.Next()

		currRunIndex, run, more = runs.Peek()
	}
}

// lineConfig tracks settings for line wrapping a single line of text.
type lineConfig struct {
	// truncating indicates whether this line is being truncated (if sufficiently long).
	truncating bool
	// maxWidth is the maximum space a line can occupy.
	maxWidth fixed.Int26_6
	// truncatedMaxWidth holds the maximum width of the line available for text if the truncator
	// is occupying part of the line.
	truncatedMaxWidth fixed.Int26_6
}

// WrappedLine is the result of wrapping one line of text.
type WrappedLine struct {
	// Line is the content of the line, as a slice of shaped runs
	Line Line
	// Truncated is the count of runes truncated from the end of the line,
	// if this line was truncated.
	Truncated int
	// NextLine is the indice (in the input text slice) of the begining
	// of the next line. It will equal len(text) if all the text
	// fit in one line.
	NextLine int

	// TrimmedTrailingWhitespace is the space taken by trailing whitespace
	// before if was trimmed (usually positive).
	// It is zero if [DisableTrailingWhitespaceTrim] is set to true,
	// or if there is no whitespace at the end of the line.
	TrimmedTrailingWhitespace fixed.Int26_6
}

// swapVisualOrder inverts the visual index of runs in [subline], by swapping pairs of visual indices across the midpoint
// of the slice.
func swapVisualOrder(subline Line) {
	L := len(subline)
	for i := range subline[0 : L/2] {
		j := (L - i) - 1
		subline[i].VisualIndex, subline[j].VisualIndex = subline[j].VisualIndex, subline[i].VisualIndex
	}
}

// computeBidiOrdering resolves the [VisualIndex] of each run.
func computeBidiOrdering(dir di.Direction, finalLine Line) {
	bidiStart := -1
	for idx, run := range finalLine {
		basePosition := idx
		if dir.Progression() == di.TowardTopLeft {
			basePosition = len(finalLine) - 1 - idx
		}
		finalLine[idx].VisualIndex = int32(basePosition)
		if run.Direction == dir {
			if bidiStart != -1 {
				swapVisualOrder(finalLine[bidiStart:idx])
				bidiStart = -1
			}
		} else if bidiStart == -1 {
			bidiStart = idx
		}
	}
	if bidiStart != -1 {
		swapVisualOrder(finalLine[bidiStart:])
	}
}
