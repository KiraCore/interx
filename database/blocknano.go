package database

import (
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/log"
	"github.com/sonyarouje/simdb/db"
)

// BlockNanoData is a struct for block details.
type BlockNanoData struct {
	Height    int64 `json:"height"`
	Timestamp int64 `json:"timestamp"`
}

// ID is a field for facuet claim struct.
func (c BlockNanoData) ID() (jsonField string, value interface{}) {
	value = c.Height
	jsonField = "height"
	return
}

func LoadBlockNanoDbDriver() {
	DisableStdout()
	driver, _ := db.New(config.GetDbCacheDir() + "/blocknano")
	EnableStdout()

	blockNanoDb = driver
}

// GetBlockNanoTime is a function to get blockTime
func GetBlockNanoTime(height int64) (int64, error) {

	log.CustomLogger().Info("Starting 'GetBlockNanoTime' request...")

	if blockNanoDb == nil {
		log.CustomLogger().Error(" `GetBlockNanoTime` block Nano Db is null.")
		panic("cache dir not set")
	}

	data := BlockNanoData{}

	DisableStdout()
	err := blockNanoDb.Open(BlockNanoData{}).Where("height", "=", height).First().AsEntity(&data)
	EnableStdout()

	if err != nil {
		return 0, err
	}

	log.CustomLogger().Info("Finished 'GetBlockNanoTime' request.")

	return data.Timestamp, nil
}

// AddBlockNanoTime is a function to add blockTime
func AddBlockNanoTime(height int64, timestamp int64) {

	log.CustomLogger().Info("Starting 'AddBlockNanoTime' request...")

	if blockNanoDb == nil {
		log.CustomLogger().Error(" `AddBlockNanoTime` block Nano Db is null.")
		panic("cache dir not set")
	}

	data := BlockNanoData{
		Height:    height,
		Timestamp: timestamp,
	}

	_, err := GetBlockNanoTime(height)

	if err != nil {
		DisableStdout()
		err = blockNanoDb.Open(BlockNanoData{}).Insert(data)
		EnableStdout()

		if err != nil {
			panic(err)
		}

	}

	log.CustomLogger().Info("Finished 'AddBlockNanoTime' request.")
}

var (
	blockNanoDb *db.Driver
)
