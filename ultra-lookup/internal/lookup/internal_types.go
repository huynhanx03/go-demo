package lookup

type kindRecord struct {
	id    uint64
	shard uint8
}

func alreadySet(bits []uint64, slot uint64) bool {
	word := slot / 64
	bit := slot % 64
	return bits[word]&(uint64(1)<<bit) != 0
}

func markSet(bits []uint64, slot uint64) {
	word := slot / 64
	bit := slot % 64
	bits[word] |= uint64(1) << bit
}
