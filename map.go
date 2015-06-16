package worker

import (
	"sync"
)

type Map struct {
	jobs map[int64]*Package
	lock *sync.RWMutex
}

func NewMap() Map {
	m := Map{
		jobs: make(map[int64]*Package),
		lock: new(sync.RWMutex),
	}

	return m
}

func (m *Map) Set(val *Package) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.jobs[val.ID] = val
}

func (m *Map) Get(id int64) *Package {
	m.lock.Lock()
	defer m.lock.Unlock()

	val := m.jobs[id]

	return val
}
