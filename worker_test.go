package worker

import (
	"math/rand"
	"testing"
	"time"

	"fmt"
	"log"
)

type SampleJob struct {
	Name     string
	Duration time.Duration
}

func (s *SampleJob) Run() {
	time.Sleep(s.Duration)
	log.Printf("%s done, slept for %s\n", s.Name, s.Duration)
}

func init() {
	MaxJobs = 7
	rand.Seed(time.Now().Unix())
}

func randomIntDuration(min, max int) time.Duration {
	r := int64(rand.Int31n(int32(max-min)) + int32(min))
	dur := time.Duration(int64(r)) * time.Second

	return dur
}

func randomFloatDuration(min, max float64) time.Duration {
	r := rand.Float64()*(max-min) + min
	dur := time.Duration(int64(r*1e6)) * time.Microsecond

	return dur
}

func TestAdd(t *testing.T) {
	worker := NewWorker()
	for i := 0; i < 60; i++ {
		j := SampleJob{}
		worker.Add(&j)
	}
}

func TestRun(t *testing.T) {
	MaxJobs = 5
	worker := NewWorker()

	for i := 0; i < 20; i++ {
		j := SampleJob{Name: fmt.Sprintf("Sample job %d", i+1), Duration: randomIntDuration(2, 6)}
		worker.Add(&j)
	}
	worker.RunUntilDone()

	dur := time.Duration(5 * time.Second)
	log.Printf("Sleeping for %s\n", dur)
	time.Sleep(dur)

	for i := 20; i < 35; i++ {
		j := SampleJob{Name: fmt.Sprintf("Sample job %d", i+1), Duration: randomFloatDuration(2, 6)}
		worker.Add(&j)
	}
	worker.RunUntilDone()
}

func TestSmallRun(t *testing.T) {
	MaxJobs = 15
	worker := NewWorker()

	for i := 0; i < 5; i++ {
		j := SampleJob{Name: fmt.Sprintf("Sample job %d", i+1), Duration: randomIntDuration(2, 6)}
		worker.Add(&j)
	}
	worker.RunUntilDone()
}

func TestRunUntilFinished(t *testing.T) {
	MaxJobs = 5
	worker := NewWorker()
	worker.On(JobAdded, func(p *Package, args ...interface{}) {
		job := p.Job().(*SampleJob)

		log.Printf("Job %s added, duration %s\n", job.Name, job.Duration)
	})

	killCh := make(chan ExitCode)
	go worker.RunUntilStopped(killCh)

	for i := 0; i < 10; i++ {
		j := SampleJob{Name: fmt.Sprintf("Sample job %d (graceful)", i+1), Duration: randomIntDuration(1, 5)}
		worker.Add(&j)
	}

	for i := 10; i < 15; i++ {
		j := SampleJob{Name: fmt.Sprintf("Sample job %d (graceful)", i+1), Duration: randomIntDuration(3, 6)}
		worker.Add(&j)
	}

	time.Sleep(5 * time.Second)
	killCh <- ExitWhenDone
	log.Printf("Exiting when done")

	code := <-killCh
	log.Printf("Exited with code %d", code)
}

func TestExtendedRun(t *testing.T) {
	MaxJobs = 5
	worker := NewWorker()
	worker.On(JobAdded, func(p *Package, args ...interface{}) {
		job := p.Job().(*SampleJob)

		log.Printf("Job %s added, duration %s\n", job.Name, job.Duration)
	})

	killCh := make(chan ExitCode)
	go worker.RunUntilStopped(killCh)

	for i := 0; i < 10; i++ {
		j := SampleJob{Name: fmt.Sprintf("Sample job %d (extended)", i+1), Duration: randomIntDuration(1, 5)}
		worker.Add(&j)
	}

	log.Printf("Sleeping 20 seconds")
	time.Sleep(20 * time.Second)

	for i := 10; i < 15; i++ {
		j := SampleJob{Name: fmt.Sprintf("Sample job %d (extended)", i+1), Duration: randomIntDuration(3, 6)}
		worker.Add(&j)
	}

	log.Printf("Sleeping 15 seconds")
	time.Sleep(15 * time.Second)

	killCh <- ExitWhenDone
	log.Printf("Exiting when done")

	code := <-killCh
	log.Printf("Exited with code %d", code)
}
