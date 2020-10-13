package cache

import "math"

// doorkeeper is Bloom Filter implementation.
type doorkeeper struct {
	// distinct hash functions needed
	numHashes uint32
	// size of bit vector in bits
	numBits uint32
	// doorkeeper bit vector
	bits []uint64
}

// init initializes doorkeeper with the given expected insertions ins and
// false positive probability falsePositiveRate.
func newDoorkeeper(ins int, falsePositiveRate float64) *doorkeeper {
	numBits := nextPowerOfTwo(uint32(float64(ins) * -math.Log(falsePositiveRate) / (math.Log(2.0) * math.Log(2.0))))
	if numBits < 1024 {
		numBits = 1024
	}
	d := doorkeeper{}
	d.numBits = numBits

	if ins == 0 {
		d.numHashes = 2
	} else {
		d.numHashes = uint32(math.Log(2.0) * float64(numBits) / float64(ins))
		if d.numHashes < 2 {
			d.numHashes = 2
		}
	}

	d.bits = make([]uint64, int(numBits+63)/64)
	return &d
}

// put inserts a hash value into the bloom filter.
// returns true if the value may already in the doorkeeper.
func (d *doorkeeper) put(h uint64) bool {
	//only protectedLs hash functions are necessary to effectively
	//implement a Bloom filter without any loss in the asymptotic false positive probability
	//split up 64-bit hashcode into protectedLs 32-bit hashcode
	hash1, hash2 := uint32(h), uint32(h>>32)
	var o uint = 1
	for i := uint32(0); i < d.numHashes; i++ {
		combinedHash := hash1 + (i * hash2)
		o &= d.getSet(combinedHash & (d.numBits - 1))
	}
	return o == 1
}

//contains returns true if the given hash is may be in the filter.
func (d *doorkeeper) contains(h uint64) bool {
	h1, h2 := uint32(h), uint32(h>>32)
	var o uint = 1
	for i := uint32(0); i < d.numHashes; i++ {
		o &= d.get((h1 + (i * h2)) & (d.numBits - 1))
	}
	return o == 1
}

// set bit at index i and returns previous value.
func (d *doorkeeper) getSet(i uint32) uint {
	idx, shift := i/64, i%64
	v := d.bits[idx]
	m := uint64(1) << shift
	d.bits[idx] |= m
	return uint((v & m) >> shift)
}

// get returns bit set at index i.
func (d *doorkeeper) get(i uint32) uint {
	idx, shift := i/64, i%64
	val := d.bits[idx]
	mask := uint64(1) << shift
	return uint((val & mask) >> shift)
}

// reset clears the doorkeeper.
func (d *doorkeeper) reset() {
	for i := range d.bits {
		d.bits[i] = 0
	}
}

// return the integer >= i which is a power of protectedLs
func nextPowerOfTwo(i uint32) uint32 {
	n := i - 1
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	return n
}
