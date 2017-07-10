package mapreduce

import "fmt"

// schedule starts and waits for all tasks in the given phase (Map or Reduce).
func (mr *Master) schedule(phase jobPhase) {
	var ntasks int
	var nios int // number of inputs (for reduce) or outputs (for map)
	switch phase {
	case mapPhase:
		ntasks = len(mr.files)
		nios = mr.nReduce
	case reducePhase:
		ntasks = mr.nReduce
		nios = len(mr.files)
	}

	fmt.Printf("Schedule: %v %v tasks (%d I/Os)\n", ntasks, phase, nios)

	var sem = make(chan int, ntasks)

	for taskNum := 0; taskNum < ntasks; taskNum++ {

		go func (tn int) {
			//keep trying the same task until we are done
			for {
				worker := <-mr.registerChannel
				taskArgs :=  DoTaskArgs {
					mr.jobName,
					mr.files[tn],
					phase,
					tn,
					nios}

					//actual work needs to happen in a different thread
					//note call is SYNC in this lib, because it is RPC
					ok := call(worker, "Worker.DoTask", taskArgs, new(struct{}))

					//regarless of result, this work has done its job, release it back to the pool
					//note that we HAVE TO put it in a separate goroutine or it may stuck?
					go func(w string) {
						mr.registerChannel <- w
					} (worker)

					if ok {
						//this task done
						sem <- 1
						break;
					}
				}

			} (taskNum)
		}

		for taskNum := 0; taskNum < ntasks; taskNum++ {
			<- sem
		}

		fmt.Printf("Schedule: %v phase done\n", phase)
	}
