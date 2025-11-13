package main

import (
    "encoding/json"
    "strings"
    "time"
)

// naive parser - in real world we would use regexes / syslog parser
func ParseRaw(raw []byte) (ParsedLog, error) {
    // try as JSON with "message"
    var r map[string]interface{}
    if err := json.Unmarshal(raw, &r); err != nil {
        // fallback: treat raw as message
        msg := strings.TrimSpace(string(raw))
        return ParsedLog{
            Timestamp:  time.Now().UTC(),
            EventCategory: "unknown",
            Username:   extractUsername(msg),
            Hostname:   extractHostname(msg),
            Severity:   "INFO",
            RawMessage: msg,
        }, nil
    }
    msg := ""
    if m, ok := r["message"].(string); ok {
        msg = m
    }
    p := ParsedLog{
        Timestamp:  time.Now().UTC(),
        EventCategory: guessCategory(msg),
        Username:   extractUsername(msg),
        Hostname:   extractHostname(msg),
        Severity:   guessSeverity(msg),
        RawMessage: msg,
    }
    return p, nil
}

func extractUsername(msg string) string {
    // simple heuristics
    if strings.Contains(msg, "root") {
        return "root"
    }
    if strings.Contains(msg, "Motadata") {
        return "motadata"
    }
    return "unknown"
}
func extractHostname(msg string) string {
    // very naive
    parts := strings.Fields(msg)
    if len(parts) > 1 {
        return parts[1]
    }
    return "unknown"
}
func guessCategory(msg string) string {
    if strings.Contains(strings.ToLower(msg), "login") {
        return "login.audit"
    }
    if strings.Contains(strings.ToLower(msg), "logout") {
        return "logout.audit"
    }
    return "syslog"
}
func guessSeverity(msg string) string {
    if strings.Contains(strings.ToLower(msg), "error") || strings.Contains(strings.ToLower(msg), "fail") {
        return "ERROR"
    }
    return "INFO"
}
