package db

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/georgewheatcroft/simple-pass/internal/item"
	"github.com/georgewheatcroft/simple-pass/internal/store"
	log "github.com/sirupsen/logrus"
)

var (
	ErrFileAlreadyExists              = errors.New("file already exists")
	ErrDBNameMismatch                 = errors.New("passdb name does not match expected name")
	ErrInvalidItem                    = errors.New("item is invalid")
	ErrItemDoesNotExist               = errors.New("item does not exist in the passdb")
	ErrNilItem                        = errors.New("nil was returned from storage for the item")
	ErrItemUnchanged                  = errors.New("item was unchanged")
	ErrCannotRenameToExistingItemName = errors.New("item cannot be renamed to the name it already has")
	ErrItemNameAlreadyInUse           = errors.New("cannot rename item - name already in use by another item")
)

// TODO perhaps turn this into an interface and have different types (e.g. file db, remote db,  etc) implement this
type PassDB struct {
	store *store.Store
	// this is very specific to a file or webserver related implementation - TODO some type infront of this?
	path string
}

// createDBLocalFile creates the file for the  file based db on local disk
func createDBLocalFile(path string) (*os.File, error) {
	_, err := os.Stat(path)
	if err == nil {
		return nil, ErrFileAlreadyExists
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	/* #nosec */
	fh, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return fh, nil
}

func serialiseItem(passItem *item.Item) (string, error) {
	serialisedBytes, err := json.Marshal(passItem)
	if err != nil {
		return "", err
	}
	return string(serialisedBytes), nil
}

func CreatePassDB(path, dbName, dbPassword string) (*PassDB, error) {
	fh, err := createDBLocalFile(path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	createdStore, err := store.CreateStore(fh, dbName, dbPassword)
	if err != nil {
		return nil, err
	}
	return &PassDB{store: createdStore, path: path}, nil
}

func LoadExistingPassDB(path, password string) (*PassDB, error) {
	/* #nosec */
	fh, err := os.OpenFile(path, os.O_RDONLY, 0o600)
	if err != nil {
		return nil, err
	}
	loadedStore, err := store.Load(fh, password)
	if err != nil {
		return nil, err
	}
	log.Debugf("retrieved store: %v", *loadedStore)
	return &PassDB{store: loadedStore, path: path}, nil
}

// ListAllItems returns a slice of strings for all names of items which exist
func (db *PassDB) ListAllItems() []string {
	ret := []string{}
	keyVals := db.store.GetAllStoreDataKeyValues()
	for key := range keyVals {
		ret = append(ret, key)
	}
	return ret
}

func (db *PassDB) GetPassDBName() string {
	return db.store.GetStoreName()
}

// SaveNewItem writes new items to db storage
func (db *PassDB) SaveNewItem(passItem *item.Item) error {
	if passItem == nil {
		return ErrInvalidItem
	}
	serialisedItem, err := serialiseItem(passItem)
	if err != nil {
		log.Debugf("failed to serialise item:%s", err)
		return err
	}

	err = db.store.CreateStoreDataKeyValue(passItem.Name, serialisedItem)
	if err != nil {
		log.Debugf("failed to create new db storedata key:%s", err)
		return err
	}
	return db.commit()
}

// RetrieveItem returns items which are stored in the db
func (db *PassDB) RetrieveItem(itemName string) (*item.Item, error) {
	serialisedItem, err := db.store.GetStoreDataKeyValue(itemName)
	if err != nil {
		if errors.Is(err, store.ErrStoreDataKeyDoesNotExist) {
			return nil, ErrItemDoesNotExist
		}
		return nil, err
	}
	if serialisedItem == "" {
		return nil, ErrNilItem
	}
	var retItem item.Item
	err = json.Unmarshal([]byte(serialisedItem), &retItem)
	if err != nil {
		log.Debugf("failed to retrieve item:%s - %s", itemName, err)
		return nil, err
	}
	return &retItem, nil
}

// UpdateItem updates the item stored for a given item name, if it exists in the db
func (db *PassDB) UpdateItem(passItem *item.Item) error {
	if passItem == nil {
		return ErrInvalidItem
	}
	_, err := db.store.GetStoreDataKeyValue(passItem.Name)
	if err != nil {
		if errors.Is(err, store.ErrStoreDataKeyDoesNotExist) {
			return ErrItemDoesNotExist
		}
		return err
	}
	serialised, err := serialiseItem(passItem)
	if err != nil {
		return err
	}

	err = db.store.UpdateStoreDataKeyValue(passItem.Name, serialised)
	if err != nil {
		if errors.Is(err, store.ErrNoChangeToStoreDataKeyValueMade) {
			return ErrItemUnchanged
		}
	}

	return db.commit()
}

// RenameItem renames an existing item in persistent store or aborts the change
func (db *PassDB) RenameItem(current, desired string) error {
	if current == desired {
		return ErrCannotRenameToExistingItemName
	}
	//what we are renaming should already exist
	retrieved, err := db.store.GetStoreDataKeyValue(current)
	if err != nil {
		if errors.Is(err, store.ErrStoreDataKeyDoesNotExist) {
			return ErrItemDoesNotExist
		}
		return err
	}
	//we shouldn't attempt to rename an item to something that already exists
	_, err = db.store.GetStoreDataKeyValue(desired)
	if err == nil {
		return ErrItemNameAlreadyInUse
	}

	var newItem item.Item
	err = json.Unmarshal([]byte(retrieved), &newItem)
	if err != nil {
		return err
	}

	newItem.Name = desired

	serialisedItem, err := serialiseItem(&newItem)
	if err != nil {
		log.Debugf("failed to serialise item:%s", err)
		return err
	}

	err = db.store.CreateStoreDataKeyValue(desired, serialisedItem)
	if err != nil {
		log.Debugf("in rename - failed to create new db storedata key:%s", err)
		return err
	}

	err = db.store.DeleteStoreDataKey(current)
	if err != nil {
		log.Debugf("failed to remove old item name:%s", err)
		return err
	}
	//we have removed the old item name (key) and the new one exists
	return db.commit()
}

// TODO current implementation relys heavily on persistent storage medium being a local file (see below).
// extending to other persistent storage mediums will require changes - see below

// commit ensures that any changes to items are writen to file
func (db *PassDB) commit() error {
	tmpPath := db.path + ".tmp"
	/* #nosec */
	fh, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	// after changing file contents, close it and only then replace the now stale passDB with the updated passDB
	defer func() error {
		err := fh.Close()
		if err != nil {
			return err
		}
		return os.Rename(tmpPath, db.path)
	}()

	return db.store.Save(fh)
}

func (db *PassDB) DeleteItem(name string) error {
	_, err := db.store.GetStoreDataKeyValue(name)
	if err != nil {
		if errors.Is(err, store.ErrStoreDataKeyDoesNotExist) {
			return ErrItemDoesNotExist
		}
		return err
	}

	err = db.store.DeleteStoreDataKey(name)
	if err != nil {
		return err
	}
	return db.commit()
}
