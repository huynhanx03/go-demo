package lookup

const packed52Bits = 52

type packed52 struct {
	data []uint64
}

func newPacked52(n int) packed52 {
	totalBits := uint64(n * packed52Bits)
	words := int((totalBits + 63) / 64)
	return packed52{data: make([]uint64, words)}
}

func (p packed52) set(i int, v uint64) {
	v &= maxEncoded52
	bitPos := uint64(i * packed52Bits)
	word := int(bitPos / 64)
	off := bitPos % 64

	p.data[word] &= ^(maxEncoded52 << off)
	p.data[word] |= v << off

	if off > 12 {
		highBits := off - 12
		p.data[word+1] &= ^((uint64(1) << highBits) - 1)
		p.data[word+1] |= v >> (64 - off)
	}
}

func (p packed52) get(i int) uint64 {
	bitPos := uint64(i * packed52Bits)
	word := int(bitPos / 64)
	off := bitPos % 64

	v := p.data[word] >> off
	if off > 12 {
		highBits := off - 12
		v |= p.data[word+1] << (64 - off)
		v &= (uint64(1) << (52 + highBits)) - 1
	}
	return v & maxEncoded52
}

func (p packed52) byteSize() uint64 {
	return uint64(len(p.data) * 8)
}
