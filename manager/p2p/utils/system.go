package utils

import (
	"crypto/rand"
	"encoding/hex"
	"runtime"
	"sync"
	"time"
)

var (
	lastCPUSample time.Time
	lastCPUUsage  float64
	cpuUsageMutex sync.Mutex
)

func GetCPUUsage() float64 {
	cpuUsageMutex.Lock()
	defer cpuUsageMutex.Unlock()

	if time.Since(lastCPUSample) < time.Second {
		return lastCPUUsage
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	lastCPUUsage = float64(m.NumGC) / 100.0
	lastCPUSample = time.Now()

	return lastCPUUsage
}

func GetMemoryUsage() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return float64(m.Alloc) / float64(m.Sys) * 100
}

func GenerateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
