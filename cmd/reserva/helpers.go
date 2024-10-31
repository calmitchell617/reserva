package main

import (
	"sync"
)

type SafeInt64Map struct {
	mu     sync.Mutex
	valMap map[int64]bool
}

func (s *SafeInt64Map) Add(element int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.valMap[element] = true
}

func (s *SafeInt64Map) GetRandom() (element int64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key := range s.valMap {
		element = key
		break
	}

	return element, nil
}

func (s *SafeInt64Map) Remove(element int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.valMap, element)
}
