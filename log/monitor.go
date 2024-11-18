package log

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

// Monitor will continuously collect system information
func Monitor(interval time.Duration) {

	printLogs, err := strconv.ParseBool(os.Getenv("PrintLogs"))

	if err != nil {
		fmt.Println("[CustomLogger] Error parsing PrintLogs environment variable:", err)
	}

	if printLogs {
		for {
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			v, _ := mem.VirtualMemory()
			cpuPercent, _ := cpu.Percent(0, true)
			loadAvg, _ := load.Avg()

			log.Printf("Timestamp: %v", time.Now().Format(time.RFC3339))

			log.Printf("Memory Total: %v, Free: %v, Used Percent: %.2f%%, Active: %v",
				v.Total, v.Free, v.UsedPercent, v.Active)

			log.Printf("CPU Usage Percentage: %v%%, CPU Usage: %v", cpuPercent, runtime.NumCPU())

			// It logs the system load average for 1, 5, and 15 minutes.
			log.Printf("Load Average: 1m: %.2f, 5m: %.2f, 15m: %.2f",
				loadAvg.Load1, loadAvg.Load5, loadAvg.Load15)

			log.Printf("Memory Alloc: %v, Total Memory Alloc: %v, System Memory: %v",
				memStats.Alloc, memStats.TotalAlloc, memStats.Sys)

			time.Sleep(interval)
		}
	}

}
