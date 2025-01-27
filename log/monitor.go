package log

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

// Monitor will continuously collect system information
func Monitor(interval time.Duration, enableLogs bool) {

	if enableLogs {
		for {
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			v, _ := mem.VirtualMemory()
			cpuPercent, _ := cpu.Percent(0, true)
			loadAvg, _ := load.Avg()

			CustomLogger().Info("################# System Resource Usage #################")

			CustomLogger().Info("Timestamp",
				"current time",
				time.Now().Format(time.RFC3339),
			)

			CustomLogger().Info("Memory Usage",
				"Memory Total", v.Total,
				"Free", v.Free,
				"Used Percent", v.UsedPercent,
				"Active", v.Active,
				"Memory Alloc", memStats.Alloc,
				"Total Memory Alloc", memStats.TotalAlloc,
				"System Memory", memStats.Sys,
			)

			CustomLogger().Info("CPU Usage",
				"Percentage", cpuPercent,
				"CPU Usage", runtime.NumCPU(),
			)

			// It logs the system load average for 1, 5, and 15 minutes.
			CustomLogger().Info("Load Average",
				"1min", loadAvg.Load1,
				"5min", loadAvg.Load5,
				"15min", loadAvg.Load15,
			)

			CustomLogger().Info("##################################")

			time.Sleep(interval)
		}
	}
}
