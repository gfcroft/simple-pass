package cmd

import (
	"fmt"

	"github.com/georgewheatcroft/simple-pass/internal/db"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	LoadPassDBCmdName               = "load-pass-db"
	SuccessfullyLoadedPassDBMessage = "loaded passdb at %s"
)

func NewLoadPassDbCmd() *cobra.Command {
	var (
		password string
		filePath string
	)

	cmd := &cobra.Command{
		Use:   LoadPassDBCmdName,
		Short: "loads an existing simple-pass db",
		Long:  fmt.Sprintf(`e.g. simple-pass %s --password <valid-password> --filePath <valid-path>`, LoadPassDBCmdName),
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Debugf("%s called with - password:%s,filePath:%s\n", LoadPassDBCmdName, password, filePath)
			_, err := db.LoadExistingPassDB(filePath, password)
			if err != nil {
				return fmt.Errorf("failed to load passDB - %s", err)
			}
			log.Infof(SuccessfullyLoadedPassDBMessage, filePath)
			err = SetPassDBCache(filePath, password)
			if err != nil {
				return fmt.Errorf("failed to update the passDBCache with the details for this new passDB: %s", err)
			}
			log.Debugf(SuccessfullySetPassDBCacheMessage)
			return nil
		},
	}
	cmd.Flags().StringVarP(&password, PassDBPasswordFlag, PassDBPasswordShortFlag, "", "password for the passdb (Required)")
	cmd.Flags().StringVarP(&filePath, PassDBFilePathFlag, PassDBFilePathShortFlag, "", "path to the new passdb file (Required)")
	err := cmd.MarkFlagRequired("password")
	if err != nil {
		panic(fmt.Sprintf("cannot setup cobra command:%s", err))
	}
	err = cmd.MarkFlagRequired("filePath")
	if err != nil {
		panic(fmt.Sprintf("cannot setup cobra command:%s", err))
	}
	return cmd
}
