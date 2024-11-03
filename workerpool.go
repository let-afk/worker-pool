package main

import (
	"fmt"
	"sync"
	"time"
)

type WorkerPool struct {
	jobQueue         chan string
	workers          []*Worker
	addWorkerChan    chan struct{}
	removeWorkerChan chan struct{}
	mu               sync.Mutex
}

type Worker struct {
	id       int
	jobQueue chan string
	stopChan chan struct{}
}

func NewWorkerPool() *WorkerPool {
	pool := &WorkerPool{
		jobQueue:         make(chan string),
		addWorkerChan:    make(chan struct{}),
		removeWorkerChan: make(chan struct{}),
	}
	go pool.manageWorkers()
	return pool
}

func (pool *WorkerPool) manageWorkers() {
	for {
		select {
		case <-pool.addWorkerChan:
			pool.mu.Lock()
			id := len(pool.workers) + 1
			w := &Worker{
				id:       id,
				jobQueue: pool.jobQueue,
				stopChan: make(chan struct{}),
			}
			pool.workers = append(pool.workers, w)
			go w.start()
			fmt.Printf("Worker %d added\n", id)
			pool.mu.Unlock()

		case <-pool.removeWorkerChan:
			pool.mu.Lock()
			if len(pool.workers) > 0 {
				w := pool.workers[len(pool.workers)-1]
				close(w.stopChan)
				pool.workers = pool.workers[:len(pool.workers)-1]
				fmt.Printf("Worker %d removed\n", w.id)
			}
			pool.mu.Unlock()
		}
	}
}

func (w *Worker) start() {
	for {
		select {
		case job := <-w.jobQueue:
			fmt.Printf("Worker %d processing job: %s\n", w.id, job)
			time.Sleep(time.Second)

		case <-w.stopChan:
			fmt.Printf("Worker %d stopped\n", w.id)
			return
		}
	}
}

func (pool *WorkerPool) AddJob(job string) {
	pool.jobQueue <- job
}

func (pool *WorkerPool) AddWorker() {
	pool.addWorkerChan <- struct{}{}
}

func (pool *WorkerPool) RemoveWorker() {
	pool.removeWorkerChan <- struct{}{}
}
