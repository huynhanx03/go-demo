package driven

import (
	"log"

	"github.com/huynhanx03/go-demo/promotion-file/internal/ports"
	"github.com/huynhanx03/go-demo/promotion-file/pkg/datastructs/bloom"
)

type bloomRepository struct {
	bf *bloom.Bloom
}

func NewBloomRepository(bf *bloom.Bloom) ports.IPromotionRepository {
	if bf == nil {
		log.Panic("Bloom Filter cannot be nil")
	}
	return &bloomRepository{bf: bf}
}

func (r *bloomRepository) Contains(code string) bool {
	return r.bf.HasString(code)
}
