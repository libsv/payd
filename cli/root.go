package cli

import (
	"fmt"
	"os"

	"github.com/libsv/payd/cli/output"
	"github.com/spf13/cobra"
)

var (
	verbose      bool
	outputFormat string
	account      string
)

var printer output.PrintFunc

var rootCmd = &cobra.Command{
	Use:           "payctl",
	SilenceErrors: true,
	SilenceUsage:  true,
	Short:         "Interface with payd",
	Long:          "Interface with payd",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		printer = output.NewPrinter(output.Format(outputFormat))
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "version", "v", false, "verbose printing")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "output format")
	rootCmd.PersistentFlags().StringVarP(&account, "account", "a", "client", "account name")
}

// Execute the command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
