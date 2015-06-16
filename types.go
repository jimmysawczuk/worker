package worker

import (
	"sync"
)

type ExitCode int

const (
	ExitNormally ExitCode = 0
	ExitWhenDone          = 4
)

type Event int

const (
	jobAdded Event = 1 << iota
	jobStarted
	jobFinished

	JobAdded Event = 1 << iota
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

type registerEntry struct {
	ch   chan bool
	lock *sync.RWMutex
}

func (lc *registerEntry) init() {
	if lc.lock == nil {
		lc.lock = new(sync.RWMutex)
	}
}

func (lc *registerEntry) Ch() chan bool {
	lc.init()

	lc.lock.RLock()
	defer lc.lock.RUnlock()

	v := lc.ch

	return v
}

func (lc *registerEntry) SetCh(ch chan bool) {
	lc.init()

	lc.lock.Lock()
	defer lc.lock.Unlock()

	lc.ch = ch
}

type register []registerEntry

func (r *register) Empty() bool {
	for i := 0; i < len(*r); i++ {
		if (*r)[i].Ch() != nil {
			return false
		}
	}

	return true
}

type WorkerStats struct {
	Total    int64
	Running  int64
	Finished int64
	Queued   int64
	Errored  int64
}
