package tasks

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"time"

	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/log"
)

var (
	SnapshotChecksumAvailable bool = false
	SnapshotLength            int64
	SnapshotChecksum          string
	SnapshotModTime           time.Time
)

func calcChecksum(isLog bool) {
	SnapshotChecksumAvailable = false
	SnapshotLength = 0

	log.CustomLogger().Info("Starting 'calcChecksum' request...")

	f, err := os.Open(config.SnapshotPath())
	log.CustomLogger().Info("[calcChecksum] Opening snapshot file.",
		"Snapshot_Path", config.SnapshotPath(),
	)

	if err != nil {
		if isLog {
			log.CustomLogger().Error("[calcChecksum] Failed to open snapshot file.",
				"Snapshot_Path", config.SnapshotPath(),
				"error", err,
			)
		}

		return
	}

	defer func() {
		_ = f.Close()
	}()

	buf := make([]byte, 1024*1024)
	h := sha256.New()

	var totalRead int64 = 0
	for {
		bytesRead, err := f.Read(buf)
		if err != nil {
			if err != io.EOF {
				if isLog {
					log.CustomLogger().Error("[calcChecksum] Failed to read snapshot file.",
						"error", err,
					)
				}
				return
			}

			break
		}

		// do some other work with buf before adding it to the hasher
		// processBuffer(buf)

		h.Write(buf[:bytesRead])
		totalRead += int64(bytesRead)

		log.CustomLogger().Debug("[calcChecksum] Read chunk from snapshot file.",
			"bytesRead", bytesRead,
			"totalRead", totalRead,
		)
	}

	SnapshotChecksumAvailable = true
	SnapshotChecksum = hex.EncodeToString(h.Sum(nil))
	SnapshotLength = totalRead

	log.CustomLogger().Info("Finished 'calcChecksum' request.")
}

// CalcSnapshotChecksum is a function for syncing sekaid status.
func CalcSnapshotChecksum(isLog bool) {

	log.CustomLogger().Info("`CalcSnapshotChecksum` Starting snapshot checksum monitoring.",
		"snapshotInterval", config.Config.SnapshotInterval,
	)

	available := 0
	for {
		time.Sleep(time.Duration(config.Config.SnapshotInterval) * time.Millisecond)
		file, err := os.Stat(config.SnapshotPath())

		if err != nil {
			if available != 1 && isLog {
				log.CustomLogger().Error("[CalcSnapshotChecksum] Failed to access snapshot file.",
					"Snapshot_Path", config.SnapshotPath(),
					"error", err,
				)
				available = 1
			}

			continue
		}

		available = 0

		if file.ModTime().Equal(SnapshotModTime) {
			log.CustomLogger().Debug("`CalcSnapshotChecksum` No changes detected in snapshot file.",
				"lastModifiedTime", file.ModTime(),
			)
			continue
		}

		log.CustomLogger().Info("`CalcSnapshotChecksum` Detected changes in snapshot file.",
			"lastModifiedTime", file.ModTime(),
		)

		SnapshotModTime = file.ModTime()

		calcChecksum(isLog)

		log.CustomLogger().Info("Finished 'CalcSnapshotChecksum' request.")
	}
}
