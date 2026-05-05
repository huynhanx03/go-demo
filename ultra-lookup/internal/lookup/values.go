package lookup

type shardValues struct {
	data []uint8
}

func newShardValues(n int) shardValues {
	return shardValues{data: make([]uint8, n)}
}

func (v shardValues) set(i uint64, shard uint8) {
	v.data[i] = shard
}

func (v shardValues) get(i uint64) uint8 {
	return v.data[i]
}

func (v shardValues) byteSize() uint64 {
	return uint64(len(v.data))
}
