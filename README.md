# Simple job scheduler

Job scheduler that executes jobs based on priority and then time order.
To enhance fairness, priority can be automatically adjusted based on waiting time.

## Goals
1.Isolation of child process from main process
2.Execution control of the child processes.It will suspend and resume the child processes based on their run time, priority and scheduler policy.
3.Resource-aware scheduling. Will scheduler jobs based on total number of CPUs
4.Various policies - from ensuring fairness to being aggressive to run high priority jobs asap
5.Strict state machine of job status. It can be extended to save the job states into a persistent store such that the scheduler will become more robust.
6.Use of Golang pipeline model to handle jobs and errors. The jobs will be passed in several channels and the error will go to an error sink for external handling.
7.Use process groups to control processes together
8.Affinity of the main process

qmlab 11/2017
