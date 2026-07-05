package infrastructure

type Config struct {
	Addr      string
	Customers int
	Accounts  int
	Shards    int
	Seed      uint64
	RHLoad    float64
}
