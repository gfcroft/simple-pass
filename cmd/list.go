package cmd

import (
	"fmt"
	"strings"

	"github.com/georgewheatcroft/simple-pass/internal/db"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	ListCmdName = "list"
)

func NewListCmd(passDB *db.PassDB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   ListCmdName,
		Short: "display the names of all items in your simple-pass",
		Long: fmt.Sprintf(`e.g.
			simple-pass %s`, ListCmdName),
		PreRunE: passDBCacheExistsOrErr,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Debugf("%s called\n", ListCmdName)
			items := passDB.ListAllItems()
			log.Debugf("retrieved the following item names: %v", items)

			// TODO nicer/pretty output would be good
			fmt.Println(strings.Join(items, "\n"))
			return nil
		},
	}
	return cmd
}
