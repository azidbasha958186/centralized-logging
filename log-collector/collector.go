package main

import (
    "bufio"
    "context"
    "io"
    "log"
    "net"
    "net/http"
    "os"
    "os/signal"
    "strconv"
    "syscall"
	"encoding/json"
    // "time"
)

func main() {
    udpAddr := os.Getenv("COLLECTOR_UDP_ADDR")
    if udpAddr == "" {
        udpAddr = ":1514"
    }
    tcpAddr := os.Getenv("COLLECTOR_TCP_ADDR")
    if tcpAddr == "" {
        tcpAddr = ":1515"
    }
    forwardURL := os.Getenv("LOG_SERVER_URL")
    if forwardURL == "" {
        forwardURL = "http://localhost:9000/ingest"
    }
    workers, _ := strconv.Atoi(os.Getenv("WORKER_COUNT"))
    if workers == 0 {
        workers = 4
    }
    queueSize, _ := strconv.Atoi(os.Getenv("QUEUE_SIZE"))
    if queueSize == 0 {
        queueSize = 100
    }
    pool := NewWorkerPool(workers, queueSize, forwardURL)
    ctx, cancel := context.WithCancel(context.Background())
    pool.Start(ctx)

    // UDP listener
    go func() {
        addr, err := net.ResolveUDPAddr("udp", udpAddr)
        if err != nil {
            log.Fatalf("udp resolve: %v", err)
        }
        l, err := net.ListenUDP("udp", addr)
        if err != nil {
            log.Fatalf("udp listen: %v", err)
        }
        defer l.Close()
        log.Printf("listening udp %s", udpAddr)
        buf := make([]byte, 65536)
        for {
            n, _, err := l.ReadFromUDP(buf)
            if err != nil {
                log.Printf("udp read err: %v", err)
                continue
            }
            b := make([]byte, n)
            copy(b, buf[:n])
            p, _ := ParseRaw(b)
            // enrich
            p.IsBlacklisted = isBlacklisted(p.Username)
            // enqueue with backpressure
            if ok := pool.Enqueue(p); !ok {
                // dropped
            }
        }
    }()

    // TCP listener
    go func() {
        ln, err := net.Listen("tcp", tcpAddr)
        if err != nil {
            log.Fatalf("tcp listen: %v", err)
        }
        log.Printf("listening tcp %s", tcpAddr)
        for {
            conn, err := ln.Accept()
            if err != nil {
                log.Printf("accept err: %v", err)
                continue
            }
            go handleConn(conn, pool)
        }
    }()

    // metrics server
    go func() {
        http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
            json.NewEncoder(w).Encode(pool.Metrics())
        })
        http.ListenAndServe(":8081", nil)
    }()

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
    <-stop
    log.Println("shutting down collector")
    cancel()
    pool.Shutdown()
}

func handleConn(conn net.Conn, pool *WorkerPool) {
    defer conn.Close()
    rd := bufio.NewReader(conn)
    for {
        line, err := rd.ReadBytes('\n')
        if err != nil {
            if err != io.EOF {
                log.Printf("tcp read err: %v", err)
            }
            return
        }
        p, _ := ParseRaw(line)
        p.IsBlacklisted = isBlacklisted(p.Username)
        if ok := pool.Enqueue(p); !ok {
            // dropped
        }
    }
}

// simple blacklist check
var blacklist = map[string]bool{
    "baduser": true,
    "192.168.0.10": true,
}

func isBlacklisted(username string) bool {
    return blacklist[username]
}
