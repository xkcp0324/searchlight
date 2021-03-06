package check_env

import (
	"fmt"
	"os"

	"github.com/appscode/searchlight/pkg/icinga"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "check_env",
		Run: func(c *cobra.Command, args []string) {
			envList := os.Environ()
			fmt.Fprintln(os.Stdout, "Total ENV: ", len(envList))
			fmt.Fprintln(os.Stdout)
			for _, env := range envList {
				fmt.Fprintln(os.Stdout, env)
			}
			icinga.Output(icinga.OK, "A-OK")
		},
	}
	return cmd
}
