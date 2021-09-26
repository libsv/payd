package cli

import (
	"net/http"

	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/payd/cli/data/regtest"
	"github.com/libsv/payd/cli/models"
	"github.com/libsv/payd/cli/service"
	"github.com/spf13/cobra"
)

var spvEnvelopeCmd = &cobra.Command{
	Use:     "spvenvelope",
	Aliases: []string{"spvenvelopes", "spvenvel", "spvenv", "spve"},
	Short:   "spv envelope operations",
	Long:    "spv envelope operations",
}

var envelopeCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"cr", "c"},
	Short:   "create an spv envelope",
	Long:    "create an spv envelope",
	Args:    cobra.MinimumNArgs(1),
	RunE:    envelopeCreate,
}

func init() {
	rootCmd.AddCommand(spvEnvelopeCmd)
	spvEnvelopeCmd.AddCommand(envelopeCreateCmd)
}

func envelopeCreate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	rt := regtest.NewRegtest(&http.Client{})
	txSvc := service.NewTxService(rt)
	mpSvc := service.NewMerkleProofStore(rt)
	spvEnvelopeBuilder, err := spv.NewEnvelopeCreator(txSvc, mpSvc)
	if err != nil {
		return err
	}

	fns := map[bool]func(string) (string, error){
		true: func(s string) (string, error) {
			result, err := rt.RawTransaction(ctx, s)
			if err != nil {
				return "", err
			}
			return *result.Result, nil

		},
		false: func(s string) (string, error) { return s, nil },
	}

	rawTx, err := fns[len(args[0]) == 64](args[0])
	if err != nil {
		return err
	}

	tx, err := bt.NewTxFromString(rawTx)
	if err != nil {
		return err
	}

	envelope, err := spvEnvelopeBuilder.CreateEnvelope(ctx, tx)
	if err != nil {
		return err
	}

	return printer(models.SPVEnvelope{
		Envelope: envelope,
	})
}
