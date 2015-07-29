package worker

import (
	"sync"
)

type Package struct {
	ID     int64
	status JobStatus
	job    Job

	statusLock *sync.RWMutex
}

func NewPackage(id int64, j Job) *Package {
	p := &Package{
		ID:         id,
		status:     Queued,
		job:        j,
		statusLock: new(sync.RWMutex),
	}

	return p
}

func (p *Package) Job() Job {
	return p.job
}

func (p *Package) SetStatus(inc JobStatus) {
	p.statusLock.Lock()
	defer p.statusLock.Unlock()

	p.status = inc
}

func (p *Package) Status() JobStatus {
	p.statusLock.RLock()
	defer p.statusLock.RUnlock()

	return p.status
}
