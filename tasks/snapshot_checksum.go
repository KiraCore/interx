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

func calcChecksum() {
	SnapshotChecksumAvailable = false
	SnapshotLength = 0

	log.CustomLogger().Info("Starting `calcChecksum` calculating request...")

	f, err := os.Open(config.SnapshotPath())
	if err != nil {
		log.CustomLogger().Error("[CalcChecksum] Failed open snapshot file",
			"error", err,
		)
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
				log.CustomLogger().Error("[CalcChecksum] failed to read snapshot from the file",
					"error", err,
				)
				return
			}

			break
		}

		// do some other work with buf before adding it to the hasher
		// processBuffer(buf)

		h.Write(buf[:bytesRead])
		totalRead += int64(bytesRead)
	}

	SnapshotChecksumAvailable = true
	SnapshotChecksum = hex.EncodeToString(h.Sum(nil))
	SnapshotLength = totalRead
	log.CustomLogger().Info("[CalcChecksum] encoding to the string",
		"SnapshotChecksum", SnapshotChecksum,
	)
}

// CalcSnapshotChecksum is a function for syncing sekaid status.
func CalcSnapshotChecksum() {
	available := 0
	for {
		time.Sleep(time.Duration(config.Config.SnapshotInterval) * time.Millisecond)
		file, err := os.Stat(config.SnapshotPath())

		if err != nil {
			if available != 1 {
				log.CustomLogger().Error("[CalcSnapshotChecksum] Failed to describe the file. File not available.",
					"error", err,
				)
				available = 1
			}

			continue
		}

		available = 0

		if file.ModTime().Equal(SnapshotModTime) {
			continue
		}

		SnapshotModTime = file.ModTime()

		calcChecksum()
	}
}
