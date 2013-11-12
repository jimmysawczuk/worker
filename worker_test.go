package worker

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var worker Worker

type SampleJob struct {
	Name string
	Duration time.Duration
}

func (s *SampleJob) Run(ch chan int) {

	time.Sleep(s.Duration)
	fmt.Printf("Done, slept for %s\n", s.Duration)

	ch <- 1
}

func init() {
	MaxJobs = 7
	worker = NewWorker()

	rand.Seed(time.Now().Unix())
}

func randomIntDuration(min, max int) time.Duration {
	r := int64(rand.Int31n(int32(max - min)) + int32(min))
	dur := time.Duration(int64(r)) * time.Second

	return dur
}

func randomFloatDuration(min, max float64) time.Duration {
	r := rand.Float64() * (max - min) + min
	dur := time.Duration(int64(r * 1e6)) * time.Microsecond

	return dur
}

func TestAdd(t *testing.T) {
	for i := 0; i < 60; i++ {
		j := SampleJob{}
		worker.Add(&j)
	}
}

func TestRun(t *testing.T) {
	worker.reset()

	for i := 0; i < 20; i++ {
		j := SampleJob{Name: fmt.Sprintf("Sample job %d", i+1), Duration: randomIntDuration(1, 5) }
		worker.Add(&j)
	}
	worker.RunUntilDone()

	time.Sleep(15 * time.Second)

	for i := 20; i < 35; i++ {
		j := SampleJob{Name: fmt.Sprintf("Sample job %d", i+1), Duration: randomIntDuration(1, 5)}
		worker.Add(&j)
	}
	worker.RunUntilDone()
}

func TestAddAndRun(t *testing.T) {
	worker.reset()
	worker.On(JobAdded, func(args ...interface{}) {
		p := args[0].(Package)
		job := *(p.job.(*SampleJob))

		fmt.Printf("Job %s added\n", job.Name)
	})

	for i := 0; i < 15; i++ {
		j := SampleJob{Name: fmt.Sprintf("Sample job %d", i+1), Duration: randomIntDuration(1, 5)}
		worker.Add(&j)
	}

	ch := make(chan int)
	go worker.Start(ch)

	time.Sleep(20 * time.Second)

	for i := 15; i < 30; i++ {
		j := SampleJob{Name: fmt.Sprintf("Sample job %d (delayed)", i+1), Duration: randomIntDuration(1, 5)}
		worker.Add(&j)
	}

	<-ch

	worker.RunUntilDone()
}

func TestFloatTimes(t *testing.T) {
	worker.reset()
	for i := 0; i < 15; i++ {
		j := SampleJob{Name: fmt.Sprintf("Sample job %d", i + 1), Duration: randomFloatDuration(3, 7) }
		worker.Add(&j)
	}
	worker.RunUntilDone()
}
