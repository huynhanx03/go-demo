package lookup

func indexKey(id uint64) uint64 {
	return id
}

func splitmix64(x uint64) uint64 {
	z := x + 0x9e37_79b9_7f4a_7c15
	z = (z ^ (z >> 30)) * 0xbf58_476d_1ce4_e5b9
	z = (z ^ (z >> 27)) * 0x94d0_49bb_1331_11eb
	return z ^ (z >> 31)
}
