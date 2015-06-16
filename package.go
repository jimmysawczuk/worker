package worker

import (
// "sync"
)

type Package struct {
	ID     int64
	Status JobStatus

	job Job
}

func (p *Package) Job() Job {
	return p.job
}
