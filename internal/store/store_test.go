package store_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/georgewheatcroft/simple-pass/internal/common/constants"
	"github.com/georgewheatcroft/simple-pass/internal/store"
	"github.com/georgewheatcroft/simple-pass/pkg/crypt"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	storePassword = "!Password!123"
	storeName     = "test-store"
)


func init() {
	log.SetLevel(log.DebugLevel)
	// avoid overwritting local dev .passdb TODO better way
	os.Setenv(constants.PassDBLocalDevEnvVar,"True")
}


func TestShouldCreateStoreForValidInputs(t *testing.T) {
	var storage bytes.Buffer
	writer := io.Writer(&storage)

	newStore, err := store.CreateStore(writer, storeName, storePassword)
	require.NotNil(t, newStore)
	require.NoError(t, err)

	if err != nil {
		log.Fatal("cant write outputs from store creation - ", err)
	}
}

func TestShouldLoadStoreForValidInputs(t *testing.T) {
	var storage bytes.Buffer
	nullWriter := io.Writer(&storage)

	s, err := store.CreateStore(nullWriter, storeName, storePassword)
	require.NotNil(t, s)
	require.NoError(t, err)

	const testDataKey = "this"
	const testDataValue = "this test data"

	err = s.CreateStoreDataKeyValue(testDataKey, testDataValue)
	require.NoError(t, err)
	retVal, err := s.GetStoreDataKeyValue(testDataKey)
	require.NoError(t, err)
	require.Exactly(t, testDataValue, retVal)

	//save it
	var output bytes.Buffer
	writer := io.Writer(&output)

	err = s.Save(writer)
	require.NoError(t, err)

	reader := io.Reader(&output)

	//load what we saved into a new store instance
	newStoreInstance, err := store.Load(reader, storePassword)
	require.NoError(t, err)

	// kv pairs from what we currently hold in memory
	storeKVPairs := s.GetAllStoreDataKeyValues()
	// kv pairs from what we have retrieved from memory
	newStoreKVPairs := newStoreInstance.GetAllStoreDataKeyValues()

	require.Len(t, storeKVPairs, len(newStoreKVPairs))
	for k := range storeKVPairs {
		require.Exactly(t, storeKVPairs[k], newStoreKVPairs[k])
	}
}

func TestStoreShouldReplaceValidStoreData(t *testing.T) {
	var storage bytes.Buffer
	storageWriter := io.Writer(&storage)

	newStore, err := store.CreateStore(storageWriter, storeName, storePassword)
	require.NotNil(t, newStore)
	require.NoError(t, err)

	var output bytes.Buffer
	const testDataKey = "this"
	const testDataValue = "this test data"
	testData := map[string]string{testDataKey: testDataValue}
	writer := io.Writer(&output)

	newStore.ReplaceStoreData(testData)
	err = newStore.Save(writer)
	require.NoError(t, err)

	decrypted, err := crypt.Decrypt(output.Bytes(), []byte(storePassword))
	require.NoErrorf(t, err, "cant decrypt the store data that has been written to: %s", err)
	require.Contains(t, string(decrypted), testDataValue)
}

func TestShouldUpdateStoreDataKeyValueForValidInputs(t *testing.T) {
	var storage bytes.Buffer
	writer := io.Writer(&storage)

	newStore, err := store.CreateStore(writer, storeName, storePassword)
	require.NotNil(t, newStore)
	require.NoError(t, err)

	const testDataKey = "this"
	const testDataValue = "this test data"

	err = newStore.CreateStoreDataKeyValue(testDataKey, testDataValue)
	require.NoError(t, err)

	retVal, err := newStore.GetStoreDataKeyValue(testDataKey)
	require.NoError(t, err)
	require.Exactly(t, testDataValue, retVal)

	const newTestDataValue = "different test data"
	err = newStore.UpdateStoreDataKeyValue(testDataKey, newTestDataValue)
	require.NoError(t, err)

	retVal, err = newStore.GetStoreDataKeyValue(testDataKey)
	require.NoError(t, err)
	require.Exactly(t, newTestDataValue, retVal)
}

func TestShouldDeleteStoreDataKey(t *testing.T) {
	var storage bytes.Buffer
	writer := io.Writer(&storage)

	newStore, err := store.CreateStore(writer, storeName, storePassword)
	require.NotNil(t, newStore)
	require.NoError(t, err)

	const testDataKey = "this"
	const testDataValue = "this test data"

	err = newStore.CreateStoreDataKeyValue(testDataKey, testDataValue)
	require.NoError(t, err)

	err = newStore.DeleteStoreDataKey(testDataKey)
	require.NoError(t, err)

	retVal, err := newStore.GetStoreDataKeyValue(testDataKey)
	require.ErrorIs(t, store.ErrStoreDataKeyDoesNotExist, err)
	require.Exactly(t, retVal, "")

	err = newStore.DeleteStoreDataKey("does not exit")
	require.ErrorIs(t, store.ErrStoreDataKeyDoesNotExist, err)
}

func TestShouldErrLoadInvalidStoreData(t *testing.T) {
	var invalid bytes.Buffer
	writer := io.Writer(&invalid)
	writer.Write([]byte("foobar"))

	s, err := store.Load(&invalid, "foo")
	require.Error(t, err)
	require.Nil(t, s)
}

func TestShouldErrWriteInvalidStoreData(t *testing.T) {
	var storage bytes.Buffer
	nullWriter := io.Writer(&storage)

	newStore, err := store.CreateStore(nullWriter, storeName, storePassword)
	require.NotNil(t, newStore)
	require.NoError(t, err)

	// setup a valid entry
	var output bytes.Buffer
	const testDataKey = "this"
	const testDataValue = "this test data"
	testData := map[string]string{testDataKey: testDataValue}
	writer := io.Writer(&output)

	newStore.ReplaceStoreData(testData)
	err = newStore.Save(writer)
	require.NoError(t, err)

	// test errors related to write operations
	err = newStore.CreateStoreDataKeyValue(testDataKey, testDataValue)
	require.ErrorIs(t, err, store.ErrStoreDataKeyAlreadyExists)

	err = newStore.UpdateStoreDataKeyValue(testDataKey, testDataValue)
	require.ErrorIs(t, err, store.ErrNoChangeToStoreDataKeyValueMade)

	err = newStore.UpdateStoreDataKeyValue("does-not-exist", testDataValue)
	require.ErrorIs(t, err, store.ErrStoreDataKeyDoesNotExist)

	err = newStore.CreateStoreDataKeyValue("", testDataValue)
	require.ErrorIs(t, err, store.ErrInvalidStoreDataKey)
}
