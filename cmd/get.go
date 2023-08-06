package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/georgewheatcroft/simple-pass/internal/db"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	GetCmdName = "get"
)

var ErrItemDoesNotExist = db.ErrItemDoesNotExist

func NewGetCmd(passDB *db.PassDB) *cobra.Command {
	var (
		urlFlag      bool
		notesFlag    bool
		passwordFlag bool
		usernameFlag bool
	)

	cmd := &cobra.Command{
		Use:   GetCmdName,
		Short: "retrieve an item, or a part of an item, from your simple-pass",
		Long: fmt.Sprintf(`e.g.

			   simple-pass %s <existing-item-name>
			   simple-pass %s <existing-item-name> --password`, GetCmdName, GetCmdName),
		PreRunE: passDBCacheExistsOrErr,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				err := cmd.Help()
				if err != nil {
					log.Fatalln(err)
				}
				os.Exit(1)
			}
			log.Debugf("%s called with %v", GetCmdName, args)

			flags := cmd.Flags()
			noFlags := flags.NFlag()
			// cannot have more than 1 flag currently
			if noFlags > 1 {
				err := cmd.Help()
				if err != nil {
					log.Fatalln(err)
				}
				os.Exit(1)
			}

			itemName := args[0]
			itemRetrieved, err := passDB.RetrieveItem(itemName)
			if err != nil {
				if errors.Is(err, db.ErrItemDoesNotExist) {

				}
				return fmt.Errorf("cannot retrieve item details from passDB: %s\n", err)
			}

			if !cmd.HasFlags() {
				// TODO better print of this
				log.Infof("%v\n", itemRetrieved)
				return nil
			} else {
				err := cmd.ParseFlags(args)
				if err != nil {
					return fmt.Errorf("cannot access flags provided to command: %s\n", err)
				}

				if noFlags == 0 {
					json, err := json.Marshal(itemRetrieved)
					if err != nil {
						return err
					}
					fmt.Println(string(json))
					return nil
				}

				// TODO currently can't find a way to get out a list/something of all flags present in
				// command issued (e.g. you supplied -a, so [-a]) so resorting this this horrible show for now
				getFlags := map[string]string{
					"url":      itemRetrieved.URL,
					"notes":    strings.Join(itemRetrieved.Notes, "\n"),
					"password": itemRetrieved.Password,
					"username": itemRetrieved.Username,
				}

				for flag, value := range getFlags {
					exists, err := strconv.ParseBool(flags.Lookup(flag).Value.String())
					if err != nil {
						return fmt.Errorf("can't parse boolean flag - %s", err)
					}
					if exists {
						fmt.Printf("%s", value)
						return nil
					}
				}

				return fmt.Errorf("cannot determine what to retrieve for item based on inputs\n")
			}

		},
	}

	//setting more than one flag should cause an err
	//if none set, retrieve and spit out everything for the item
	cmd.Flags().BoolVarP(&usernameFlag, "username", "u", false, "username")
	cmd.Flags().BoolVarP(&passwordFlag, "password", "p", false, "password")
	cmd.Flags().BoolVarP(&notesFlag, "notes", "n", false, "notes")
	cmd.Flags().BoolVarP(&urlFlag, "url", "w", false, "url")
	cmd.MarkFlagsMutuallyExclusive("username", "password", "notes", "url")
	return cmd
}
