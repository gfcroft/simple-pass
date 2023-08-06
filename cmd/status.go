package cmd

import (
	"fmt"
	"text/template"

	"github.com/georgewheatcroft/simple-pass/internal/db"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type Status struct {
	PassDBSet  bool
	PassDBName string
	TotalItems int
}

const (
	StatusCmdName      = "status"
	PassDBNotLoadedMsg = "there is no passDB loaded - please set one up"
	//TODO embed this eventually using go:embed ??
	statusTmpl = `passDB Set: {{ .PassDBSet }}
{{- if .PassDBSet }}
passDB Name: {{ .PassDBName }}
TotalItems: {{ .TotalItems }}
{{- end }}
`
)

func NewStatusCmd(passDB *db.PassDB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   StatusCmdName,
		Short: "return information relating to your simple-pass setup",
		Long: fmt.Sprintf(`e.g.
			simple-pass %s`, StatusCmdName),
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Debugf(`%s called`, StatusCmdName)

			var status Status
			if passDB != nil {
				status = Status{PassDBSet: true, PassDBName: passDB.GetPassDBName(), TotalItems: len(passDB.ListAllItems())}
			} else {
				log.Println(PassDBNotLoadedMsg)
				return nil
			}

			tmpl, err := template.New("status").Parse(statusTmpl)
			if err != nil {
				return fmt.Errorf("cant parse template: %s\n", err)
			}
			err = tmpl.Execute(cmd.OutOrStdout(), status)
			if err != nil {
				return fmt.Errorf("failed to execute template: %s\n", err)
			}
			return nil
		},
	}
	return cmd
}
