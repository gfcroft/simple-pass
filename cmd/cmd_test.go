package cmd_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/georgewheatcroft/simple-pass/cmd"
	"github.com/georgewheatcroft/simple-pass/internal/common/constants"
	"github.com/georgewheatcroft/simple-pass/internal/db"
	"github.com/georgewheatcroft/simple-pass/internal/item"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

const (
	dbName         = "testpassdb"
	dbPassword     = "!ProbablyThis!10"
	testFileDBPath = "./this.tmp.db"
	invalidPath    = "///"

	testValidItemName = "test"
	testValidUsername = "test-username"
	testValidPassword = "test-password123"
	testValidNotes    = "test-notes"
	testValidURL      = "http://test.faketld"

	testValidPassDBName     = "test-passdb"
	testValidPassDBPath     = "/tmp/test-path"
	testValidPassDBCachePath     = "/tmp/passdb.dev"
	testValidPassDBPassword = "test-password123"

	testDummyPassDBContents = ""
)

func init() {
	log.SetLevel(log.DebugLevel)
	// avoid overwritting local dev .passdb TODO better way
	os.Setenv(constants.PassDBLocalDevEnvVar,"True")
}

// setupNewPassDBForCmdTesting sets up both a new passDB AND the cache (both of which are required by most cmds)
func setupNewPassDBAndPassCache() (*db.PassDB, error) {
	err := ensureNotExists(testValidPassDBPath)
	if err != nil {
		return nil, err
	}

	passDB, err := db.CreatePassDB(testValidPassDBPath, dbName, dbPassword)
	if err != nil {
		return nil, err
	}

	err = cmd.SetPassDBCache(testValidPassDBCachePath, dbPassword)
	if err != nil {
		return nil, err
	}
	return passDB, nil
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

// testCmdExecute ensures that any necessary pre-sets before command execution occur
func testCmdExecute(cobraCmd *cobra.Command) error {
	log.SetOutput(cobraCmd .OutOrStdout())
	return cobraCmd.Execute()
}

/*
* NOTE - if this starts to grow, we could potentially do with using cobra command grouping and separation of
*        commands into packages
 */

func TestAddCmdShouldAddValidItem(t *testing.T) {
	passDB, err := setupNewPassDBAndPassCache()
	require.NoError(t, err)
	require.Len(t, passDB.ListAllItems(), 0)

	cmdOutput := bytes.NewBufferString("")

	rootCmd := cmd.NewRootCmd(cmdOutput, cmdOutput)

	addCmd := cmd.NewAddCmd(passDB)
	rootCmd.AddCommand(addCmd)

	rootCmd.SetArgs([]string{cmd.AddCmdName, testValidItemName})
	addCmd.Flags().Set(cmd.UsernameFlag, testValidUsername)
	addCmd.Flags().Set(cmd.PasswordFlag, testValidPassword)
	addCmd.Flags().Set(cmd.NotesFlag, testValidNotes)
	addCmd.Flags().Set(cmd.URLFlag, testValidURL)

	err = testCmdExecute(rootCmd)
	require.NoError(t, err)

	out, err := ioutil.ReadAll(cmdOutput)
	require.NoError(t, err)

	require.Equal(t, 1, len(passDB.ListAllItems()))
	require.Equal(t, passDB.ListAllItems()[0], testValidItemName)
	expectedMsg :=fmt.Sprintf(cmd.SuccessfullyAddedMessage, testValidItemName)
	require.Contains(t, string(out), strings.Trim(expectedMsg,"\n"))
	retItem, err := passDB.RetrieveItem(passDB.ListAllItems()[0])
	require.NoError(t, err)

	require.Equal(t, testValidUsername, retItem.Username)
	require.Equal(t, testValidPassword, retItem.Password)
	require.Equal(t, testValidURL, retItem.URL)
	require.Equal(t, testValidNotes, retItem.Notes[0])
	require.Equal(t, testValidItemName, retItem.Name)
}

func TestAddCmdShouldNotAddInvalidInputs(t *testing.T) {
	passDB, err := setupNewPassDBAndPassCache()
	require.NoError(t, err)
	require.Len(t, passDB.ListAllItems(), 0)

	cmdOutput := bytes.NewBufferString("")

	rootCmd := cmd.NewRootCmd(cmdOutput, cmdOutput)

	addCmd := cmd.NewAddCmd(passDB)
	rootCmd.AddCommand(addCmd)

	// no item name
	rootCmd.SetArgs([]string{cmd.AddCmdName, ""})

	err = testCmdExecute(rootCmd)
	require.Error(t, err)

	out, err := ioutil.ReadAll(cmdOutput)

	require.Contains(t, string(out), item.ErrNoItemNameSupplied.Error())

	// valid item name... but no other attributes
	rootCmd.SetArgs([]string{cmd.AddCmdName, testValidItemName})
	// empty buffer for next run
	cmdOutput.Truncate(0)

	err = testCmdExecute(rootCmd)
	require.Error(t, err)

	out, err = ioutil.ReadAll(cmdOutput)
	require.NoError(t, err)
	require.Contains(t, string(out), item.ErrInsufficientInformation.Error())
}

func TestCreatePassDbCmdShouldCreateDbAndCacheForValidInputs(t *testing.T) {
	// no need to do this for  the cache - we just upsert for that
	err := ensureNotExists(testValidPassDBPath)
	require.NoError(t, err)

	cmdOutput := bytes.NewBufferString("")

	rootCmd := cmd.NewRootCmd(cmdOutput, cmdOutput)

	createSimplePassCmd := cmd.NewCreatePassDbCmd()
	rootCmd.AddCommand(createSimplePassCmd)

	rootCmd.SetArgs([]string{cmd.CreatePassDBCmdName, testValidPassDBName})
	createSimplePassCmd.Flags().Set(cmd.PassDBNameFlag, testValidPassDBName)
	createSimplePassCmd.Flags().Set(cmd.PassDBPasswordFlag, testValidPassword)
	createSimplePassCmd.Flags().Set(cmd.PassDBFilePathFlag, testValidPassDBPath)

	err = testCmdExecute(rootCmd)
	require.NoError(t, err)

	out, err := ioutil.ReadAll(cmdOutput)

	require.Contains(t, string(out), fmt.Sprintf(cmd.SuccessfullyCreatedPassDBMessage, testValidPassDBName, testValidPassDBPath))
	require.Contains(t, string(out), cmd.SuccessfullySetPassDBCacheMessage)

	require.FileExists(t, testValidPassDBPath)
	require.FileExists(t, cmd.GetPassDBCachePath())
}

func TestCreatePassDbCmdShouldNotCreateDbOrCacheForInvalidInputs(t *testing.T) {
	err := ensureNotExists(testValidPassDBPath)
	require.NoError(t, err)
	err = ensureNotExists(cmd.GetPassDBCachePath())
	require.NoError(t, err)

	invalidInputs := []struct {
		caseName          string
		password          string
		name              string
		setupFn           func(t *testing.T, caseName string)
		cleanupFn         func(t *testing.T, caseName string)
		path              string
		expectedErrSubStr string //TODO really should be using a const val from module here... but needs either a
		//common errs package or wrapping & typing errors across numerous levels, due to
		//location of message invocation.
		//Also some errors result from dependencies we use (e.g. cobra, os)
	}{
		{
			caseName: "noFlags",
			// expectedErrSubStr  : - depends on the cobra framework itself
		},
		{
			caseName:          "invalidPassword",
			password:          "",
			name:              testValidPassDBName,
			path:              testValidPassDBPath,
			expectedErrSubStr: "password",
		},
		{
			caseName: "invalidPath",
			password: testValidPassDBPassword,
			name:     testValidPassDBName,
			path:     "",
			// expectedErrSubStr  : - depends on the cobra framework itself
		},
		{
			caseName: "pathValidButFileExists",
			password: testValidPassDBPassword,
			name:     testValidPassDBName,
			path:     testValidPassDBPath,
			setupFn: func(t *testing.T, caseName string) {
				f, err := os.Create(testValidPassDBPath)
				require.NoErrorf(t, err, "cannot setup for test case: %s", caseName)
				_, err = f.WriteString(testDummyPassDBContents)
				require.NoError(t, err)
				err = f.Close()
				require.NoError(t, err)
			},
			cleanupFn: func(t *testing.T, caseName string) {
				err := ensureNotExists(testValidPassDBPath)
				require.NoErrorf(t, err, "cannot setup for test case: %s", caseName)
			},
			expectedErrSubStr: "exists",
		},
		{
			caseName: "invalidName",
			password: testValidPassDBPassword,
			name:     "",
			path:     testValidPassDBPath,
			// expectedErrSubStr  : - depends on the cobra framework itself
		},
	}

	for _, input := range invalidInputs {
		if input.setupFn != nil {
			input.setupFn(t, input.caseName)
		}

		cmdOutput := bytes.NewBufferString("")
		rootCmd := cmd.NewRootCmd(cmdOutput, cmdOutput)

		createSimplePassCmd := cmd.NewCreatePassDbCmd()
		rootCmd.AddCommand(createSimplePassCmd)

		rootCmd.SetArgs([]string{cmd.CreatePassDBCmdName, cmd.CreatePassDBCmdName})
		createSimplePassCmd.Flags().Set(cmd.PassDBNameFlag, input.name)
		createSimplePassCmd.Flags().Set(cmd.PassDBPasswordFlag, input.password)
		createSimplePassCmd.Flags().Set(cmd.PassDBFilePathFlag, input.path)

		err = testCmdExecute(rootCmd)
		require.Errorf(t, err,"did not see expected err in test case %s",input.caseName)
		if input.expectedErrSubStr != "" {
			require.Containsf(t, err.Error(), input.expectedErrSubStr, "error: '%s' does not contain: '%s'", err.Error(), input.expectedErrSubStr)
		}
		// if there is a file at the given path, it should be the dummy file placed there in setup to create a
		// fileExists err
		if _, err := os.Stat(input.path); err == nil {
			contents, err := os.ReadFile(input.path)
			require.NoError(t, err)
			//TODO some failure cases may leave an empty passdb on disk - need to avoid this and clean up somehow
			require.Equalf(t, testDummyPassDBContents, string(contents), "legitimate passdb potentially created in %s", input.caseName)
		}

		require.NoFileExists(t, cmd.GetPassDBCachePath())
		if input.cleanupFn != nil {
			input.cleanupFn(t, input.caseName)
		}
	}
}

func TestStatusCmdShouldReturnCorrectStatus(t *testing.T) {
	// ensure that status still reports even when no passdb is setup yet
	// init both with no passdb
	err := ensureNotExists(testValidPassDBPath)
	require.NoError(t, err)

	cmdOutput := bytes.NewBufferString("")

	rootCmd := cmd.NewRootCmd(cmdOutput, cmdOutput)

	statusCmd := cmd.NewStatusCmd(nil)

	rootCmd.AddCommand(statusCmd)

	rootCmd.SetArgs([]string{cmd.StatusCmdName})

	err = testCmdExecute(rootCmd)
	require.NoError(t, err)

	out, err := ioutil.ReadAll(cmdOutput)
	require.NoError(t, err)
	require.Contains(t, string(out), cmd.PassDBNotLoadedMsg)

	//ensure that status reports name and number of items properly
	cmdOutput.Truncate(0)
	rootCmd.RemoveCommand(statusCmd) //must replace with one which has a passDB configured

	passDB, err := setupNewPassDBAndPassCache()
	require.NoError(t, err)
	require.Len(t, passDB.ListAllItems(), 0)

	rootCmd.AddCommand(
		cmd.NewStatusCmd(passDB),
	)

	rand.Seed(time.Now().UnixNano())
	totalItems := 0
	for i := 0; i > rand.Intn(100-7)+7; i++ {
		id := fmt.Sprintf("%d", i)
		newItem, err := item.NewItem(id, id, id, id, []string{id})
		require.NoError(t, err)
		err = passDB.SaveNewItem(newItem)
		require.NoError(t, err)

		totalItems++
	}

	err = testCmdExecute(rootCmd)
	require.NoError(t, err)

	out, err = ioutil.ReadAll(cmdOutput)
	require.NoError(t, err)

	// right db name
	require.Contains(t, string(out), dbName)
	// right number of items present somewhere
	require.Contains(t, string(out), fmt.Sprintf("%d", totalItems))
}

func TestGetCmdShouldReturnValidItems(t *testing.T) {
	passDB, err := setupNewPassDBAndPassCache()
	require.NoError(t, err)

	cmdOutput := bytes.NewBufferString("")
	rootCmd := cmd.NewRootCmd(cmdOutput, cmdOutput)

	newItem, err := item.NewItem("test-item-name", "test-item-username", "test-item-password", "test-item-url", []string{"test-item-notes"})
	require.NoError(t, err)
	err = passDB.SaveNewItem(newItem)
	require.NoError(t, err)

	rootCmd.AddCommand(
		cmd.NewGetCmd(passDB),
	)

	rootCmd.SetArgs([]string{cmd.GetCmdName, newItem.Name})
	err = testCmdExecute(rootCmd)
	require.NoError(t, err)

	out, err := ioutil.ReadAll(cmdOutput)
	require.NoError(t, err)
	require.Contains(t, string(out), newItem.Name)
	require.Contains(t, string(out), newItem.Password)
	require.Contains(t, string(out), newItem.URL)
	require.Contains(t, string(out), newItem.Notes[0])

	//ensure that get fails when item does not exist
	cmdOutput.Truncate(0)
	rootCmd.SetArgs([]string{cmd.GetCmdName, "invalid-item"})

	err = testCmdExecute(rootCmd)
	require.Error(t, err)

	out, err = ioutil.ReadAll(cmdOutput)
	require.Contains(t, string(out), cmd.ErrItemDoesNotExist.Error())
}

func TestListCmdShouldReturnValidItems(t *testing.T) {
	passDB, err := setupNewPassDBAndPassCache()
	require.NoError(t, err)

	// populate items in db
	addItems := []string{"test-case-a", "test-case-b", "test-case-c"}

	for _, addItem := range addItems {
		// just set a basic item with any other attributes - doesn't matter as list only returns item names
		newItem, err := item.NewItem(addItem, addItem, addItem, addItem, []string{addItem})
		err = passDB.SaveNewItem(newItem)
		require.NoError(t, err)
	}

	cmdOutput := bytes.NewBufferString("")
	rootCmd := cmd.NewRootCmd(cmdOutput, cmdOutput)
	listCmd := cmd.NewListCmd(passDB)

	rootCmd.AddCommand(listCmd)
	rootCmd.SetArgs([]string{cmd.ListCmdName})

	err = testCmdExecute(rootCmd)
	require.NoError(t, err)

	out, err := ioutil.ReadAll(cmdOutput)
	require.NoError(t, err)

	//output should contain them all:
	for _, item := range addItems {
		require.Contains(t, string(out), item)
	}
}

func TestUpdateCmdShouldUpdateValidItems(t *testing.T) {
	passDB, err := setupNewPassDBAndPassCache()
	require.NoError(t, err)

	// just set a basic item with any attributes - doesn't matter as list only returns item names
	newItem, err := item.NewItem(testValidItemName, testValidUsername, testValidPassword, testValidURL, []string{testValidNotes})
	err = passDB.SaveNewItem(newItem)
	require.NoError(t, err)

	cmdOutput := bytes.NewBufferString("")

	rootCmd := cmd.NewRootCmd(cmdOutput, cmdOutput)

	updateCmd := cmd.NewUpdateCmd(passDB)
	rootCmd.AddCommand(updateCmd)

	rootCmd.SetArgs([]string{cmd.UpdateCmdName, testValidItemName})
	const updatedUsername = "test-username-updated"
	updateCmd.Flags().Set(cmd.UsernameFlag, updatedUsername)

	err = testCmdExecute(rootCmd)
	require.NoError(t, err)

	//ensure that updated item matches what is expected
	retrievedItem, err := passDB.RetrieveItem(newItem.Name)
	require.NoError(t, err)
	require.Equal(t, retrievedItem.Username, updatedUsername)
}

// TODO could do with more testing to cover other cases - e.g. items other than the provided item aren't affected
func TestDeleteCmdShouldDeleteValidItems(t *testing.T) {
	passDB, err := setupNewPassDBAndPassCache()
	require.NoError(t, err)

	// just set a basic item with any attributes - doesn't matter as list only returns item names
	newItem, err := item.NewItem(testValidItemName, testValidUsername, testValidPassword, testValidURL, []string{testValidNotes})
	err = passDB.SaveNewItem(newItem)
	require.NoError(t, err)

	cmdOutput := bytes.NewBufferString("")

	rootCmd := cmd.NewRootCmd(cmdOutput, cmdOutput)

	deleteCmd := cmd.NewDeleteCmd(passDB)
	rootCmd.AddCommand(deleteCmd)

	rootCmd.SetArgs([]string{cmd.DeleteCmdName, testValidItemName})

	err = testCmdExecute(rootCmd)
	require.NoError(t, err)

	//ensure that the item is definitely gone
	_, err = passDB.RetrieveItem(newItem.Name)
	require.ErrorIs(t, err, cmd.ErrItemDoesNotExist)
}
