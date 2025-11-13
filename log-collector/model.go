package main

import "time"

type RawMessage struct {
    Message string `json:"message"`
}

type ParsedLog struct {
    Timestamp     time.Time `json:"timestamp"`
    EventCategory string    `json:"event.category"`
    Username      string    `json:"username"`
    Hostname      string    `json:"hostname"`
    Severity      string    `json:"severity"`
    RawMessage    string    `json:"raw.message"`
    IsBlacklisted bool      `json:"is.blacklisted"`
    Service       string    `json:"service,omitempty"`
}
