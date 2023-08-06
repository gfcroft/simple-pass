package cmd

import (
	"fmt"
	"strings"

	"github.com/georgewheatcroft/simple-pass/internal/db"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	UpdateCmdName = "update"
)

func NewUpdateCmd(passDB *db.PassDB) *cobra.Command {
	var (
		itemUsername string
		itemPassword string
		itemNotes    []string
		itemURL      string
	)

	cmd := &cobra.Command{
		Use:   UpdateCmdName,
		Short: "update an item part in your simple-pass",
		Long: fmt.Sprintf(`e.g.
			simple-pass %s <existing-item-name> --url <new value> 
			simple-pass %s <existing-item-name> --notes <new value> --password <new value>`, UpdateCmdName, UpdateCmdName),
		PreRunE: passDBCacheExistsOrErr,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Debugf("%s called with %v", UpdateCmdName, args)
			if len(args) != 1 {
				err := cmd.Help()
				if err != nil {
					return fmt.Errorf("attempting to show help prompt caused error: %s\n", err)
				}
				return nil
			}

			itemName := args[0]
			retrievedItem, err := passDB.RetrieveItem(itemName)
			if err != nil {
				return fmt.Errorf("cannot update item: %s\n", err)
			}
			newItem := *retrievedItem

			err = cmd.ParseFlags(args)
			if err != nil {
				return fmt.Errorf("cannot update item: %s\n", err)
			}
			flags := cmd.Flags()
			setUsername := flags.Lookup("username")
			setPassword := flags.Lookup("password")
			setNotes := flags.Lookup("notes")
			setURL := flags.Lookup("url")

			if setUsername.Changed {
				newItem.Username = setUsername.Value.String()
			}
			if setPassword.Changed {
				newItem.Password = setPassword.Value.String()
			}
			if setNotes.Changed {
				newItem.Notes = strings.Split(setNotes.Value.String(), "\n")
			}
			if setURL.Changed {
				newItem.URL = setURL.Value.String()
			}

			err = passDB.UpdateItem(&newItem)
			if err != nil {
				return fmt.Errorf("can't update item - %s", err)
			}
			fmt.Printf("updated item:'%s'", itemName)
			return nil
		},
	}
	cmd.Flags().StringVarP(&itemUsername, "username", "u", "", "username for the item")
	cmd.Flags().StringVarP(&itemPassword, "password", "p", "", "password for the item")
	cmd.Flags().StringArrayVarP(&itemNotes, "notes", "n", nil, "notes for the item")
	cmd.Flags().StringVarP(&itemURL, "url", "w", "", "url for the item")
	return cmd
}
