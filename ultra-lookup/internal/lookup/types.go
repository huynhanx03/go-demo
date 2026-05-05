package lookup

import "fmt"

// Kind identifies which ID-space the record belongs to.
type Kind uint8

const (
	KindCustomer Kind = iota
	KindAccount
)

func (k Kind) String() string {
	switch k {
	case KindCustomer:
		return "customer"
	case KindAccount:
		return "account"
	default:
		return fmt.Sprintf("kind(%d)", uint8(k))
	}
}

// Record is one input mapping from (kind, base36-10 id) -> shard.
type Record struct {
	Kind  Kind
	ID    string
	Shard uint8
}

type EncodedRecord struct {
	Kind  Kind
	ID    uint64
	Shard uint8
}

// Config controls lookup engine build behavior.
type Config struct {
	RobinHoodLoadF float64
}

func DefaultConfig() Config {
	return Config{
		RobinHoodLoadF: 0.95,
	}
}

func (c Config) normalized() Config {
	cfg := c
	if cfg.RobinHoodLoadF <= 0.5 || cfg.RobinHoodLoadF >= 0.99 {
		cfg.RobinHoodLoadF = DefaultConfig().RobinHoodLoadF
	}
	return cfg
}

type MemoryStats struct {
	Entries         int
	KeyBytes        uint64
	ValueBytes      uint64
	IndexBytes      uint64
	TotalEstimatedB uint64
}
