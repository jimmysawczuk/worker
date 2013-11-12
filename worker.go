package worker

import (
	"fmt"
	"log"
	"sync"
	"time"
)

var MaxJobs int

type Job interface {
	Run(chan int)
}

type Package struct {
	ID     int64
	Status JobStatus

	job Job
}

type Worker struct {
	queue Queue
	jobs  map[int64]*Package

	max_jobs     int
	running_jobs Register

	next_id int64
	id_lock sync.Mutex

	started  Switch
	start_ch chan int

	events map[Event][]func(...interface{})
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
	w.jobs = make(map[int64]*Package)

	w.events = make(map[Event][]func(...interface{}))
	w.running_jobs = make(Register, MaxJobs)
}

func (w *Worker) builtInEvents() {
	w.events = make(map[Event][]func(...interface{}))

	w.On(jobFinished, w.jobFinished)
}

func (w *Worker) Add(j Job) {
	w.id_lock.Lock()
	defer w.id_lock.Unlock()

	p := Package{
		ID:     w.next_id,
		Status: Queued,
		job:    j,
	}
	w.jobs[w.next_id] = &p
	w.queue.Add(p)

	w.emit(jobAdded, p)
	w.emit(JobAdded, p)

	w.next_id++
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
	w.start_ch = make(chan int)
	go w.start(w.start_ch)
	<-w.start_ch
}

func (w *Worker) Start(ch chan int) {
	w.start_ch = make(chan int)
	go w.start(w.start_ch)
	<-w.start_ch
	ch <- 1
}

func (w *Worker) start(ch chan int) {
	if !w.started.On() {
		w.started.Set(true)

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

			time.Sleep(25 * time.Millisecond)
		}

		fmt.Println("Shutting down worker")
		w.started.Set(false)
		ch <- 1
	}
}

func (w *Worker) runJob(p *Package, return_ch chan int) {

	log.Printf("Starting job %d", p.ID)
	job_ch := make(chan int)
	go p.job.Run(job_ch)

	w.emit(jobStarted, p)
	w.emit(JobStarted, p)

	<-job_ch

	log.Printf("Job %d finished", p.ID)
	w.emit(jobFinished, p)
	w.emit(JobFinished, p)

	return_ch <- 1
}

func (w *Worker) jobFinished(args ...interface{}) {
	id := args[0].(Package).ID

	w.jobs[id].Status = Completed
}
