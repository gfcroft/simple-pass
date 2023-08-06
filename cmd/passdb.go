package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/georgewheatcroft/simple-pass/internal/common/constants"
	"github.com/georgewheatcroft/simple-pass/internal/db"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type passDBCache struct {
	Name     string
	DBPath   string
	Password string
}

const passDBCacheFileName = ".passdb"

func SetPassDBCache(dbPath, password string) error {
	cacheData := passDBCache{
		DBPath:   dbPath,
		Password: password,
	}
	// os.Create will truncate file if it already exists; allowing overwrite to occur
	fh, err := os.Create(GetPassDBCachePath())
	if err != nil {
		return err
	}
	defer fh.Close()

	serialised, err := json.Marshal(cacheData)
	if err != nil {
		return err
	}
	_, err = fh.Write(serialised)
	if err != nil {
		return err
	}

	log.Debugf("successfully set passDBCache in: %s", GetPassDBCachePath())
	return nil
}

// isLocalDevExec checks if this program execution appears to be in a local dev environment
func isLocalDevExec() bool {
	envVar, isSet := os.LookupEnv(constants.PassDBLocalDevEnvVar)
	if ! isSet {
		return false
	}
	res, err := strconv.ParseBool(envVar)
	if err != nil {
		panic(err)
	}
	return res
}

func GetPassDBCachePath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatalln("can't determine user's home directory: ", err)
	}
	homeDir := usr.HomeDir

	path := filepath.Join(homeDir, passDBCacheFileName)
	// avoid impacting local dev passDB cache during tests/dev
	if isLocalDevExec() {
		path += ".dev"
	}
	return path
}

// getPassDBPath reads the current active passdb location path
func getPassDBPath() string {
	path := []byte{}
	resortToPromptFn := func(err error) string {
		log.Warningf("cannot get passDB path from cache - %s", err)
		log.Println("Enter the path to the passDB cache: ")
		_, scanErr := fmt.Scanln(&path)
		if scanErr != nil {
			panic(scanErr)
		}
		return string(path)
	}

	cache, err := os.ReadFile(GetPassDBCachePath())
	if err != nil {
		return resortToPromptFn(err)
	}

	var deserialised passDBCache
	err = json.Unmarshal(cache, &deserialised)
	if err != nil {
		return resortToPromptFn(err)
	}

	return deserialised.DBPath
}

// getPassDBPassword reads the current active passdb password from the cache
func getPassDBPassword() string {
	pass := []byte{}
	resortToPromptFn := func(err error) string {
		log.Warningf("cannot get password from cache - %s", err)
		log.Println("Enter your password: ")
		_, scanErr := fmt.Scanln(&pass)
		if scanErr != nil {
			panic(scanErr)
		}

		return string(pass)
	}

	cache, err := os.ReadFile(GetPassDBCachePath())
	if err != nil {
		return resortToPromptFn(err)
	}

	var deserialised passDBCache
	err = json.Unmarshal(cache, &deserialised)
	if err != nil {
		return resortToPromptFn(err)
	}

	return deserialised.Password
}

func loadPassDB(path, password string) *db.PassDB {
	passDB, err := db.LoadExistingPassDB(path, password)
	if err != nil {
		log.Fatalf("can't load pass db at %s - %s", path, err)
	}
	return passDB
}

// passDBCacheExists is a convenience method for cmds which returns an error if the cache path does not exist
func passDBCacheExistsOrErr(cmd *cobra.Command, args []string) error {
	if !passDBCacheExists() {
		return fmt.Errorf("error - passDB cache does not exist\nyou must create a new passdb or load an existing one\n")

	}

	return nil
}

func passDBCacheExists() bool {
	cachePath := GetPassDBCachePath()
	_, err := os.Stat(cachePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false
		}
		panic(fmt.Sprintf("saw fatal error when attempting to verify simple-pass cache exists: %s", err))
	}

	return true
}
