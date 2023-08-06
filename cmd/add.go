package cmd

import (
	"fmt"

	"github.com/georgewheatcroft/simple-pass/internal/db"
	"github.com/georgewheatcroft/simple-pass/internal/item"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	AddCmdName = "add"

	UsernameFlag      = "username"
	UsernameShortFlag = "u"
	PasswordFlag      = "password"
	PasswordShortFlag = "p"
	NotesFlag         = "notes"
	NotesShortFlag    = "n"
	URLFlag           = "url"
	URLShortFlag      = "w"

	SuccessfullyAddedMessage = "successfully added %s to the passDB\n"
)

func NewAddCmd(passDB *db.PassDB) *cobra.Command {
	var (
		itemUsername string
		itemPassword string
		itemNotes    []string
		itemURL      string
	)

	cmd := &cobra.Command{
		Use:   AddCmdName,
		Short: "add new item to your simple-pass",
		Long: `e.g. 
			   simple-pass add <new-item-name> --password <some password>`,
		PreRunE: passDBCacheExistsOrErr,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Debugf("add called with %v", args)
			// TODO surely this cobra fw has something for this
			if len(args) != 1 {
				err := cmd.Help()
				if err != nil {
					log.Fatalln(err)
				}
				return err
			}

			itemName := args[0]

			newItem, err := item.NewItem(itemName, itemUsername, itemPassword, itemURL, itemNotes)
			if err != nil {
				return fmt.Errorf("cannot add new item to passDB: %s\n", err)
			}

			err = passDB.SaveNewItem(newItem)
			if err != nil {
				return fmt.Errorf("cannot add new item to passDB: %s\n", err)
			}
			log.Infof(SuccessfullyAddedMessage, itemName)
			return nil
		},
	}
	cmd.Flags().StringVarP(&itemUsername, UsernameFlag, UsernameShortFlag, "", "username for the item")
	cmd.Flags().StringVarP(&itemPassword, PasswordFlag, PasswordShortFlag, "", "password for the item")
	cmd.Flags().StringArrayVarP(&itemNotes, NotesFlag, NotesShortFlag, nil, "notes for the item")
	cmd.Flags().StringVarP(&itemURL, URLFlag, URLShortFlag, "", "url for the item")

	return cmd
}
