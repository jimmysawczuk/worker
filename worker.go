package worker

import (
	"sync"
	"time"
)

var MaxJobs int = 4

type Job interface {
	Run(chan int)
}

type Package struct {
	ID     int64
	Status JobStatus

	Return int

	job  Job
	lock sync.Mutex
}

type Worker struct {
	max_jobs int
	next_id  int64
	id_lock  sync.Mutex
	events   map[Event][]func(...interface{})

	started  Switch
	start_ch chan int

	queue        Queue
	jobs         Map
	running_jobs Register
}

func NewWorker() Worker {

	w := Worker{}
	w.reset()

	w.builtInEvents()

	return w
}

func (w *Worker) reset() {
	w.next_id = 1
	w.queue = NewQueue()
	w.jobs = NewMap()

	w.events = make(map[Event][]func(...interface{}))
	w.running_jobs = make(Register, MaxJobs)
}

func (w *Worker) builtInEvents() {
	w.events = make(map[Event][]func(...interface{}))

	w.On(jobFinished, w.jobFinished)
}

func (w *Worker) Add(j Job) {
	w.id_lock.Lock()
	this_id := w.next_id
	w.next_id++
	w.id_lock.Unlock()

	p := Package{
		ID:     this_id,
		Status: Queued,
		job:    j,
	}

	w.jobs.Set(p)
	w.queue.Add(p)

	w.emit(jobAdded, w.jobs.Get(p.ID))
	w.emit(JobAdded, w.jobs.Get(p.ID))
}

func (w *Worker) On(e Event, cb func(...interface{})) {
	if _, exists := w.events[e]; !exists {
		w.events[e] = make([]func(...interface{}), 0)
	}

	w.events[e] = append(w.events[e], cb)
}

func (w *Worker) emit(e Event, arguments ...interface{}) {
	if _, exists := w.events[e]; exists {
		for _, v := range w.events[e] {
			v(arguments...)
		}
	}
}

func (w *Worker) RunUntilDone() {
	w.runUntilDone()
}

func (w *Worker) RunUntilStopped(stop_ch chan ExitCode) {
	internal_ch := make(chan ExitCode)
	go w.runUntilKilled(stop_ch, internal_ch)
	ret := <-internal_ch
	stop_ch <- ret
}

func (w *Worker) runUntilKilled(kill_ch chan ExitCode, return_ch chan ExitCode) {
	if !w.started.On() {
		w.started.Set(true)
		defer w.started.Set(false)

		exit := false

		for {
			select {
			case code := <-kill_ch:
				if code == ExitWhenDone {
					exit = true
				}

			default:
				for i := 0; i < len(w.running_jobs); i++ {
					if w.running_jobs[i].Ch() == nil {
						p := w.queue.Top()
						if p == nil {
							break
						}

						w.running_jobs[i].SetCh(make(chan int))
						go (func(i int) {
							<-w.running_jobs[i].Ch()
							w.running_jobs[i].SetCh(nil)
						})(i)
						go w.runJob(p, w.running_jobs[i].Ch())
					}
				}

				if exit && w.queue.Len() == 0 && w.running_jobs.Empty() {
					return_ch <- ExitWhenDone
					return
				}
				time.Sleep(5 * time.Millisecond)
			}
		}
	}

	return_ch <- ExitNormally
	return
}

func (w *Worker) runUntilDone() {
	if !w.started.On() {
		w.started.Set(true)
		defer w.started.Set(false)

		for w.queue.Len() > 0 || !w.running_jobs.Empty() {

			for i := 0; i < len(w.running_jobs); i++ {
				if w.running_jobs[i].Ch() == nil {
					p := w.queue.Top()
					if p == nil {
						break
					}

					w.running_jobs[i].SetCh(make(chan int))
					go (func(i int) {
						<-w.running_jobs[i].Ch()
						w.running_jobs[i].SetCh(nil)
					})(i)
					go w.runJob(p, w.running_jobs[i].Ch())
				}
			}

			time.Sleep(5 * time.Millisecond)
		}
	}
}

func (w *Worker) runJob(p *Package, return_ch chan int) {

	// log.Printf("Starting job %d", p.ID)
	job_ch := make(chan int)
	go p.job.Run(job_ch)

	p.Status = Running

	w.emit(jobStarted, w.jobs.Get(p.ID))
	w.emit(JobStarted, w.jobs.Get(p.ID))

	w.jobs.Get(p.ID).Return = <-job_ch

	// log.Printf("Job %d finished", p.ID)
	w.emit(jobFinished, w.jobs.Get(p.ID))
	w.emit(JobFinished, w.jobs.Get(p.ID))

	return_ch <- 1
}

func (w *Worker) jobFinished(args ...interface{}) {
	pk := args[0].(*Package)

	pk.Status = Finished
}

func (w Worker) Stats() (stats WorkerStats) {
	for _, p := range w.jobs.jobs {
		switch p.Status {
		case Queued:
			stats.Queued++
		case Running:
			stats.Running++
		case Finished:
			stats.Finished++
		case Errored:
			stats.Errored++
		}

		stats.Total++
	}

	return
}

func (p Package) Job() Job {
	return p.job
}
