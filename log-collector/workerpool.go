package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"sync"
	"time"
)

type WorkerPool struct {
    queue      chan ParsedLog
    workers    int
    wg         sync.WaitGroup
    forwardURL string

    // metrics
    totalReceived int64
    totalDropped  int64
}

func NewWorkerPool(workers, queueSize int, forwardURL string) *WorkerPool {
    return &WorkerPool{
        queue:      make(chan ParsedLog, queueSize),
        workers:    workers,
        forwardURL: forwardURL,
    }
}

func (wp *WorkerPool) Start(ctx context.Context) {
    for i := 0; i < wp.workers; i++ {
        wp.wg.Add(1)
        go func(id int) {
            defer wp.wg.Done()
            client := &http.Client{Timeout: 5 * time.Second}
            for {
                select {
                case <-ctx.Done():
                    log.Printf("worker %d shutting down", id)
                    return
                case item, ok := <-wp.queue:
                    if !ok {
                        return
                    }
                    wp.processOne(client, item)
                }
            }
        }(i)
    }
}

func (wp *WorkerPool) processOne(client *http.Client, item ParsedLog) {
    b, _ := json.Marshal(item)
    req, err := http.NewRequest("POST", wp.forwardURL, bytesReader(b))
    if err != nil {
        log.Printf("forward req create err: %v", err)
        return
    }
    req.Header.Set("Content-Type", "application/json")
    resp, err := client.Do(req)
    if err != nil {
        log.Printf("forward err: %v", err)
        wp.totalDropped++
        return
    }
    resp.Body.Close()
}

func (wp *WorkerPool) Enqueue(item ParsedLog) bool {
    select {
    case wp.queue <- item:
        wp.totalReceived++
        return true
    default:
        wp.totalDropped++
        // queue full
        return false
    }
}

func (wp *WorkerPool) Shutdown() {
    close(wp.queue)
    wp.wg.Wait()
}

func (wp *WorkerPool) Metrics() map[string]interface{} {
    return map[string]interface{}{
        "workers": wp.workers,
        "queue_len": len(wp.queue),
        "total_received": wp.totalReceived,
        "total_dropped": wp.totalDropped,
    }
}

// helper to avoid importing bytes package multiple times
func bytesReader(b []byte) *bytesReaderWrapper {
    return &bytesReaderWrapper{b: b, idx: 0}
}

type bytesReaderWrapper struct {
    b   []byte
    idx int
}

func (r *bytesReaderWrapper) Read(p []byte) (int, error) {
    if r.idx >= len(r.b) {
        return 0, io.EOF
    }
    n := copy(p, r.b[r.idx:])
    r.idx += n
    return n, nil
}
