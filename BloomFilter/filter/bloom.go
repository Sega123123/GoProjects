package filter

import (
	"hash/fnv"
	"runtime"
	"sync"
)

type writeRequest struct {
	index int
	done  chan struct{}
}

type BloomFilter struct {
	size        int
	hashesCount int
	bits        map[int]bool
	channel     chan writeRequest
	mutex       sync.Mutex
}

func NewBloomFilter(size int, hashesCount int) *BloomFilter {
	bf := &BloomFilter{
		size:        size,
		hashesCount: hashesCount,
		bits:        make(map[int]bool),
		channel:     make(chan writeRequest),
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
	for _, i := range bf.hashesAsync(item, bf.hashesCount) {
		done := make(chan struct{})
		bf.channel <- writeRequest{
			index: i,
			done:  done,
		}
		<-done
	}
}

func (bf *BloomFilter) Exists(item string) bool {
	for _, i := range bf.hashesAsync(item, bf.hashesCount) {
		if !bf.bits[i] {
			return false
		}
	}
	return true
}

func (bf *BloomFilter) hashesAsync(s string, n int) []int {
	type hashJob struct {
		index int
		salt  string
	}

	type hashResult struct {
		index int
		value int
	}

	numWorkers := runtime.NumCPU()
	jobs := make(chan hashJob, n)
	results := make(chan hashResult, n)
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				h := fnv.New32a()
				h.Write([]byte(s + job.salt))
				hashed := int(h.Sum32()) % bf.size
				results <- hashResult{index: job.index, value: hashed}
			}
		}()
	}

	go func() {
		for i := 0; i < n; i++ {
			salt := string(rune('a' + i))
			jobs <- hashJob{index: i, salt: salt}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	hashes := make([]int, n)
	for res := range results {
		hashes[res.index] = res.value
	}
	return hashes
}
