package db_test

import (
	"errors"
	"os"
	"testing"

	"github.com/georgewheatcroft/simple-pass/internal/common/constants"
	"github.com/georgewheatcroft/simple-pass/internal/db"
	"github.com/georgewheatcroft/simple-pass/internal/item"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	dbName         = "testpassdb"
	dbPassword     = "!ProbablyThis!10"
	testFileDBPath = "./this.tmp.db"
	invalidPath    = "///"
)

func init() {
	log.SetLevel(log.DebugLevel)
	// avoid overwritting local dev .passdb TODO better way
	os.Setenv(constants.PassDBLocalDevEnvVar,"True")
}


func ensureNotExists(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	return os.Remove(path)
}

func TestShouldCreateNewPassDBForValidInputs(t *testing.T) {
	//TODO more tests of more invalid inputs (table it)
	err := ensureNotExists(testFileDBPath)
	require.NoError(t, err)

	passDB, err := db.CreatePassDB(testFileDBPath, dbName, dbPassword)
	require.NoError(t, err)
	require.NotNil(t, passDB)

	invalidDB, err := db.CreatePassDB(testFileDBPath, dbName, dbPassword)
	require.ErrorIs(t, err, db.ErrFileAlreadyExists)
	require.Nil(t, invalidDB)

	invalidDB, err = db.CreatePassDB(invalidPath, dbName, dbPassword)
	require.Error(t, err)
	require.Nil(t, invalidDB)
}

func TestShouldLoadExistingValidPassDB(t *testing.T) {
	err := ensureNotExists(testFileDBPath)
	require.NoError(t, err)

	passDB, err := db.CreatePassDB(testFileDBPath, dbName, dbPassword)
	require.NoError(t, err)
	require.NotNil(t, passDB)

	loadedPassDB, err := db.LoadExistingPassDB(testFileDBPath, dbPassword)
	require.NoError(t, err)
	require.NotNil(t, loadedPassDB)

	invalidDB, err := db.LoadExistingPassDB(invalidPath, dbPassword)
	require.Error(t, err)
	require.Nil(t, invalidDB)
}

func TestShouldSaveValidItem(t *testing.T) {
	err := ensureNotExists(testFileDBPath)
	require.NoError(t, err)

	passDB, err := db.CreatePassDB(testFileDBPath, dbName, dbPassword)
	require.NoError(t, err)
	require.NotNil(t, passDB)

	validItem, err := item.NewItem("foobar", "foobar", "foobar", "foobar", nil)
	require.NoError(t, err)
	require.NotNil(t, validItem)

	err = passDB.SaveNewItem(validItem)
	require.NoError(t, err)

	invalidItem, _ := item.NewItem("", "", "", "", nil)
	err = passDB.SaveNewItem(invalidItem)
	require.ErrorIs(t, db.ErrInvalidItem, err)
}

func TestShouldRetrieveValidItem(t *testing.T) {
	err := ensureNotExists(testFileDBPath)
	require.NoError(t, err)

	passDB, err := db.CreatePassDB(testFileDBPath, dbName, dbPassword)
	require.NoError(t, err)
	require.NotNil(t, passDB)

	validItem, err := item.NewItem("foobar", "foobar", "foobar", "foobar", nil)
	require.NoError(t, err)
	require.NotNil(t, validItem)

	err = passDB.SaveNewItem(validItem)
	require.NoError(t, err)

	retrievedItem, err := passDB.RetrieveItem(validItem.Name)
	require.NoError(t, err)
	require.Exactly(t, retrievedItem, validItem)

	nilItem, err := passDB.RetrieveItem("does not exist")
	require.ErrorIs(t, err, db.ErrItemDoesNotExist)
	require.Nil(t, nilItem)
}

func TestShouldUpdateValidItem(t *testing.T) {
	err := ensureNotExists(testFileDBPath)
	require.NoError(t, err)

	passDB, err := db.CreatePassDB(testFileDBPath, dbName, dbPassword)
	require.NoError(t, err)
	require.NotNil(t, passDB)

	validItem, err := item.NewItem("foobar", "foobar", "foobar", "foobar", nil)
	require.NoError(t, err)
	require.NotNil(t, validItem)

	err = passDB.SaveNewItem(validItem)
	require.NoError(t, err)

	validItem.URL = "whatever"
	err = passDB.UpdateItem(validItem)
	require.NoError(t, err)

	nonExistant, _ := item.NewItem("does not exist", "a", "a", "a", nil)
	err = passDB.UpdateItem(nonExistant)
	require.ErrorIs(t, err, db.ErrItemDoesNotExist)
}

