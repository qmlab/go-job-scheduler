package main

import (
	"context"
	"fmt"
	"time"

	".."
)

func main() {
	j := job.NewJob(context.Background(), "find", "/", "test.go")
	err := j.Start()
	if err != nil {
		//fmt.Printf("err:%v\n", err)
	}

	if j.GetPid() != 0 {
		fmt.Printf("Job.Start succeeded. pid=%d\n", j.GetPid())
	}
	err = j.Pause()
	if err != nil {
		fmt.Printf("Job.Pause err:%v\n", err)
	} else {
		println("Job.Pause succeeded.")
	}

	reportTime(j)
	time.Sleep(2 * time.Second)
	err = j.Resume()
	if err != nil {
		fmt.Printf("Job.Resume err:%v\n", err)
	} else {
		println("Job.Resume succeeded.\n")
	}

	for i := 0; i < 3; i++ {
		time.Sleep(1 * time.Second)
		reportTime(j)
		if !j.HasProcessExited() {
			println("Job.HasProcessExited succeeded.\n")
		}
	}

	err = j.Wait()
	if err != nil {
		//fmt.Printf("wait err:%v\n", err)
	}

	if j.HasProcessExited() {
		println("Job.Wait succeeded.\n")
	} else {
		println("Job.Wait failed.\n")
	}

	ctx, cancel := context.WithCancel(context.Background())
	j2 := job.NewJob(ctx, "find", "/", "test.go")
	j2.Start()
	if !j.IsCancelled() {
		println("Job.IsCancelled succeeded.\n")
	} else {
		println("Job.IsCancelled failed.\n")
	}

	cancel()
	if j.IsCancelled() {
		println("Job.IsCancelled succeeded.\n")
	} else {
		println("Job.IsCancelled failed.\n")
	}
}

func reportTime(j *job.Job) {
	ut, st, err := j.GetProcessCpu()
	mem := j.GetProcessMem()
	if err == nil {
		fmt.Printf("user_time:%d, system_time:%d, mem:%d\n", ut, st, mem)
	} else {
		fmt.Printf("report_time error:%v", err)
	}
}
