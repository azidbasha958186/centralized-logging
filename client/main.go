package main

import (
    "encoding/json"
    // "flag"
    "log"
    "math/rand"
    "net"
    "os"
    "strconv"
    "time"
)

var sysExamples = []string{
    `{"message":"<86> aiops9242 sudo: pam_unix(sudo:session): session opened for user root(uid=0) by motadata(uid=1000)"}`,
    `{"message":"<134> WIN-EQ5V3RA5F7H Microsoft-Windows-Security-Auditing: A user account was successfully logged on. Account Name: Motadata"}`,
    `{"message":"<13> aiops9242 cron[1234]: (root) CMD (run-parts /etc/cron.daily) "}`,
    `{"message":"<87> aiops9242 sshd[2234]: Failed password for invalid user admin from 10.0.0.5 port 55444 ssh2"}`,
}

func main() {
    rand.Seed(time.Now().UnixNano())
    collectorHost := os.Getenv("COLLECTOR_HOST")
    if collectorHost == "" {
        collectorHost = "localhost"
    }
    collectorPort := os.Getenv("COLLECTOR_PORT")
    if collectorPort == "" {
        collectorPort = "1514"
    }
    cnt, _ := strconv.Atoi(os.Getenv("CLIENT_COUNT"))
    if cnt == 0 {
        cnt = 2
    }
    intervalMs, _ := strconv.Atoi(os.Getenv("INTERVAL_MS"))
    if intervalMs == 0 {
        intervalMs = 1200
    }

    addr := net.UDPAddr{
        IP:   net.ParseIP("0.0.0.0"),
        Port: 0,
    }
    conn, err := net.DialUDP("udp", &addr, &net.UDPAddr{IP: net.ParseIP(collectorHost), Port: mustAtoi(collectorPort)})
    if err != nil {
        log.Fatalf("udp dial: %v", err)
    }
    defer conn.Close()

    for i := 0; i < cnt; i++ {
        go func(id int) {
            t := time.NewTicker(time.Duration(intervalMs+rand.Intn(800)) * time.Millisecond)
            defer t.Stop()
            for range t.C {
                msg := sysExamples[rand.Intn(len(sysExamples))]
                // add some dynamic fields
                payload := map[string]interface{}{}
                json.Unmarshal([]byte(msg), &payload)
                payload["client_id"] = id
                b, _ := json.Marshal(payload)
                conn.Write(b)
            }
        }(i)
    }

    select {} // run forever
}

func mustAtoi(s string) int {
    v, _ := strconv.Atoi(s)
    return v
}
