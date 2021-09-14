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
)

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

var printer output.PrintFunc

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "version", "v", false, "verbose printing")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "output format")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
