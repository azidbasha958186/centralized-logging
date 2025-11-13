package main

import (
    // "context"
    // "log"
    "net/http"
    // "os"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
)

func setupRouter(store *Store) *gin.Engine {
    r := gin.Default()

    r.POST("/ingest", func(c *gin.Context) {
        var rec LogRecord
        if err := c.ShouldBindJSON(&rec); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        if rec.Timestamp.IsZero() {
            rec.Timestamp = time.Now().UTC()
        }
        if err := store.Ingest(c.Request.Context(), rec); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    r.GET("/logs", func(c *gin.Context) {
        service := c.Query("service")
        level := c.Query("level")
        username := c.Query("username")
        blq := c.Query("is.blacklisted")
        var isBlacklisted *bool
        if blq != "" {
            b, _ := strconv.ParseBool(blq)
            isBlacklisted = &b
        }
        limit, _ := strconv.Atoi(c.DefaultQuery("limit", "0"))
        sortBy := c.DefaultQuery("sort", "")
        out := store.Query(service, level, username, isBlacklisted, limit, sortBy)
        c.JSON(http.StatusOK, out)
    })

    r.GET("/metrics", func(c *gin.Context) {
        total, byCat, bySev := store.Metrics()
        c.JSON(200, gin.H{
            "total": total,
            "by_category": byCat,
            "by_severity": bySev,
        })
    })

    return r
}
