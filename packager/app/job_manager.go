// Copyright 2017 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package app

import (
	"sync"

	"github.com/SoulPancake/shaka-packager/packager/status"
)

// OnCompleteFunction is called when a job completes
type OnCompleteFunction func(*Job)

// Job represents a single line of work that runs in parallel with other jobs
type Job struct {
	name        string
	work        WorkHandler
	onComplete  OnCompleteFunction
	status      *status.Status
	thread      *sync.WaitGroup
	cancelChan  chan bool
	initialized bool
	started     bool
	mu          sync.RWMutex
}

// WorkHandler interface represents work that can be executed
type WorkHandler interface {
	Initialize() *status.Status
	Run() *status.Status
	Cancel()
}

// NewJob creates a new job
func NewJob(name string, work WorkHandler, onComplete OnCompleteFunction) *Job {
	return &Job{
		name:       name,
		work:       work,
		onComplete: onComplete,
		status:     status.NewOkStatus(),
		cancelChan: make(chan bool, 1),
		thread:     &sync.WaitGroup{},
	}
}

// Initialize initializes the work object. Call before Start() or Run()
func (j *Job) Initialize() *status.Status {
	j.mu.Lock()
	defer j.mu.Unlock()
	
	if j.initialized {
		return j.status
	}
	
	if j.work != nil {
		j.status = j.work.Initialize()
	}
	
	j.initialized = true
	return j.status
}

// Start begins the job in a new thread. This is a request and will not block
func (j *Job) Start() {
	j.mu.Lock()
	defer j.mu.Unlock()
	
	if j.started || !j.initialized {
		return
	}
	
	j.started = true
	j.thread.Add(1)
	
	go func() {
		defer j.thread.Done()
		defer func() {
			if j.onComplete != nil {
				j.onComplete(j)
			}
		}()
		
		j.status = j.work.Run()
	}()
}

// Run executes the job's work synchronously, blocking until complete
func (j *Job) Run() *status.Status {
	j.mu.Lock()
	defer j.mu.Unlock()
	
	if j.started || !j.initialized {
		return j.status
	}
	
	j.started = true
	
	if j.work != nil {
		j.status = j.work.Run()
	}
	
	if j.onComplete != nil {
		j.onComplete(j)
	}
	
	return j.status
}

// Cancel requests that the job stops executing
func (j *Job) Cancel() {
	j.mu.RLock()
	defer j.mu.RUnlock()
	
	if j.work != nil {
		j.work.Cancel()
	}
	
	// Signal cancellation
	select {
	case j.cancelChan <- true:
	default:
	}
}

// Join waits for the job thread to complete
func (j *Job) Join() {
	j.thread.Wait()
}

// Status returns the current status of the job
func (j *Job) Status() *status.Status {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.status
}

// Name returns the job name
func (j *Job) Name() string {
	return j.name
}

// JobManager manages multiple jobs
type JobManager struct {
	mu   sync.RWMutex
	jobs []*Job
}

// NewJobManager creates a new job manager
func NewJobManager() *JobManager {
	return &JobManager{
		jobs: make([]*Job, 0),
	}
}

// AddJob adds a job to the manager
func (jm *JobManager) AddJob(job *Job) {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	jm.jobs = append(jm.jobs, job)
}

// InitializeJobs initializes all jobs
func (jm *JobManager) InitializeJobs() *status.Status {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	
	for _, job := range jm.jobs {
		if stat := job.Initialize(); !stat.Ok() {
			return stat
		}
	}
	
	return status.NewOkStatus()
}

// RunJobs starts all jobs and waits for completion
func (jm *JobManager) RunJobs() *status.Status {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	
	// Start all jobs
	for _, job := range jm.jobs {
		job.Start()
	}
	
	// Wait for all jobs to complete
	for _, job := range jm.jobs {
		job.Join()
		if stat := job.Status(); !stat.Ok() {
			return stat
		}
	}
	
	return status.NewOkStatus()
}

// RunJobsSingleThreaded runs all jobs sequentially in a single thread
func (jm *JobManager) RunJobsSingleThreaded() *status.Status {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	
	for _, job := range jm.jobs {
		if stat := job.Run(); !stat.Ok() {
			return stat
		}
	}
	
	return status.NewOkStatus()
}

// CancelJobs cancels all running jobs
func (jm *JobManager) CancelJobs() {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	
	for _, job := range jm.jobs {
		job.Cancel()
	}
}

// GetJobs returns a copy of the job list
func (jm *JobManager) GetJobs() []*Job {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	
	jobs := make([]*Job, len(jm.jobs))
	copy(jobs, jm.jobs)
	return jobs
}