// Copyright 2022 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package file

import (
	"context"
	"sync"
	"time"
)

// Task represents a function to be executed by the thread pool
type Task func()

// ThreadPool implements a simple thread pool that grows and shrinks as needed
type ThreadPool struct {
	mu             sync.Mutex
	tasks          chan Task
	numIdleThreads int
	terminated     bool
	wg             sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
}

// Global thread pool instance
var Instance = NewThreadPool()

// NewThreadPool creates a new thread pool
func NewThreadPool() *ThreadPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &ThreadPool{
		tasks:      make(chan Task, 100), // Buffered channel for tasks
		terminated: false,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// PostTask finds or spawns a worker thread to handle the task
func (tp *ThreadPool) PostTask(task Task) {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	
	if tp.terminated {
		return // Don't accept new tasks if terminated
	}
	
	// Try to send task to existing idle thread
	select {
	case tp.tasks <- task:
		// Task queued successfully
		return
	default:
		// No idle thread available, spawn a new one
		tp.wg.Add(1)
		go tp.threadMain()
		
		// Queue the task
		tp.tasks <- task
	}
}

// Close terminates the thread pool and waits for all threads to finish
func (tp *ThreadPool) Close() {
	tp.mu.Lock()
	if tp.terminated {
		tp.mu.Unlock()
		return
	}
	tp.terminated = true
	tp.mu.Unlock()
	
	// Cancel context to signal all threads to exit
	tp.cancel()
	
	// Close the task channel
	close(tp.tasks)
	
	// Wait for all threads to finish
	tp.wg.Wait()
}

// waitForTask waits for a task to become available or for termination
func (tp *ThreadPool) waitForTask() (Task, bool) {
	// Wait for a task with timeout for idle thread management
	select {
	case task, ok := <-tp.tasks:
		if !ok {
			return nil, false // Channel closed
		}
		return task, true
	case <-tp.ctx.Done():
		return nil, false // Terminated
	case <-time.After(30 * time.Second):
		// Idle timeout - thread should exit to reduce pool size
		return nil, false
	}
}

// threadMain is the main function for worker threads
func (tp *ThreadPool) threadMain() {
	defer tp.wg.Done()
	
	for {
		tp.mu.Lock()
		tp.numIdleThreads++
		tp.mu.Unlock()
		
		// Wait for a task
		task, ok := tp.waitForTask()
		
		tp.mu.Lock()
		tp.numIdleThreads--
		tp.mu.Unlock()
		
		if !ok {
			// No task available or terminated
			return
		}
		
		// Execute the task
		if task != nil {
			task()
		}
	}
}

// GetNumIdleThreads returns the number of idle threads (for testing/monitoring)
func (tp *ThreadPool) GetNumIdleThreads() int {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	return tp.numIdleThreads
}

// IsTerminated returns whether the pool has been terminated
func (tp *ThreadPool) IsTerminated() bool {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	return tp.terminated
}