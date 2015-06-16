package worker

import (
	"sync"
)

type Queue struct {
	jobs []*Package
	lock *sync.RWMutex
}

func NewQueue() Queue {
	q := Queue{
		jobs: make([]*Package, 0),
		lock: new(sync.RWMutex),
	}

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

func (q *Queue) Add(j *Package) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.jobs = append(q.jobs, j)
}

func (q *Queue) Len() int {
	q.lock.RLock()
	defer q.lock.RUnlock()

	l := len(q.jobs)

	return l
}
