package cache

const sketchDepth = 4

// countMinSketch is an implementation of count-min sketch with 4-bit counters.
type countMinSketch struct {
	counters []uint64
	mask     uint32
}

// init initialize count-min sketch with the given width.
func newCountMinSketch(width int) *countMinSketch {
	size := nextPowerOfTwo(uint32(width)) >> 2
	if size < 1 {
		size = 1
	}
	countMinSketch := countMinSketch{
		make([]uint64, size),
		size - 1,
	}
	return &countMinSketch
}

// add increase counters with given hash
func (c *countMinSketch) add(h uint64) {
	hash1, hash2 := uint32(h), uint32(h>>32)

	for i := uint32(0); i < sketchDepth; i++ {
		combinedHash := hash1 + (i * hash2)
		idx, off := c.position(combinedHash)
		c.inc(idx, (16*i)+off)
	}
}

// estimate returns minimum value of counters associated with the given hash.
func (c *countMinSketch) estimate(h uint64) uint8 {
	hash1, hash2 := uint32(h), uint32(h>>32)

	var min uint8 = 0xFF
	for i := uint32(0); i < sketchDepth; i++ {
		combinedHash := hash1 + (i * hash2)
		idx, off := c.position(combinedHash)
		count := c.val(idx, (16*i)+off)
		if count < min {
			min = count
		}
	}
	return min
}

func (c *countMinSketch) reset() {
	for i, v := range c.counters {
		if v != 0 {
			//divides all by two.
			c.counters[i] = (v >> 1) & 0x7777777777777777
		}
	}
}

func (c *countMinSketch) position(h uint32) (idx uint32, off uint32) {
	idx = (h >> 2) & c.mask
	off = (h & 3) << 2
	return
}

// inc increases value at index idx.
func (c *countMinSketch) inc(idx, off uint32) {
	v := c.counters[idx]
	count := uint8(v>>off) & 0x0F
	if count < 15 {
		c.counters[idx] = v + (1 << off)
	}
}

// val returns value at index idx.
func (c *countMinSketch) val(idx, off uint32) uint8 {
	v := c.counters[idx]
	return uint8(v>>off) & 0x0F
}

func (c *countMinSketch) clear() {
	for i := range c.counters {
		c.counters[i] = 0
	}
}
