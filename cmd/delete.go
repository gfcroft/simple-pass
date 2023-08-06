package cmd

import (
	"fmt"
	"os"

	"github.com/georgewheatcroft/simple-pass/internal/db"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	DeleteCmdName = "delete"
)

func NewDeleteCmd(passDB *db.PassDB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   DeleteCmdName,
		Short: "remove items from your simple-pass",
		Long: fmt.Sprintf(`e.g.
			   simple-pass %s <existing-item-name>`, DeleteCmdName),
		PreRunE: passDBCacheExistsOrErr,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				err := cmd.Help()
				if err != nil {
					log.Fatalln(err)
				}
				os.Exit(1)
			}
			log.Debugf("%s called with %v", DeleteCmdName, args)

			itemName := args[0]
			err := passDB.DeleteItem(itemName)
			if err != nil {
				return fmt.Errorf("cannot delete item from passDB: %s\n", err)
			}
			log.Infof("successfully deleted item '%s' from passDB\n", itemName)
			return nil
		},
	}
	return cmd
}
