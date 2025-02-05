package replication

import (
	"github.com/bits-and-blooms/bloom/v3"
)

type BloomFilter struct {
	filter *bloom.BloomFilter
}

func NewBloomFilter() *BloomFilter {
	// Parameters: number of elements, false positive rate
	return &BloomFilter{
		filter: bloom.NewWithEstimates(20000, 0.01),
	}
}

func (bf *BloomFilter) Add(dataID string) {
	bf.filter.Add([]byte(dataID))
}

func (bf *BloomFilter) Check(dataID string) bool {
	return bf.filter.Test([]byte(dataID))
}
