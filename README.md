# Go Job Scheduler

Job scheduler that executes jobs as forked processes based on priority and then time order.
To enhance fairness, priority can be automatically adjusted based on waiting time.

Jobs added longer than a TTL will be cancelled.

Jobs running for longer than TTR will be suspended and requeued.

Jobs can be controlled (stop/resume/pause) at any time. The scheduler will learn the change and adapt the queues to it.

Scheduler can be safely shut down using Close() call at any time. It will cancel all running forked processes.

## Goals
* Isolation of child process from main process
* Execution control of the child processes.It will suspend and resume the child processes based on their run time, priority and scheduler policy.
* Resource-aware scheduling. Will scheduler jobs based on total number of CPUs
* Various policies - from ensuring fairness to being aggressive to run high priority jobs asap
* Strict state machine of job status. It can be extended to save the job states into a persistent store such that the scheduler will become more robust.
* Use of Golang pipeline model to handle jobs and errors. The jobs will be passed in several channels and the error will go to an error sink for external handling.
* Use process groups to control processes together
* Affinity of the main process

## License

MIT License