func TestShouldRenameValidItem(t *testing.T) {
	err := ensureNotExists(testFileDBPath)
	require.NoError(t, err)

	passDB, err := db.CreatePassDB(testFileDBPath, dbName, dbPassword)
	require.NoError(t, err)
	require.NotNil(t, passDB)

	const oldItemName = "old-name"
	validItem, err := item.NewItem(oldItemName, "foobar", "foobar", "foobar", nil)
	require.NoError(t, err)
	err = passDB.SaveNewItem(validItem)
	require.NoError(t, err)
	require.NotNil(t, validItem)
	require.EqualValues(t, oldItemName, validItem.Name)

	const newItemName = "new-name"
	err = passDB.RenameItem(oldItemName, newItemName)
	require.NoError(t, err)

	retItem, err := passDB.RetrieveItem(newItemName)
	require.NoError(t, err)
	require.NotNil(t, retItem)
	require.EqualValues(t, newItemName, retItem.Name)

	//should be that the only difference was the name - assert that changing the name means equality
	equivalentItem := *validItem
	equivalentItem.Name = newItemName
	require.Equal(t, equivalentItem, *retItem)

	//should no longer exist and fail
	_, err = passDB.RetrieveItem(oldItemName)
	require.ErrorIs(t, err, db.ErrItemDoesNotExist)

	//shouldn't be able to rename an item to the name it already has
	err = passDB.RenameItem(newItemName, newItemName)
	require.ErrorIs(t, err, db.ErrCannotRenameToExistingItemName)

	//shouldn't be able to use "" as a new item name due to constraints imposed by lower level module
	err = passDB.RenameItem(newItemName, "")
	require.Error(t, err)

	//shouldn't be able to reuse an existing item name in rename op
	otherValidItem := *validItem
	const occupiedItemName = "in-use-name"
	otherValidItem.Name = occupiedItemName
	err = passDB.SaveNewItem(&otherValidItem)
	require.NoError(t, err)

	err = passDB.RenameItem(newItemName, occupiedItemName)
	require.ErrorIs(t, err, db.ErrItemNameAlreadyInUse)

	//no change should have happened to either
	retItem, err = passDB.RetrieveItem(newItemName)
	require.NoError(t, err)
	require.Equal(t, equivalentItem, *retItem)

	otherRetItem, err := passDB.RetrieveItem(occupiedItemName)
	require.NoError(t, err)
	require.Equal(t, otherValidItem, *otherRetItem)
}

func TestShouldDeleteValidItem(t *testing.T) {
	err := ensureNotExists(testFileDBPath)
	require.NoError(t, err)

	passDB, err := db.CreatePassDB(testFileDBPath, dbName, dbPassword)
	require.NoError(t, err)
	require.NotNil(t, passDB)

	validItem, err := item.NewItem("foobar", "foobar", "foobar", "foobar", nil)
	require.NoError(t, err)
	require.NotNil(t, validItem)

	// add valid item
	err = passDB.SaveNewItem(validItem)
	require.NoError(t, err)
	require.Contains(t, passDB.ListAllItems(), validItem.Name)

	err = passDB.DeleteItem(validItem.Name)
	require.NoError(t, err)

	// check does not exist in memory
	require.NotContains(t, passDB.ListAllItems(), validItem.Name)

	// check this is also the case after loading from persistent storage medium
	newPassDB, err := db.LoadExistingPassDB(testFileDBPath, dbPassword)
	require.NoError(t, err)
	require.NotContains(t, newPassDB.ListAllItems(), validItem.Name)

	err = passDB.DeleteItem("does not exist")
	require.ErrorIs(t, err, db.ErrItemDoesNotExist)
}

func TestListAllItems(t *testing.T) {
	err := ensureNotExists(testFileDBPath)
	require.NoError(t, err)

	passDB, err := db.CreatePassDB(testFileDBPath, dbName, dbPassword)
	require.NoError(t, err)
	require.NotNil(t, passDB)

	validInputs := map[int]struct {
		name     string
		username string
		password string
		url      string
		notes    []string
	}{
		1: {
			name:     "1",
			username: "g",
			password: "a",
			url:      "whatever",
			notes:    []string{""},
		},
		2: {
			name:     "2",
			username: "a",
			password: "b",
			url:      "nonsense.com",
			notes:    nil,
		},
		3: {
			name:     "3",
			username: "c",
			password: "d",
			url:      "nonsense.com",
			notes:    nil,
		},
	}

	//setup db with valid items
	createdItems := []item.Item{}
	for _, input := range validInputs {
		newItem, err := item.NewItem(input.name, input.username, input.password, input.url, input.notes)
		require.NoError(t, err)

		err = passDB.SaveNewItem(newItem)
		require.NoError(t, err)

		createdItems = append(createdItems, *newItem)
	}

	//now test that the items are all there
	retrievedItemNames := passDB.ListAllItems()
	for _, input := range validInputs {
		inputName := input.name
		require.Contains(t, retrievedItemNames, inputName)
	}
}
