package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-bt/v2"
	chttp "github.com/libsv/payd/cli/data/http"
	"github.com/libsv/payd/cli/data/regtest"
	"github.com/libsv/payd/cli/models"
	"github.com/libsv/payd/cli/service"
	"github.com/spf13/cobra"
)

var (
	requestJSON  string
	payToURL     string
	payToContext string
	lowFees      bool
)

var paymentCmd = &cobra.Command{
	Use:     "payment",
	Aliases: []string{"p"},
	Short:   "payments",
	Long:    "payments",
}

var paymentRequestCmd = &cobra.Command{
	Use:     "request",
	Aliases: []string{"req"},
	Short:   "request a payment",
	Long:    "request a payment",
	PreRunE: paymentValidator,
	Args:    cobra.MinimumNArgs(1),
	RunE:    paymentRequest,
}

var paymentBuildCmd = &cobra.Command{
	Use:     "build",
	Aliases: []string{"buil", "bui", "bu", "b"},
	Short:   "build a payment message",
	Long:    "build a payment message",
	RunE:    paymentBuild,
}

var sendCmd = &cobra.Command{
	Use:     "send",
	Aliases: []string{"send", "s"},
	Short:   "send satoshis to address",
	Long:    "send satoshis to address",
	PreRunE: paymentValidator,
	RunE:    paymentSend,
}

func init() {
	rootCmd.AddCommand(paymentCmd)

	paymentRequestCmd.PersistentFlags().StringVarP(&payToURL, "pay-to-url", "u", "", "the payto url")
	paymentRequestCmd.PersistentFlags().StringVarP(&payToContext, "pay-to-context", "c", "", "the payto context")
	paymentCmd.AddCommand(paymentRequestCmd)

	sendCmd.PersistentFlags().StringVarP(&payToURL, "pay-to-url", "u", "", "the payto url")
	sendCmd.PersistentFlags().StringVarP(&payToContext, "pay-to-context", "c", "", "the payto context")
	paymentCmd.AddCommand(sendCmd)

	paymentBuildCmd.Flags().BoolVarP(&lowFees, "low-fees", "l", false, "build tx with too low fees")
	paymentCmd.AddCommand(paymentBuildCmd)
}

func paymentValidator(cmd *cobra.Command, args []string) error {
	if payToURL == "" && payToContext == "" {
		return errors.New("either --pay-to-url or --pay-to-context must be provided")
	}

	if payToContext != "" {
		if ok := cfg.HasContext(payToContext); !ok {
			return ErrContextNotFound
		}

		payToURL = cfg.Contexts[payToContext].P4.URLFor()
	}

	_, err := url.Parse(payToURL)
	return err
}

func paymentRequest(cmd *cobra.Command, args []string) error {
	svc := service.NewPaymentService(chttp.NewPaymentAPI(&http.Client{}), nil)

	req, err := svc.Request(cmd.Context(), models.PaymentRequestArgs{
		ID:    args[0],
		PayTo: payToURL,
	})
	if err != nil {
		return err
	}

	return printer(req)
}

func paymentBuild(cmd *cobra.Command, args []string) error {
	var rdr io.Reader = os.Stdin
	if requestJSON != "" {
		rdr = bytes.NewBufferString(requestJSON)
	}

	var payReq models.PaymentRequest
	if err := json.NewDecoder(rdr).Decode(&payReq); err != nil {
		return err
	}

	if lowFees {
		fmt.Fprintln(os.Stderr, "creating tx with low fees")
		payReq.Fee.AddQuote(bt.FeeTypeStandard, &bt.Fee{
			MiningFee: bt.FeeUnit{
				Satoshis: 100,
				Bytes:    500,
			},
			RelayFee: bt.FeeUnit{
				Satoshis: 100,
				Bytes:    500,
			},
		})
		payReq.Fee.AddQuote(bt.FeeTypeData, &bt.Fee{
			MiningFee: bt.FeeUnit{
				Satoshis: 100,
				Bytes:    500,
			},
			RelayFee: bt.FeeUnit{
				Satoshis: 100,
				Bytes:    500,
			},
		})
	}

	rt := regtest.NewRegtest(&http.Client{})
	spvb, err := spv.NewEnvelopeCreator(service.NewTxService(rt), service.NewMerkleProofStore(rt))
	if err != nil {
		return err
	}
	svc := service.NewFundService(rt, chttp.NewPaymentAPI(&http.Client{}), spvb)

	tx, err := svc.FundedTx(cmd.Context(), payReq)
	if err != nil {
		return err
	}

	spvEnvelope, err := spvb.CreateEnvelope(cmd.Context(), tx)
	if err != nil {
		return err
	}

	return printer(models.PaymentSendArgs{
		Transaction:    tx.String(),
		PaymentRequest: payReq,
		MerchantData:   payReq.MerchantData,
		Memo:           payReq.Memo,
		SPVEnvelope:    spvEnvelope,
	})
}

func paymentSend(cmd *cobra.Command, args []string) error {
	var rdr io.Reader = os.Stdin
	if requestJSON != "" {
		rdr = bytes.NewBufferString(requestJSON)
	}

	var paySend models.PaymentSendArgs
	if err := json.NewDecoder(rdr).Decode(&paySend); err != nil {
		return err
	}

	paySend.PaymentRequest.PaymentURL = fmt.Sprintf("%s/api/v1/payment/%s", payToURL, paySend.MerchantData.ExtendedData["paymentReference"])

	svc := service.NewPaymentService(chttp.NewPaymentAPI(&http.Client{}), nil)

	ack, err := svc.Send(cmd.Context(), paySend)
	if err != nil {
		return err
	}

	return printer(ack)
}
