package lookup

type SyntheticConfig struct {
	Customers int
	Accounts  int
	Shards    int
	Seed      uint64
}

func GenerateSyntheticEncoded(cfg SyntheticConfig) []EncodedRecord {
	shards := cfg.Shards
	if shards <= 0 {
		shards = 1
	}
	if shards > 256 {
		shards = 256
	}

	records := make([]EncodedRecord, 0, cfg.Customers+cfg.Accounts)
	seed := cfg.Seed
	next := func() uint64 {
		seed = splitmix64(seed + 0x9e37_79b9_7f4a_7c15)
		return seed
	}

	for i := 0; i < cfg.Customers; i++ {
		records = append(records, EncodedRecord{
			Kind:  KindCustomer,
			ID:    uint64(i + 1),
			Shard: uint8(next() % uint64(shards)),
		})
	}

	for i := 0; i < cfg.Accounts; i++ {
		records = append(records, EncodedRecord{
			Kind:  KindAccount,
			ID:    uint64(i + 1),
			Shard: uint8(next() % uint64(shards)),
		})
	}

	return records
}
