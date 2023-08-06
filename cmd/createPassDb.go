package cmd

import (
	"fmt"

	"github.com/georgewheatcroft/simple-pass/internal/db"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

/*
	TODO consolidate the semantically similar createPassDB/Add [item] commands properly, so that add has those as
	subresources below it - see /simple-pass/TODOs
*/

const (
	CreatePassDBCmdName     = "create-pass-db"
	PassDBNameFlag          = "name"
	PassDBNameShortFlag     = "n"
	PassDBPasswordFlag      = "password"
	PassDBPasswordShortFlag = "p"
	PassDBFilePathFlag      = "filePath"
	PassDBFilePathShortFlag = "f"

	SuccessfullyCreatedPassDBMessage = "created new passdb: '%s' at %s"
	/* #nosec */
	SuccessfullySetPassDBCacheMessage = "passdb cache updated with new passDB details"
)

func NewCreatePassDbCmd() *cobra.Command {
	var (
		name     string
		password string
		filePath string
	)

	cmd := &cobra.Command{
		Use:   CreatePassDBCmdName,
		Short: "creates a new simple-pass db",
		Long: fmt.Sprintf(`e.g.
			simple-pass %s --name <non-blank-name>  --password <valid-password> --filePath <valid-path>`, CreatePassDBCmdName),
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Debugf("%s called with - name:%s,password:%s,filePath:%s\n", CreatePassDBCmdName, name, password, filePath)
			_, err := db.CreatePassDB(filePath, name, password)
			if err != nil {
				//TODO some failure cases may leave an empty passdb on disk - need to avoid this
				return fmt.Errorf("failed to create passDB - %s", err)
			}
			log.Infof(SuccessfullyCreatedPassDBMessage, name, filePath)
			err = SetPassDBCache(filePath, password)
			if err != nil {
				return fmt.Errorf("failed to update the passDBCache with the details for this new passDB: %s", err)
			}
			log.Debugf(SuccessfullySetPassDBCacheMessage)
			return nil
		},
	}
	cmd.Flags().StringVarP(&name, PassDBNameFlag, PassDBNameShortFlag, "", "name for the passdb (Required)")
	cmd.Flags().StringVarP(&password, PassDBPasswordFlag, PassDBPasswordShortFlag, "", "password for the passdb (Required)")
	cmd.Flags().StringVarP(&filePath, PassDBFilePathFlag, PassDBFilePathShortFlag, "", "path to the new passdb file (Required)")

	err := cmd.MarkFlagRequired(PassDBNameFlag)
	if err != nil {
		panic(fmt.Sprintf("cannot setup cobra command:%s", err))
	}
	err = cmd.MarkFlagRequired(PasswordFlag)
	if err != nil {
		panic(fmt.Sprintf("cannot setup cobra command:%s", err))
	}
	err = cmd.MarkFlagRequired(PassDBFilePathFlag)
	if err != nil {
		panic(fmt.Sprintf("cannot setup cobra command:%s", err))
	}
	return cmd
}
