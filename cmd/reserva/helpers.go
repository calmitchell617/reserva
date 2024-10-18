package main

import (
	"errors"
	"math/rand"
	"sync"
)

type SafeInt64Slice struct {
	mu    sync.Mutex
	slice []int64
}

func (s *SafeInt64Slice) Add(element int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.slice = append(s.slice, element)
}

func (s *SafeInt64Slice) GetRandom() (index int64, element int64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.slice) == 0 {
		return 0, 0, errors.New("empty slice")
	}
	index = rand.Int63n(int64(len(s.slice)))
	element = s.slice[index]
	return index, element, nil
}

func (s *SafeInt64Slice) Remove(index int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if index < 0 || index >= int64(len(s.slice)) {
		return
	}

	if index == int64(len(s.slice))-1 {
		s.slice = s.slice[:index]
	} else {
		s.slice = append(s.slice[:index], s.slice[index+1:]...)
	}
}
