package database

import (
	"time"

	"github.com/KiraCore/interx/config"
	"github.com/KiraCore/interx/log"
	"github.com/sonyarouje/simdb/db"
)

var (
	faucetDb *db.Driver
)

// FaucetClaim is a struct for facuet claim.
type FaucetClaim struct {
	Address string    `json:"address"`
	Claim   time.Time `json:"claim"`
}

// ID is a field for facuet claim struct.
func (c FaucetClaim) ID() (jsonField string, value interface{}) {
	value = c.Address
	jsonField = "address"
	return
}

func LoadFaucetDbDriver() {

	log.CustomLogger().Info("`LoadFaucetDbDriver` Starting load faucet db driver...",
		"base db cache directory", config.GetDbCacheDir(),
		"path", config.GetDbCacheDir()+"/faucet",
	)

	DisableStdout()
	driver, _ := db.New(config.GetDbCacheDir() + "/faucet")
	EnableStdout()

	faucetDb = driver

	log.CustomLogger().Info("`LoadFaucetDbDriver` faucetDb is fetched",
		"faucetDb", driver,
	)
}

func isClaimExist(address string) bool {

	log.CustomLogger().Info("Starting `isClaimExist` request...",
		"address", address,
		"faucetDb", faucetDb,
	)

	if faucetDb == nil {
		log.CustomLogger().Error("[isClaimExist] Db is null.")
		panic("[isClaimExist] db not set")
	}

	data := FaucetClaim{}

	DisableStdout()
	err := faucetDb.Open(FaucetClaim{}).Where("address", "=", address).First().AsEntity(&data)
	EnableStdout()

	log.CustomLogger().Info("Finished `isClaimExist` request.",
		"find", err == nil,
	)

	return err == nil
}

func getClaim(address string) time.Time {

	log.CustomLogger().Info("Starting 'getClaim' request...")

	if faucetDb == nil {
		panic("cache dir not set")
	}

	data := FaucetClaim{}

	DisableStdout()
	err := faucetDb.Open(FaucetClaim{}).Where("address", "=", address).First().AsEntity(&data)

	EnableStdout()

	if err != nil {
		log.CustomLogger().Error(err.Error())
		panic(err)
	}

	log.CustomLogger().Info("Finished 'getClaim' request.")

	return data.Claim
}

// GetClaimTimeLeft is a function to get left time for next claim
func GetClaimTimeLeft(address string) int64 {

	log.CustomLogger().Info("Starting 'GetClaimTimeLeft' request...")

	if faucetDb == nil {
		log.CustomLogger().Error("[GetClaimTimeLeft] faucet Db is null.")
		panic("cache dir not set")
	}

	if !isClaimExist(address) {
		return 0
	}

	diff := time.Now().UTC().Unix() - getClaim(address).Unix()

	if diff > config.Config.Faucet.TimeLimit {
		return 0
	}

	log.CustomLogger().Info("Finished 'GetClaimTimeLeft' request.")

	return config.Config.Faucet.TimeLimit - diff
}

// AddNewClaim is a function to add current claim time
func AddNewClaim(address string, claim time.Time) {

	log.CustomLogger().Info("Starting `AddNewClaim` request...",
		"faucetDb", faucetDb,
		"address", address,
	)

	if faucetDb == nil {
		log.CustomLogger().Error("[AddNewClaim] faucet Db is null.")
		panic("[AddNewClaim] faucet cache dir not set")
	}

	data := FaucetClaim{
		Address: address,
		Claim:   claim,
	}

	exists := isClaimExist(address)

	log.CustomLogger().Info("`AddNewClaim` result for the searching for the claim",
		"exists", exists,
	)

	if exists {
		DisableStdout()
		err := faucetDb.Open(FaucetClaim{}).Update(data)
		EnableStdout()

		if err != nil {
			log.CustomLogger().Error("[AddNewClaim] failed to update facucet db",
				"error", err.Error(),
			)
			panic(err)
		}
	} else {
		DisableStdout()
		err := faucetDb.Open(FaucetClaim{}).Insert(data)
		EnableStdout()

		if err != nil {
			log.CustomLogger().Error("[AddNewClaim] failed to insert to facucet db",
				"error", err.Error(),
			)
			panic(err)
		}
	}

	log.CustomLogger().Info("Finished 'AddNewClaim' request. Claim is updated")
}
