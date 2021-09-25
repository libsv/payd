package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"

	"github.com/libsv/payd/cli/config"
	"github.com/libsv/payd/cli/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	verbose      bool
	outputFormat string
	useContext   string
)

var (
	printer output.PrintFunc
	cfg     *config.Config
)

var (
	// ErrContextNotFound when context doesn't exist.
	ErrContextNotFound      = errors.New("context not found")
	ErrContextAlreadyExists = errors.New("context already exists")
)

var rootCmd = &cobra.Command{
	Use:           "payctl",
	SilenceErrors: true,
	SilenceUsage:  true,
	Short:         "Interface with payd",
	Long:          "Interface with payd",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		printer = output.NewPrinter(output.Format(outputFormat))
		if useContext != "" {
			if ok := cfg.LoadContext(useContext); !ok {
				return ErrContextNotFound
			}
		}

		return nil
	},
}

func init() {
	initConfig()

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose printing")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "output format")
	rootCmd.PersistentFlags().StringVar(&useContext, "use-context", "", "output format")
}

func initConfig() {
	home, err := homedir.Dir()
	cobra.CheckErr(err)

	viper.AddConfigPath(home)
	viper.SetConfigName(".payctl")
	viper.SetConfigType("yml")

	_ = viper.ReadInConfig()

	cfg = config.NewConfig().
		WithPayd().
		WithP4().
		WithAccount().
		WithContexts()

	_ = viper.SafeWriteConfig()
}

// Execute the command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
