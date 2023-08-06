package item_test

import (
	"os"
	"testing"

	"github.com/georgewheatcroft/simple-pass/internal/common/constants"
	"github.com/georgewheatcroft/simple-pass/internal/item"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)


func init() {
	log.SetLevel(log.DebugLevel)
	// avoid overwritting local dev .passdb TODO better way
	os.Setenv(constants.PassDBLocalDevEnvVar,"True")
}

func TestShouldReturnValidItemForValidInput(t *testing.T) {
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

	for inputNo, input := range validInputs {
		genItem, err := item.NewItem(input.name, input.username, input.password, input.url, input.notes)
		if err != nil {
			t.Fatalf("saw error for input %d: %s", inputNo, err)
		}
		require.Equal(t, input.name, genItem.Name)
		require.Equal(t, input.password, genItem.Password)
		require.Equal(t, input.url, genItem.URL)
		require.Equal(t, input.notes, genItem.Notes)
	}
}

// TestShouldReturnErrForInvalidInput checks inputs are valid - there are very few cases not accepted
func TestShouldReturnErrForInvalidInput(t *testing.T) {
	invalidInputs := map[int]struct {
		name        string
		username    string
		password    string
		url         string
		notes       []string
		expectedErr error
	}{
		1: {
			name:        "",
			username:    "g",
			password:    "a",
			url:         "whatever",
			notes:       []string{""},
			expectedErr: item.ErrNoItemNameSupplied,
		},
		2: {
			name:        "2",
			username:    "",
			password:    "",
			url:         "",
			notes:       nil,
			expectedErr: item.ErrInsufficientInformation,
		},
	}

	for inputNo, input := range invalidInputs {
		_, err := item.NewItem(input.name, input.username, input.password, input.url, input.notes)
		if err == nil {
			t.Fatalf("no error for input %d", inputNo)
		}
		require.ErrorIs(t, err, input.expectedErr)
	}
}
