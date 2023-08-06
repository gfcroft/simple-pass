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
	RenameCmdName              = "rename"
	ToFlag                     = "to"
	ToShortFlag                = "t"
	SuccessfullyRenamedMessage = "successfully renamed item from:'%s' to:'%s'"
)

func NewRenameCmd(passDB *db.PassDB) *cobra.Command {
	var (
		to string
	)

	cmd := &cobra.Command{
		Use:   RenameCmdName,
		Short: "renames an item",
		Long: fmt.Sprintf(`e.g.
			simple-pass %s <item-name> --to <non-blank-name>`, RenameCmdName),
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Debugf("%s called with %v", RenameCmdName, args)
			if len(args) != 1 {
				err := cmd.Help()
				if err != nil {
					return fmt.Errorf("attempting to show help prompt caused error: %s\n", err)
				}
				return nil
			}
			oldItemName := args[0]
			err := passDB.RenameItem(oldItemName, to)
			if err != nil {
				return fmt.Errorf("cannot rename item %s to %s: %s\n", oldItemName, to, err)
			}
			log.Infof(SuccessfullyRenamedMessage, oldItemName, to)

			return nil
		},
	}
	cmd.Flags().StringVarP(&to, ToFlag, ToShortFlag, "", "name to change item to")
	err := cmd.MarkFlagRequired(ToFlag)
	if err != nil {
		panic(fmt.Sprintf("cannot setup cobra command:%s", err))
	}
	return cmd
}
