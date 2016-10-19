package workers

import (
	"sync"
	"time"
	// "golang.org/x/net/context"
)

//DOP degree of parallelism
var DOP int

//Worker keep a tract of number of tasks executed
type Worker struct {
	WorkerID int
}

//TaskExec task execution stat
type TaskExec struct {
	Task
	*Worker
	err     error
	start   time.Time
	Elapsed time.Duration
}

//Factory make tasks
type Factory interface {
	Make() Task
}

//Task interface
type Task interface {
	Exec() error
	PreExec(tsk TaskExec)
	PostExec(tsk TaskExec)
}

// Context controls interaction between the master and workers
// type Context struct {
// 	context.Context
// }

//Do execute tasks in parallel
func Do(DOP int, f Factory) error {
	numWorkers := DOP

	tasks := make(chan Task)
	//generate tasks
	go func() {
		for {
			task := f.Make()
			if task == nil { //no more tasks
				close(tasks)
				return
			}
			tasks <- task
		}
	}()

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	workers := []*Worker{}
	//launch workers
	for i := 0; i < numWorkers; i++ {
		w := &Worker{WorkerID: i}
		workers = append(workers, w)
		go func(w *Worker) {
			defer wg.Done()
			for tsk := range tasks {

				taskExec := TaskExec{Task: tsk, Worker: w, start: time.Now()}
				//make call to PreExec
				tsk.PreExec(taskExec)

				taskExec.err = tsk.Exec()
				taskExec.Elapsed = time.Since(taskExec.start)

				//make call to PostExec
				tsk.PostExec(taskExec)

			}
		}(w)
	}

	//wait for all workers done
	wg.Wait()

	return nil
}
