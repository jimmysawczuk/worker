package worker

import (
	"sync"
)

type Event int

const (
	jobAdded Event = 1 << iota
	jobStarted
	jobFinished

	JobAdded
	JobStarted
	JobFinished
)

type JobStatus int

const (
	Queued JobStatus = 1 << iota
	Running
	Finished
	Errored
)

type Counter struct {
	val  int
	lock sync.RWMutex
}

type Queue struct {
	jobs []*Package
	lock sync.RWMutex
}

type Switch struct {
	val  bool
	lock sync.RWMutex
}

type LockableChanInt struct {
	ch   chan int
	lock sync.RWMutex
}

type Map struct {
	jobs map[int64]*Package
	lock sync.RWMutex
}

type Register []LockableChanInt

func (c *Counter) Add(i int) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.val = c.val + i
}

func (c *Counter) Sub(i int) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.val = c.val - i
}

func (c *Counter) AddOne() {
	c.Add(1)
}

func (c *Counter) SubOne() {
	c.Sub(1)
}

func (c *Counter) Val() int {
	c.lock.RLock()
	defer c.lock.RUnlock()

	v := c.val

	return v
}

func NewQueue() Queue {
	q := Queue{}

	return q
}

func (q *Queue) Top() *Package {
	q.lock.Lock()
	defer q.lock.Unlock()

	if len(q.jobs) > 0 {
		j := q.jobs[0]
		q.jobs = q.jobs[1:]
		return j
	} else {
		return nil
	}

}

func (q *Queue) Add(j Package) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.jobs = append(q.jobs, &j)
}

func (q *Queue) Len() int {
	q.lock.RLock()
	defer q.lock.RUnlock()

	l := len(q.jobs)

	return l
}

func (s *Switch) On() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	r := s.val

	return r
}

func (s *Switch) Toggle() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.val = !s.val
}

func (s *Switch) Set(v bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.val = v
}

func (lc *LockableChanInt) Ch() chan int {
	lc.lock.RLock()
	defer lc.lock.RUnlock()

	v := lc.ch

	return v
}

func (lc *LockableChanInt) SetCh(ch chan int) {
	lc.lock.Lock()
	defer lc.lock.Unlock()

	lc.ch = ch
}

func (r *Register) Empty() bool {
	for i := 0; i < len(*r); i++ {
		if (*r)[i].Ch() != nil {
			return false
		}
	}

	return true
}

func NewMap() Map {
	m := Map{
		jobs: make(map[int64]*Package),
	}

	return m
}

func (m *Map) Set(val Package) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.jobs[val.ID] = &val
}

func (m *Map) Get(id int64) *Package {
	m.lock.Lock()
	defer m.lock.Unlock()

	val := m.jobs[id]

	return val
}

type WorkerStats struct {
	Total    int64
	Running  int64
	Finished int64
	Queued   int64
	Errored  int64
}
