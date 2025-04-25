package pkg

import (
	"hash/fnv"
	"sync"
)

type writeRequest struct {
	index int
	done  chan struct{}
}

type BloomFilter struct {
	size    int
	bits    map[int]bool
	channel chan writeRequest
	mutex   sync.Mutex
}

func NewBloomFilter(size int) *BloomFilter {
	bf := &BloomFilter{
		size:    size,
		bits:    make(map[int]bool),
		channel: make(chan writeRequest),
	}

	go func() {
		for req := range bf.channel {
			bf.mutex.Lock()
			bf.bits[req.index] = true
			bf.mutex.Unlock()
			req.done <- struct{}{}
		}
	}()

	return bf
}

func (bf *BloomFilter) Add(item string) {
	for _, i := range bf.hashes(item) {
		done := make(chan struct{})
		bf.channel <- writeRequest{
			index: i,
			done:  done,
		}
		<-done
	}
}

func (bf *BloomFilter) Exists(item string) bool {
	for _, i := range bf.hashes(item) {
		bf.mutex.Lock()
		exists := bf.bits[i]
		bf.mutex.Unlock()
		if !exists {
			return false
		}
	}
	return true
}

func (bf *BloomFilter) hashes(s string) []int {
	h1 := fnv.New32a()
	h1.Write([]byte(s))
	h2 := fnv.New32()
	h2.Write([]byte(s + "salt"))

	return []int{int(h1.Sum32()) % bf.size, int(h2.Sum32()) % bf.size}
}
