package database

import (
	"github.com/KiraCore/interx/config"
	"github.com/sonyarouje/simdb/db"
)

// Layer2Data is a struct for layer2 details.
type Layer2Data struct {
	Id   string `json:"id"`
	Data string `json:"data"`
}

// ID is a field for facuet claim struct.
func (c Layer2Data) ID() (jsonField string, value interface{}) {
	value = c.Id
	jsonField = "id"
	return
}

func LoadLayer2DbDriver() {
	DisableStdout()
	driver, _ := db.New(config.GetDbCacheDir() + "/layer2")
	EnableStdout()

	layer2Db = driver
}

// GetLayer2State is a function to get layer2 app state stored
func GetLayer2State(id string) (string, error) {
	if layer2Db == nil {
		panic("cache dir not set")
	}

	data := Layer2Data{}

	DisableStdout()
	err := layer2Db.Open(Layer2Data{}).Where("id", "=", id).First().AsEntity(&data)
	EnableStdout()

	if err != nil {
		return "", err
	}

	return data.Data, nil
}

// GetAllLayer2s is a function to get all layer2Times
func GetAllLayer2s() []interface{} {
	if layer2Db == nil {
		panic("cache dir not set")
	}

	DisableStdout()
	rawData := layer2Db.Open(Layer2Data{}).RawArray()
	EnableStdout()

	return rawData
}

// SetLayer2State is a function to set layer2 app status
func SetLayer2State(id string, data string) {
	if layer2Db == nil {
		panic("cache dir not set")
	}

	DisableStdout()
	err := layer2Db.Open(Layer2Data{}).Insert(Layer2Data{
		Id:   id,
		Data: data,
	})
	EnableStdout()

	if err != nil {
		panic(err)
	}
}

var (
	layer2Db *db.Driver
)
