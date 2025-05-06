package main

import (
	"BloomFilter/filter"
	"errors"
	"fmt"
	"sync"
	"time"
)

type LeakyBucket struct {
	capacity int
	rate     int
	interval time.Duration
	bf       *filter.BloomFilter
	mu       sync.Mutex
	tokens   int

	stopChan chan struct{}
}

func NewLeakyBucket(capacity, rate int, interval time.Duration) *LeakyBucket {
	lb := &LeakyBucket{
		capacity: capacity,
		rate:     rate,
		interval: interval,
		bf:       filter.NewBloomFilter(1000, 10),
		tokens:   0,
		stopChan: make(chan struct{}),
	}
	go lb.startLeaking()
	return lb
}

func (lb *LeakyBucket) startLeaking() {
	ticker := time.NewTicker(lb.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			lb.mu.Lock()
			if lb.tokens > 0 {
				lb.tokens -= lb.rate
				if lb.tokens < 0 {
					lb.tokens = 0
				}
			}
			lb.mu.Unlock()

		case <-lb.stopChan:
			fmt.Println("Leaky Bucket stopped.")
			return
		}
	}
}

func (lb *LeakyBucket) AllowRequest(id string) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if !lb.bf.Exists(id) {
		lb.bf.Add(id)
	} else {
		fmt.Println("Request", id, "is already allowed")
		return nil
	}

	if lb.tokens < lb.capacity {
		lb.tokens++
		return nil
	}
	return errors.New("Bucket Overflow")
}

func (lb *LeakyBucket) Stop() {
	close(lb.stopChan)
}

func main() {
	lb := NewLeakyBucket(3, 1, time.Second)

	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("req-%d", i)
		err := lb.AllowRequest(id)

		lb.AllowRequest(id) // только для демонстрации проверки одинаковых запросов

		if err != nil {
			fmt.Println("Request", id, "blocked:", err)
		} else {
			fmt.Println("Request", id, "allowed")
		}
		time.Sleep(400 * time.Millisecond)
	}
	lb.Stop()
	time.Sleep(1 * time.Second)
}
