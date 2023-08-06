package crypt_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/georgewheatcroft/simple-pass/internal/common/constants"
	"github.com/georgewheatcroft/simple-pass/pkg/crypt"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const validPassword = "ProbablyThis%1-0_2"
const validInput = "a"



func init() {
	log.SetLevel(log.DebugLevel)
	// avoid overwritting local dev .passdb TODO better way
	os.Setenv(constants.PassDBLocalDevEnvVar,"True")
}


func generateBasicCharStr() string {
	var sb strings.Builder
	for i := 0; i < 256; i++ {
		sb.WriteByte(byte(i))
	}
	charStr := sb.String()
	return charStr
}

func TestShouldEncryptValidInputs(t *testing.T) {
	charStr := generateBasicCharStr()
	inputs := []struct {
		password    []byte
		input       []byte
		expectedErr error
	}{
		{
			password:    []byte(validPassword),
			input:       []byte("a"),
			expectedErr: nil,
		},
		{
			password:    []byte(validPassword),
			input:       []byte(charStr),
			expectedErr: nil,
		},
		{
			password:    []byte(validPassword),
			input:       []byte(""),
			expectedErr: crypt.ErrEmptyInputText,
		},
	}
	for _, input := range inputs {
		_, err := crypt.Encrypt(input.input, input.password)
		if input.expectedErr != nil {
			require.ErrorIs(t, err, input.expectedErr)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestShouldEncryptUsingValidPassword(t *testing.T) {
	const invalidEmptyPassword = ""
	const invalidShortPassword = "1"
	inputs := []struct {
		password    []byte
		input       []byte
		expectedErr error
	}{
		{
			password:    []byte(validPassword),
			input:       []byte(validInput),
			expectedErr: nil,
		},
		{
			password:    []byte(invalidEmptyPassword),
			input:       []byte(validInput),
			expectedErr: crypt.ErrInvalidPassword,
		},
		{
			password:    []byte(invalidShortPassword),
			input:       []byte(validInput),
			expectedErr: crypt.ErrInvalidPassword,
		},
	}
	for _, input := range inputs {
		_, err := crypt.Encrypt(input.input, input.password)
		if input.expectedErr != nil {
			require.ErrorIs(t, err, input.expectedErr)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestShouldDecryptUsingValidPassword(t *testing.T) {
	const wrongPassword = validPassword + "f"
	inputs := []struct {
		encryptPassword []byte
		decryptPassword []byte
		input           []byte
		expectedErr     error
	}{
		{
			encryptPassword: []byte(validPassword),
			decryptPassword: []byte(validPassword),
			input:           []byte(generateBasicCharStr()),
			expectedErr:     nil,
		},
		{
			encryptPassword: []byte(validPassword),
			decryptPassword: []byte(wrongPassword),
			input:           []byte(generateBasicCharStr()),
			expectedErr:     crypt.ErrCannotDecrypt,
		},
		{
			encryptPassword: []byte(validPassword),
			decryptPassword: []byte(validPassword),
			input:           []byte(`{"name":"eg","version":1,"data":{},"secretKey":"ProbablyThis%1-0_2"}`),
			expectedErr:     nil,
		},
		{
			encryptPassword: []byte(validPassword),
			decryptPassword: []byte(validPassword),
			input:           []byte(`{"name":"eg","version":1,"data":{"you":{"Name":"you","ID":"91c34dfc-4112-4d60-a24a-169d97a113b3","Username":"george","Password":"me","URL":"","Notes":null}}},"secretKey":"ProbablyThis%1-0_2"}`),
			expectedErr:     nil,
		},
	}
	for _, input := range inputs {
		encrypted, err := crypt.Encrypt(input.input, input.encryptPassword)
		if err != nil {
			t.Fatalf("cannot setup encrypted bytes for decrypt test:%s", err)
		}

		decrypted, err := crypt.Decrypt(encrypted, input.decryptPassword)
		if input.expectedErr != nil {
			require.ErrorIs(t, err, input.expectedErr)
		} else {
			require.NoError(t, err)
			require.Equal(t, input.input, decrypted)
		}
	}
}

/*
	benchmarks
*/

const simpleInput = "simple-case"
const mediumInput = "more-complex_12345670809\\jfdfjpdofijafjdkcaxk ,.m/.$$-~"
const longInput = mediumInput + mediumInput + mediumInput

func BenchmarkEncryption(b *testing.B) {
	inputs := []struct {
		caseName    string
		password    []byte
		input       []byte
		expectedErr error
	}{
		{
			caseName:    "simpleCase",
			password:    []byte(validPassword),
			input:       []byte(simpleInput),
			expectedErr: nil,
		},
		{
			caseName:    "mediumCase",
			password:    []byte(validPassword),
			input:       []byte(mediumInput),
			expectedErr: nil,
		},
		{
			caseName:    "longCase",
			password:    []byte(validPassword),
			input:       []byte(longInput),
			expectedErr: nil,
		},
	}
	for _, input := range inputs {
		b.Run(fmt.Sprintf("case %s", input.caseName), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := crypt.Encrypt(input.input, input.password)
				if input.expectedErr != nil {
					require.ErrorIs(b, err, input.expectedErr)
				} else {
					require.NoError(b, err)
				}
			}
		})
	}
}

func BenchmarkDecryption(b *testing.B) {
	simpleCipher, err := crypt.Encrypt([]byte(simpleInput), []byte(validPassword))
	require.NoError(b, err)

	mediumCipher, err := crypt.Encrypt([]byte(mediumInput), []byte(validPassword))
	require.NoError(b, err)

	longCipher, err := crypt.Encrypt([]byte(longInput), []byte(validPassword))
	require.NoError(b, err)

	inputs := []struct {
		caseName   string
		password   []byte
		cipherText []byte
	}{
		{
			caseName:   "simpleCase",
			password:   []byte(validPassword),
			cipherText: simpleCipher,
		},
		{
			caseName:   "mediumCase",
			password:   []byte(validPassword),
			cipherText: []byte(mediumCipher),
		},
		{
			caseName:   "longCase",
			password:   []byte(validPassword),
			cipherText: []byte(longCipher),
		},
	}
	for _, input := range inputs {
		b.Run(fmt.Sprintf("case %s", input.caseName), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := crypt.Decrypt(input.cipherText, input.password)
				require.NoError(b, err)
			}
		})
	}
}
