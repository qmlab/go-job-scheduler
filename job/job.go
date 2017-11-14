package job

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Job struct {
	Cmd      *exec.Cmd // Execution command
	Id       int       // ID
	Gid      int       // Group ID
	Priority int       // Scheduler priority
	State    JobState  // Job atate
	ctx      context.Context

	pid     int // Process ID
	wt      time.Duration
	rt      time.Duration
	retries int
}

func NewJob(ctx context.Context, binary string, args ...string) *Job {
	cmd := exec.CommandContext(ctx, binary, args...)
	return &Job{
		Id:    rand.Int(),
		Cmd:   cmd,
		State: Created,
		ctx:   ctx,
	}
}

func (j *Job) IsCancelled() bool {
	select {
	case <-j.ctx.Done():
		return true
	default:
		return false
	}
}

func (j *Job) GetPid() int {
	return j.pid
}

func (j *Job) Start() error {
	if j.Cmd == nil {
		return nil
	}

	if err := j.Cmd.Start(); err != nil {
		return err
	}

	j.pid = j.Cmd.Process.Pid
	j.State = Started
	return syscall.Setpgid(j.pid, j.Gid)
}

func (j *Job) Pause() error {
	if j.Cmd == nil {
		return nil
	}

	return syscall.Kill(j.pid, syscall.SIGSTOP)
}

func (j *Job) Resume() error {
	if j.Cmd == nil {
		return fmt.Errorf("Job error: cmd does not exist")
	}

	return syscall.Kill(j.pid, syscall.SIGCONT)
}

func (j *Job) Stop() error {
	if j.Cmd == nil {
		return fmt.Errorf("Job error: cmd does not exist")
	}

	return syscall.Kill(j.pid, syscall.SIGTERM)
}

func (j *Job) Wait() error {
	if j.Cmd == nil {
		return fmt.Errorf("Job error: cmd does not exist")
	}

	return j.Cmd.Wait()
}

func (j *Job) GetProcessCpu() (utime, stime uint64, err error) {
	if j.Cmd == nil {
		return
	}

	lines, err := readFileLines(fmt.Sprintf("/proc/%d/stat", j.pid))
	if len(lines) > 0 {
		parts := procPidStatSplit(lines[0])
		utime, stime = readUInt(parts[13]), readUInt(parts[14])
	}
	return
}

func (j *Job) GetProcessMem() uint64 {
	if j.Cmd == nil {
		return 0
	}

	var mem runtime.MemStats
	return mem.TotalAlloc
}

func (j *Job) HasProcessExited() bool {
	process, err := os.FindProcess(int(j.pid))
	if err != nil || process.Signal(syscall.Signal(0)) != nil {
		return true
	}

	return false
}

func procPidStatSplit(line string) []string {
	line = strings.TrimSpace(line)
	splitParts := make([]string, 52)
	partnum := 0
	strpos := 0
	start := 0
	inword := false
	space := " "[0]
	open := "("[0]
	close := ")"[0]
	groupchar := space

	for ; strpos < len(line); strpos++ {
		if inword {
			if line[strpos] == space && (groupchar == space || line[strpos-1] == groupchar) {
				splitParts[partnum] = line[start:strpos]
				partnum++
				start = strpos
				inword = false
			}
		} else {
			if line[strpos] == open {
				groupchar = close
				inword = true
				start = strpos
				strpos = strings.LastIndex(line, ")") - 1
				if strpos <= start { // if we can't parse this insane field, skip to the end
					strpos = len(line)
					inword = false
				}
			} else if line[strpos] != space {
				groupchar = space
				inword = true
				start = strpos
			}
		}
	}

	if inword {
		splitParts[partnum] = line[start:strpos]
		partnum++
	}

	for ; partnum < 52; partnum++ {
		splitParts[partnum] = ""
	}
	return splitParts
}

// pull a uint64 out of a string
func readUInt(str string) uint64 {
	val, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		panic(err)
	}
	return val
}

// Read a small file and split on newline
func readFileLines(filename string) ([]string, error) {
	file, err := readSmallFile(filename)
	if err != nil {
		return nil, err
	}

	// TODO - these next two lines cause more GC than I expected
	fileStr := strings.TrimSpace(string(file))
	return strings.Split(fileStr, "\n"), nil
}

// readSmallFile is like os.ReadFile but dangerously optimized for reading files from /proc.
// The file is not statted first, and the same buffer is used every time.
func readSmallFile(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		f.Close()
		return nil, err
	}

	buf := bytes.NewBuffer(make([]byte, 0, 8192))
	_, err = buf.ReadFrom(f)
	f.Close()
	return buf.Bytes(), err
}
