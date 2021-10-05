package cli

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/labstack/gommon/log"
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

// Error codes.
var (
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

	_, err = os.Stat(path.Join(home, ".payctl.yml"))
	createConfig := err != nil && os.IsNotExist(err)

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	cfg = config.NewConfig().
		WithPayd().
		WithP4().
		WithAccount().
		WithContexts()

	if createConfig {
		_ = viper.WriteConfig()
	}
}

// Execute the command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
