package tasks

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/KiraCore/interx/common"
	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/global"
	"github.com/KiraCore/interx/types"
)

// CacheDataCheck is a function to check cache data if it's expired.
func CacheDataCheck(rpcAddr string, isLog bool) {
	for {
		err := filepath.Walk(config.GetResponseCacheDir(),
			func(path string, info os.FileInfo, err error) error {
				if _, err := os.Stat(path); os.IsNotExist(err) {
					return nil
				}

				if err != nil {
					return err
				}

				delete := false

				if !info.IsDir() && info.Size() != 0 {
					// check cache json data

					global.Mutex.Lock()
					// check if file or path exists
					if _, err := os.Stat(path); os.IsNotExist(err) {
						global.Mutex.Unlock()
						return nil
					}
					data, _ := ioutil.ReadFile(path)
					global.Mutex.Unlock()

					result := types.InterxResponse{}
					err := json.Unmarshal([]byte(data), &result)

					if err == nil && common.IsCacheExpired(result) {
						delete = true
					}
				}

				if path != config.GetResponseCacheDir() && delete {
					if isLog {
						common.GetLogger().Info("[cache] Deleting file: ", path)
					}

					global.Mutex.Lock()
					// check if file or path exists
					if _, err := os.Stat(path); os.IsNotExist(err) {
						global.Mutex.Unlock()
						return nil
					}
					err := os.Remove(path)
					global.Mutex.Unlock()

					if err != nil {
						if isLog {
							common.GetLogger().Error("[cache] Error deleting file: ", err)
						}
						return err
					}

					return nil
				}

				return nil
			})

		if err != nil {
			log.Println(err)
		}

		time.Sleep(2 * time.Second)
	}
}
