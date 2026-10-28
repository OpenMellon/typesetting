package harfbuzz

func (b *Buffer) infosSetGlyphFlags(infos []GlyphInfo, start, end, cluster int, mask GlyphMask) {
	if start == end {
		return
	}

	b.maxOps -= end - start

	clusterFirst := infos[start].Cluster
	clusterLast := infos[end-1].Cluster

	if b.ClusterLevel == Characters || (cluster != clusterFirst && cluster != clusterLast) {
		for i := start; i < end; i++ {
			if cluster != infos[i].Cluster {
				infos[i].Mask |= mask
			}
		}
		return
	}

	/* Monotone clusters */

	if cluster == clusterFirst {
		for i := end; start < i && infos[i-1].Cluster != clusterFirst; i-- {
			infos[i-1].Mask |= mask
		}
	} else /* cluster == clusterLast */ {
		for i := start; i < end && infos[i].Cluster != clusterLast; i++ {
			infos[i].Mask |= mask
		}
	}
}

// reset `b.outInfo`, and adjust `pos` to have
// same length as `Info` (without zeroing its values)
func (b *Buffer) clearPositions() {
	b.haveOutput = false
	// b.have_positions = true

	b.outInfo = b.outInfo[:0]

	L := len(b.Info)
	if cap(b.Pos) >= L {
		b.Pos = b.Pos[:L]
	} else {
		b.Pos = make([]GlyphPosition, L)
	}
}

// truncate `outInfo` and set `haveOutput`
func (b *Buffer) removeOutput(setOutput bool) {
	b.haveOutput = setOutput
	// b.have_positions = false

	b.outInfo = b.outInfo[:0]
}

// truncate `outInfo` and set `haveOutput` to true
func (b *Buffer) clearOutput() {
	b.removeOutput(true)
	b.idx = 0
}

func (b *Buffer) clearContext(side uint) { b.context[side] = b.context[side][:0] }

// reverses the subslice [start:end] of the buffer contents
func (b *Buffer) reverseRange(start, end int) {
	if end-start < 2 {
		return
	}
	info := b.Info[start:end]
	pos := b.Pos[start:end]
	L := len(info)
	_ = pos[L-1] // BCE
	for i := L/2 - 1; i >= 0; i-- {
		opp := L - 1 - i
		info[i], info[opp] = info[opp], info[i]
		pos[i], pos[opp] = pos[opp], pos[i] // same length
	}
}

// Reverse reverses buffer contents, that is the order of the `Info` and `Pos` slices.
func (b *Buffer) Reverse() { b.reverseRange(0, len(b.Info)) }

func (b *Buffer) reverseClusters() {
	b.reverseGroups(func(gi1, gi2 *GlyphInfo) bool {
		return gi1.Cluster == gi2.Cluster
	}, false)
}

// mergeClusters = false
func (b *Buffer) reverseGroups(groupFunc func(*GlyphInfo, *GlyphInfo) bool, mergeClusters bool) {
	if len(b.Info) == 0 {
		return
	}

	count := len(b.Info)
	start := 0
	var i int
	for i = 1; i < count; i++ {
		if !groupFunc(&b.Info[i-1], &b.Info[i]) {
			if mergeClusters {
				b.mergeClusters(start, i)
			}
			b.reverseRange(start, i)
			start = i
		}
	}

	if mergeClusters {
		b.mergeClusters(start, i)
	}
	b.reverseRange(start, i)

	b.Reverse()
}

// swap back the temporary outInfo buffer to `Info`
// and resets the cursor `idx`.
// Assume that haveOutput is true, and toogle it.
func (b *Buffer) swapBuffers() {
	b.nextGlyphs(len(b.Info) - b.idx)
	b.haveOutput = false
	b.Info, b.outInfo = b.outInfo, b.Info
	b.idx = 0
}

// returns an unique id
func (b *Buffer) allocateLigID() uint8 {
	ligID := uint8(b.serial & 0x07)
	b.serial++
	if ligID == 0 { // in case of overflow
		ligID = b.allocateLigID()
	}
	return ligID
}

// return false if maximum ops limit is reached
func (b *Buffer) shiftForward(count int) bool {
	//   assert (have_output);
	L := len(b.Info)
	b.maxOps -= L - b.idx
	if b.maxOps < 0 {
		return false
	}
	b.Info = append(b.Info, make([]GlyphInfo, count)...)
	copy(b.Info[b.idx+count:], b.Info[b.idx:L])
	b.idx += count
	return true
}

