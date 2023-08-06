package store

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/georgewheatcroft/simple-pass/pkg/crypt"
	log "github.com/sirupsen/logrus"
)

var (
	ErrSourceNotDefinedOnLoad          = errors.New("attempting to load store when source has not been defined for store")
	ErrLocalFileDoesNotExist           = errors.New("file does not exist in local filesystem")
	ErrStoreDataAlreadyLoaded          = errors.New("store data already loaded")
	ErrStoreDataIsEmpty                = errors.New("store data is empty")
	ErrStoreDataUnserialisable         = errors.New("store data cannot be serialised to bytes in its current form")
	ErrStoreDataKeyDoesNotExist        = errors.New("key provided does not exist in store data")
	ErrStoreDataKeyAlreadyExists       = errors.New("key provided already exists in store data")
	ErrStoreNameEmpty				   = errors.New("store name cannot be empty")
	ErrNoChangeToStoreDataKeyValueMade = errors.New("no changes to store datakey value were made")
	ErrInvalidStoreDataKey             = errors.New("key provided is invalid")
)

// LocationType informs where the store is located
type LocationType int32

// LocationTypes which  informs where the store is located
const (
	// LocalFile file on disk of the execution environment
	LocalFile LocationType = iota
	// RemoteFile file in remote location
	RemoteFile
)

// Source defines a store source that can be loaded into memory and saved to
type Source struct {
	Location     string
	LocationType LocationType
}

// Store performs storage operations on an io compatible medium containing store data
type Store struct {
	Source    *Source
	storeData *storeData
	password  string
}

// Load reads a store into memory from a given io.reader interface
func Load(r io.Reader, password string) (*Store, error) {
	sourceRetrieved, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	decrypted, err := crypt.Decrypt(sourceRetrieved, []byte(password))
	if err != nil {
		return nil, err
	}

	var storeData storeData
	err = json.Unmarshal(decrypted, &storeData)
	if err != nil {
		return nil, err
	}

	store := &Store{
		Source:    nil,
		storeData: &storeData,
		password:  password,
	}

	return store, nil
}

func (store *Store) GetAllStoreDataKeyValues() map[string]string {
	return store.storeData.Data
}

func (store *Store) GetStoreName() string {
	return store.storeData.Name
}

// Save writes storedata held in memory to an io.Writer provided
func (store *Store) Save(w io.Writer) error {
	//TODO after a certain point we may need to provide more specific errors - could  wrap these... etc
	storeData, err := json.Marshal(store.storeData)
	if err != nil {
		return err
	}

	encrypted, err := crypt.Encrypt(storeData, []byte(store.password))
	if err != nil {
		return err
	}
	_, err = w.Write(encrypted)
	if err != nil {
		return err
	}

	log.Debugf("successfully saved store data to source location")
	return nil
}

// storeData is the represenation of the simple json body that holds the stored data
// and additional metadata - if one cannot get this from the source after reading, then this  means that
// the file is corrupt or encryption has failed due to invalid secret
type storeData struct {
	Name    string `json:"name"`
	Version int64  `json:"version"`
	//TODO could do with defining some kind of abstraction here rather than just working directly with this... leave for now
	Data     map[string]string `json:"data"`
	Password string            `json:"secretKey"`
}

// getSerialisedStoreData returns the current serialised store data in memory
func (s *storeData) getSerialisedStoreData() ([]byte, error) {
	bytes, err := json.Marshal(s)
	if err != nil {
		return nil, ErrStoreDataUnserialisable
	}
	return bytes, nil
}

func (s *Store) StoreDataVersion() int64 {
	return s.storeData.Version
}

func (s *Store) StoreDataPassword() string {
	return s.storeData.Password
}

// UpdateStoreDataData performs an update to the store data held in memory,
// where store data is replaced with the input. NOTE Persisting the change requires writing this
// in memory storeData somewhere using Save
func (s *Store) ReplaceStoreData(replacement map[string]string) {
	s.storeData.Data = replacement
}

// UpdateStoreDataKeyValue updates the value for a key which exists in store data
func (s *Store) UpdateStoreDataKeyValue(key, value string) error {
	currentVal, exists := s.storeData.Data[key]
	if !exists {
		return ErrStoreDataKeyDoesNotExist
	}
	if currentVal == value {
		return ErrNoChangeToStoreDataKeyValueMade
	}
	log.Debugf("updating store data key: %s from value:'%s' to '%s'\n", key, currentVal, value)
	log.Debugf("store data before: %v\n", s.storeData.Data)
	s.storeData.Data[key] = value
	log.Debugf("updated store data:%v'\n", s.storeData.Data)
	return nil
}

// CreateStoreDataKeyValue creates an entry for a given key value pair in the storedata
// if it doesn't already exist
func (s *Store) CreateStoreDataKeyValue(key, value string) error {
	if key == "" {
		return ErrInvalidStoreDataKey
	}
	_, exists := s.storeData.Data[key]
	if exists {
		return ErrStoreDataKeyAlreadyExists
	}
	log.Debugf("creating store data key: %s with value:'%s'", key, value)
	s.storeData.Data[key] = value
	return nil
}

// GetStoreDataKeyValue retrieves the value for a given key if it exists
func (s *Store) GetStoreDataKeyValue(key string) (string, error) {
	value, exists := s.storeData.Data[key]
	if !exists {
		return "", ErrStoreDataKeyDoesNotExist
	}
	log.Debugf("for store data key: %s retrieved value:'%s'", key, value)
	return value, nil
}

// CreateStore creates a new store to hold data
func CreateStore(w io.Writer, name, password string) (*Store, error) {
	if name == "" {
		return nil, ErrStoreNameEmpty
	}
	storeData := &storeData{
		Name:     name,
		Version:  1,
		Data:     make(map[string]string),
		Password: password,
	}

	serialised, err := storeData.getSerialisedStoreData()
	if err != nil {
		log.Debugf("failed to serialise store data:%s", err)
		return nil, err
	}
	log.Debugf("serialised store data:%s", string(serialised))

	encrypted, err := crypt.Encrypt(serialised, []byte(password))
	if err != nil {
		log.Debugf("failed to encrypt serialised store data:%s", err)
		return nil, err
	}

	_, err = w.Write(encrypted)
	if err != nil {
		return nil, err
	}

	log.Debugf("created store: '%s' and wrote it to the provided io", name)

	return &Store{
		password:  password,
		storeData: storeData,
	}, nil
}

// DeleteStoreDataKey deletes a given key from the in memory data store
func (s *Store) DeleteStoreDataKey(key string) error {
	_, exists := s.storeData.Data[key]
	if !exists {
		return ErrStoreDataKeyDoesNotExist
	}
	log.Debugf("removing key '%s' from %v\n", key, s.storeData.Data)
	delete(s.storeData.Data, key)
	log.Debugf("post key removal:%v\n ", s.storeData.Data)
	return nil
}
