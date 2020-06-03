package stats

import "sync"

type Stats struct {
	counter int64
	mux     sync.Mutex
}

func (s *Stats) Add() {
	s.mux.Lock()
	s.counter++
	s.mux.Unlock()
}

func (s *Stats) Count() int64 {
	return s.counter
}