func (b *Buffer) moveTo(i int) {
	if !b.haveOutput {
		// assert(i <= len)
		b.idx = i
		return
	}

	// assert(i <= out_len+(len-idx))
	outL := len(b.outInfo)
	if outL < i {
		count := i - outL
		b.outInfo = append(b.outInfo, b.Info[b.idx:count+b.idx]...)
		b.idx += count
	} else if outL > i {
		/* Tricky part: rewinding... */
		count := outL - i

		if b.idx < count {
			ok := b.shiftForward(count - b.idx)
			if !ok {
				return
			}
		}

		// assert(idx >= count)

		b.idx -= count
		copy(b.Info[b.idx:], b.outInfo[outL-count:outL])
		b.outInfo = b.outInfo[:outL-count]
	}
}

func (b *Buffer) collectGlyphs(out *intSet) {
	out.Clear()
	for _, info := range b.Info {
		out.addGlyph(info.Glyph)
	}
}

// iterator over the grapheme of a buffer
type graphemesIterator struct {
	buffer *Buffer
	start  int
}

// at the end of the buffer, start >= len(info)
func (g *graphemesIterator) next() (start, end int) {
	info := g.buffer.Info
	count := len(info)
	start = g.start
	if start >= count {
		return
	}
	for end = g.start + 1; end < count && info[end].isContinuation(); end++ {
	}
	g.start = end
	return start, end
}

func (b *Buffer) graphemesIterator() (*graphemesIterator, int) {
	return &graphemesIterator{buffer: b}, len(b.Info)
}

// iterator over clusters of a buffer with the loop
// for start, end := iter.Next(); start < count; start, end = iter.Next() {}
type clusterIterator struct {
	buffer *Buffer
	start  int
}

// returns the next cluster delimited by [start, end[
func (c *clusterIterator) next() (start, end int) {
	info := c.buffer.Info
	count := len(info)
	start = c.start
	if start >= count {
		return
	}
	cluster := info[start].Cluster
	for end = start + 1; end < count && cluster == info[end].Cluster; end++ {
	}
	c.start = end

	return start, end
}

func (b *Buffer) clusterIterator() (*clusterIterator, int) {
	return &clusterIterator{buffer: b}, len(b.Info)
}

// iterator over syllables of a buffer
type syllableIterator struct {
	buffer *Buffer
	start  int
}

func (c *syllableIterator) next() (start, end int) {
	info := c.buffer.Info
	count := len(c.buffer.Info)
	start = c.start
	if start >= count {
		return
	}
	syllable := info[start].syllable
	for end = start + 1; end < count && syllable == info[end].syllable; end++ {
	}
	c.start = end
	return start, end
}

func (b *Buffer) syllableIterator() (*syllableIterator, int) {
	return &syllableIterator{buffer: b}, len(b.Info)
}

// only modifies Info, thus assume Pos is not used yet
func (b *Buffer) sort(start, end int, compar func(a, b *GlyphInfo) int) {
	for i := start + 1; i < end; i++ {
		j := i
		for j > start && compar(&b.Info[j-1], &b.Info[i]) > 0 {
			j--
		}
		if i == j {
			continue
		}
		// move item i to occupy place for item j, shift what's in between.
		b.mergeClusters(j, i+1)

		t := b.Info[i]
		copy(b.Info[j+1:], b.Info[j:i])
		b.Info[j] = t
	}
}

func (b *Buffer) mergeOutClusters(start, end int) {
	if b.ClusterLevel == Characters {
		return
	}

	if end-start < 2 {
		return
	}

	cluster := b.outInfo[start].Cluster

	for i := start + 1; i < end; i++ {
		cluster = min(cluster, b.outInfo[i].Cluster)
	}

	/* Extend start */
	for start != 0 && b.outInfo[start-1].Cluster == b.outInfo[start].Cluster {
		start--
	}

	/* Extend end */
	for end < len(b.outInfo) && b.outInfo[end-1].Cluster == b.outInfo[end].Cluster {
		end++
	}

	/* If we hit the end of out-buffer, continue in buffer. */
	if end == len(b.outInfo) {
		endC := b.outInfo[end-1].Cluster
		for i := b.idx; i < len(b.Info) && b.Info[i].Cluster == endC; i++ {
			b.Info[i].setCluster(cluster, 0)
		}
	}

	for i := start; i < end; i++ {
		b.outInfo[i].setCluster(cluster, 0)
	}
}
