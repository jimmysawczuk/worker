package worker

import (
	"sync"
)

type Switch struct {
	val  bool
	lock *sync.RWMutex
}

func (s *Switch) init() {
	if s.lock == nil {
		s.lock = new(sync.RWMutex)
	}
}

func (s *Switch) On() bool {
	s.init()

	s.lock.RLock()
	defer s.lock.RUnlock()

	r := s.val

	return r
}

func (s *Switch) Toggle() {
	s.init()

	s.lock.Lock()
	defer s.lock.Unlock()

	s.val = !s.val
}

func (s *Switch) Set(v bool) {
	s.init()

	s.lock.Lock()
	defer s.lock.Unlock()

	s.val = v
}
