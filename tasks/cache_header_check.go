package tasks

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/global"
	"github.com/KiraCore/interx/log"
)

// CacheHeaderCheck is a function to check cache headers if it's expired.
func CacheHeaderCheck(rpcAddr string, isLog bool) {

	log.CustomLogger().Info("`CacheHeaderCheck` Starting functionto to check cache headers if it's expired.")

	for {
		err := filepath.Walk(config.GetResponseCacheDir(),
			func(path string, info os.FileInfo, err error) error {
				if _, err := os.Stat(path); os.IsNotExist(err) {
					return nil
				}

				if err != nil {
					return err
				}

				// check file header, delete empty directory or expired cache
				delete := false

				if info.IsDir() {
					files, err := ioutil.ReadDir(path)
					if err == nil && len(files) == 0 {
						delete = true
					}
				} else if info.Size() == 0 || info.ModTime().Add(time.Duration(config.Config.Cache.CachingDuration)*time.Second).Before(time.Now().UTC()) {
					delete = true
				}

				if path != config.GetResponseCacheDir() && delete {
					if isLog {
						log.CustomLogger().Info("`CacheHeaderCheck` `cache` Deleting file.",
							"path", path,
						)
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
							log.CustomLogger().Info("`CacheHeaderCheck` `cache` Error deleting file.",
								"error", err,
							)
						}
						return err
					}

					return nil
				}

				return nil
			})

		if err != nil {
			log.CustomLogger().Error("[CacheHeaderCheck] Failed to walk for each file or directory in the tree.",
				"error", err,
			)
		}

		log.CustomLogger().Info("`CacheHeaderCheck` Finished functionto to check cache headers if it's expired.")

		time.Sleep(2 * time.Second)
	}
}
