package database

import (
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/log"
	"github.com/sonyarouje/simdb/db"
)

// BlockData is a struct for block details.
type BlockData struct {
	Height    int64 `json:"height"`
	Timestamp int64 `json:"timestamp"`
}

// ID is a field for facuet claim struct.
func (c BlockData) ID() (jsonField string, value interface{}) {
	value = c.Height
	jsonField = "height"
	return
}

func LoadBlockDbDriver() {
	DisableStdout()
	driver, _ := db.New(config.GetDbCacheDir() + "/block")
	EnableStdout()

	blockDb = driver
}

// GetBlockTime is a function to get blockTime
func GetBlockTime(height int64) (int64, error) {

	log.CustomLogger().Info("Starting 'GetBlockTime' request...")

	if blockDb == nil {
		log.CustomLogger().Error("[GetBlockTime] block db is null.",
			"height", height,
		)
		panic("cache dir not set")
	}

	data := BlockData{}

	DisableStdout()
	err := blockDb.Open(BlockData{}).Where("height", "=", height).First().AsEntity(&data)
	EnableStdout()

	if err != nil {
		return 0, err
	}

	log.CustomLogger().Info("Finished 'GetBlockTime' request.")

	return data.Timestamp, nil
}

// GetAllBlocks is a function to get all blockTimes
func GetAllBlocks() []interface{} {

	log.CustomLogger().Info("Starting 'GetAllBlocks' request...")

	if blockDb == nil {
		log.CustomLogger().Error("[GetBlockTime] block db is null.")
		panic("cache dir not set")
	}

	DisableStdout()
	rawData := blockDb.Open(BlockData{}).RawArray()
	EnableStdout()

	log.CustomLogger().Info("Finished 'GetAllBlocks' request.")

	return rawData
}

// AddBlockTime is a function to add blockTime
func AddBlockTime(height int64, timestamp int64) {

	log.CustomLogger().Info("Starting 'AddBlockTime' request...")

	if blockDb == nil {
		log.CustomLogger().Error("[AddBlockTime] block db is null.")
		panic("cache dir not set")
	}

	data := BlockData{
		Height:    height,
		Timestamp: timestamp,
	}

	_, err := GetBlockTime(height)

	if err != nil {
		DisableStdout()
		err = blockDb.Open(BlockData{}).Insert(data)
		EnableStdout()

		if err != nil {
			panic(err)
		}

	}

	log.CustomLogger().Info("Finished 'AddBlockTime' request.")
}

var (
	blockDb *db.Driver
)
