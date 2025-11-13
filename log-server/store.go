package main

import (
    "bufio"
    "context"
    "encoding/json"
    // "fmt"
    "os"
    "sort"
    "sync"
    // "time"
)

// File-backed JSON-lines store + in-memory index for queries
type Store struct {
    mu      sync.RWMutex
    file    *os.File
    writer  *bufio.Writer
    records []LogRecord
    // metrics
    total int64
}

func NewStore(filePath string) (*Store, error) {
    if err := os.MkdirAll("data", 0755); err != nil {
        return nil, err
    }
    f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
    if err != nil {
        return nil, err
    }
    s := &Store{
        file:   f,
        writer: bufio.NewWriter(f),
    }
    // load existing lines
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        var r LogRecord
        if err := json.Unmarshal(scanner.Bytes(), &r); err == nil {
            s.records = append(s.records, r)
        }
    }
    return s, nil
}

func (s *Store) Ingest(ctx context.Context, r LogRecord) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    b, _ := json.Marshal(r)
    if _, err := s.writer.Write(append(b, '\n')); err != nil {
        return err
    }
    if err := s.writer.Flush(); err != nil {
        return err
    }
    s.records = append(s.records, r)
    s.total++
    return nil
}

func (s *Store) Query(service, level, username string, isBlacklisted *bool, limit int, sortBy string) []LogRecord {
    s.mu.RLock()
    defer s.mu.RUnlock()
    out := []LogRecord{}
    for _, r := range s.records {
        if service != "" && r.Service != service {
            continue
        }
        if level != "" && r.Severity != level {
            continue
        }
        if username != "" && r.Username != username {
            continue
        }
        if isBlacklisted != nil && r.IsBlacklisted != *isBlacklisted {
            continue
        }
        out = append(out, r)
    }
    if sortBy == "timestamp" {
        sort.Slice(out, func(i, j int) bool {
            return out[i].Timestamp.Before(out[j].Timestamp)
        })
    }
    if limit > 0 && len(out) > limit {
        out = out[:limit]
    }
    return out
}

func (s *Store) Metrics() (total int64, byCategory map[string]int, bySeverity map[string]int) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    byCategory = map[string]int{}
    bySeverity = map[string]int{}
    for _, r := range s.records {
        byCategory[r.EventCategory]++
        bySeverity[r.Severity]++
    }
    return s.total, byCategory, bySeverity
}

func (s *Store) Close() error {
    s.mu.Lock()
    defer s.mu.Unlock()
    if err := s.writer.Flush(); err != nil {
        return err
    }
    return s.file.Close()
}
