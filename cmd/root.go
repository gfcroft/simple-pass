/*
Copyright © 2023 George Wheatcroft

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	_ "embed"
	"io"
	"os"

	"github.com/georgewheatcroft/simple-pass/internal/db"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

//go:embed VERSION
var cliVersion string

// rootCmd represents the base command when called without any subcommands
func NewRootCmd(setOut, setErr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "simple-pass",
		Short: "Simple CLI password manager",
		Long: `simple-pass - simple cli password management

For example:

		simple-pass create-pass-db foobar --password "something at least 5 chars"
		simple-pass add eg --username "me" --password "whatever"

		use flag: '-h' for more information`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		// Run: func(cmd *cobra.Command, args []string) { },
	}

	cmd.Version = cliVersion
	cmd.SetOutput(setOut)
	cmd.SetErr(setErr)

	return cmd
}

// Execute adds all child commands to the root command and sets flags and cmd exec logging
// appropriately. This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(setOut, setErr io.Writer) {
	rootCmd := NewRootCmd(setOut, setErr)

	log.SetOutput(rootCmd.OutOrStdout())
	var passDB *db.PassDB
	if passDBCacheExists() {
		passDB = loadPassDB(getPassDBPath(), getPassDBPassword())

	}

	//add all of the commands currently in use before exec (TODO tidy up with command groups?)
	// - opportunity to dynamically load commands based on os.Args if required
	rootCmd.AddCommand(
		NewCreatePassDbCmd(),
		NewLoadPassDbCmd(),
		NewAddCmd(passDB),
		NewGetCmd(passDB),
		NewStatusCmd(passDB),
		NewListCmd(passDB),
		NewUpdateCmd(passDB),
		NewRenameCmd(passDB),
		NewDeleteCmd(passDB),
	)

	err := rootCmd.Execute()
	if err != nil {
		// err printing logging appears to be handled by cobra fw - just exit

		os.Exit(1)
	}
}
