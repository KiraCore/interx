package tasks

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/KiraCore/interx/config"
	database "github.com/KiraCore/interx/database"
	"github.com/KiraCore/interx/global"
	"github.com/KiraCore/interx/log"
)

// RefMeta is a struct to be used for reference metadata
type RefMeta struct {
	ContentLength int64     `json:"content_length"`
	LastModified  time.Time `json:"last_modified"`
}

func getMeta(url string) (*RefMeta, error) {

	log.CustomLogger().Info("[getMeta] Starting function.")

	resp, err := http.Head(url)
	if err != nil {
		log.CustomLogger().Error("[getMeta] Failed to make a request with a specified context.",
			"error", err,
		)
		return nil, err
	}
	defer resp.Body.Close()

	contentLength, err := strconv.Atoi(resp.Header["Content-Length"][0])
	if err != nil {
		log.CustomLogger().Error("[getMeta] Failed to setup headers to the response.",
			"error", err,
		)
		return nil, err
	}

	lastModified, err := time.Parse(time.RFC1123, resp.Header["Last-Modified"][0])
	if err != nil {
		log.CustomLogger().Error("[getMeta] Failed to response modification.",
			"error", err,
		)
		return nil, err
	}

	log.CustomLogger().Info("[getMeta] Successfully completed the response generation process.",
		"function", "getMeta",
		"status", "success",
	)

	return &RefMeta{
		ContentLength: int64(contentLength),
		LastModified:  lastModified,
	}, nil
}

func saveReference(url string, path string) error {
	path = config.GetReferenceCacheDir() + "/" + path
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		global.Mutex.Lock()

		if _, err := os.Stat(filepath.Dir(path)); os.IsNotExist(err) {
			err1 := os.MkdirAll(filepath.Dir(path), 0700)
			if err1 != nil {
				return err1
			}
		}

		err = ioutil.WriteFile(path, bodyBytes, 0644)
		if err != nil {
			global.Mutex.Unlock()
			return err
		}

		global.Mutex.Unlock()
	}

	return nil
}

// DataReferenceCheck is a function to check cache data for data references.
func DataReferenceCheck(isLog bool) {

	log.CustomLogger().Info("`DataReferenceCheck` Starting function to check cache data for data references.")

	for {
		references, err := database.GetAllReferences()
		if err == nil {
			for _, v := range references {
				ref, err := getMeta(v.URL)
				if err != nil {
					continue
				}

				// Check if reference has changed (check length and last modified)
				if v.ContentLength == ref.ContentLength && ref.LastModified.Equal(v.LastModified) {
					continue
				}

				// Check the download file size limitation
				if ref.ContentLength > config.Config.Cache.DownloadFileSizeLimitation {
					continue
				}

				err = saveReference(v.URL, v.FilePath)
				if err != nil {
					log.CustomLogger().Error("[DataReferenceCheck] Failed to save reference.",
						"error", err,
					)
					continue
				}

				if isLog {
					log.CustomLogger().Info("`DataReferenceCheck` `cache` Data reference updated.")
					log.CustomLogger().Info("`DataReferenceCheck` `cache`", "key", v.Key)
					log.CustomLogger().Info("`DataReferenceCheck` `cache`", "ref", v.URL)
				}

				database.AddReference(v.Key, v.URL, ref.ContentLength, ref.LastModified, v.FilePath)
			}
		}

		log.CustomLogger().Info("`DataReferenceCheck` Finished to check cache data for data references.")

		time.Sleep(2 * time.Second)
	}
}
