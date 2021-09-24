package cli

import "github.com/spf13/cobra"

var sendCmd = &cobra.Command{
	Use:     "send",
	Aliases: []string{"send", "s"},
	Short:   "send satoshis to address",
	Long:    "send satoshis to address",
	RunE:    send,
}

func init() {
	rootCmd.AddCommand(sendCmd)
	// sendCmd.Flags().StringVarP()
}

func send(cmd *cobra.Command, args []string) error {
	return nil
}
