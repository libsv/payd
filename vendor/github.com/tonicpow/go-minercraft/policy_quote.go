package minercraft

import (
	"context"
	"encoding/json"
	"errors"
)

/*
Example policyQuote response from Merchant API:

{
    "payload": "{\"apiVersion\":\"1.4.0\",\"timestamp\":\"2021-11-12T13:17:47.7498672Z\",\"expiryTime\":\"2021-11-12T13:27:47.7498672Z\",\"minerId\":\"030d1fe5c1b560efe196ba40540ce9017c20daa9504c4c4cec6184fc702d9f274e\",\"currentHighestBlockHash\":\"45628be2fe616167b7da399ab63455e60ffcf84147730f4af4affca90c7d437e\",\"currentHighestBlockHeight\":234,\"fees\":[{\"feeType\":\"standard\",\"miningFee\":{\"satoshis\":500,\"bytes\":1000},\"relayFee\":{\"satoshis\":250,\"bytes\":1000}},{\"feeType\":\"data\",\"miningFee\":{\"satoshis\":500,\"bytes\":1000},\"relayFee\":{\"satoshis\":250,\"bytes\":1000}}],\"callbacks\":[{\"ipAddress\":\"123.456.789.123\"}],\"policies\":{\"skipscriptflags\":[\"MINIMALDATA\",\"DERSIG\",\"NULLDUMMY\",\"DISCOURAGE_UPGRADABLE_NOPS\",\"CLEANSTACK\"],\"maxtxsizepolicy\":99999,\"datacarriersize\":100000,\"maxscriptsizepolicy\":100000,\"maxscriptnumlengthpolicy\":100000,\"maxstackmemoryusagepolicy\":10000000,\"limitancestorcount\":1000,\"limitcpfpgroupmemberscount\":10,\"acceptnonstdoutputs\":true,\"datacarrier\":true,\"dustrelayfee\":150,\"maxstdtxvalidationduration\":99,\"maxnonstdtxvalidationduration\":100,\"dustlimitfactor\":10}}",
    "signature": "30440220708e2e62a393f53c43d172bc1459b4daccf9cf23ff77cff923f09b2b49b94e0a022033792bee7bc3952f4b1bfbe9df6407086b5dbfc161df34fdee684dc97be72731",
    "publicKey": "030d1fe5c1b560efe196ba40540ce9017c20daa9504c4c4cec6184fc702d9f274e",
    "encoding": "UTF-8",
    "mimetype": "application/json"
}
*/

// PolicyQuoteResponse is the raw response from the Merchant API request
//
// Specs: https://github.com/bitcoin-sv-specs/brfc-merchantapi#1-get-policy-quote
type PolicyQuoteResponse struct {
	JSONEnvelope
	Quote *PolicyPayload `json:"quote"` // Custom field for unmarshalled payload data
}

/*
Example PolicyQuoteResponse.Payload (unmarshalled):

{
    "apiVersion": "1.4.0",
    "timestamp": "2021-11-12T13:17:47.7498672Z",
    "expiryTime": "2021-11-12T13:27:47.7498672Z",
    "minerId": "030d1fe5c1b560efe196ba40540ce9017c20daa9504c4c4cec6184fc702d9f274e",
    "currentHighestBlockHash": "45628be2fe616167b7da399ab63455e60ffcf84147730f4af4affca90c7d437e",
    "currentHighestBlockHeight": 234,
    "fees": [
        {
            "feeType": "standard",
            "miningFee": {
                "satoshis": 500,
                "bytes": 1000
            },
            "relayFee": {
                "satoshis": 250,
                "bytes": 1000
            }
        },
        {
            "feeType": "data",
            "miningFee": {
                "satoshis": 500,
                "bytes": 1000
            },
            "relayFee": {
                "satoshis": 250,
                "bytes": 1000
            }
        }
    ],
    "callbacks": [
        {
            "ipAddress": "123.456.789.123"
        }
    ],
    "policies": {
        "skipscriptflags": [ "MINIMALDATA", "DERSIG", "NULLDUMMY", "DISCOURAGE_UPGRADABLE_NOPS", "CLEANSTACK" ],
        "maxtxsizepolicy": 99999,
        "datacarriersize": 100000,
        "maxscriptsizepolicy": 100000,
        "maxscriptnumlengthpolicy": 100000,
        "maxstackmemoryusagepolicy": 10000000,
        "limitancestorcount": 1000,
        "limitcpfpgroupmemberscount": 10,
        "acceptnonstdoutputs": true,
        "datacarrier": true,
        "dustrelayfee": 150,
        "maxstdtxvalidationduration": 99,
        "maxnonstdtxvalidationduration": 100,
        "dustlimitfactor": 10
    }
}
*/

// PolicyPayload is the unmarshalled version of the payload envelope
type PolicyPayload struct {
	FeePayload                   // Inherit the same structure as the fee payload
	Callbacks  []*PolicyCallback `json:"callbacks"` // IP addresses of double-spend notification servers such as mAPI reference implementation
	Policies   *Policy           `json:"policies"`  // values of miner policies as configured by the mAPI reference implementation administrator
}

// ScriptFlag is a flag used in the policy quote
type ScriptFlag string

// All known script flags
const (
	FlagCleanStack               ScriptFlag = "CLEANSTACK"
	FlagDerSig                   ScriptFlag = "DERSIG"
	FlagDiscourageUpgradableNops ScriptFlag = "DISCOURAGE_UPGRADABLE_NOPS"
	FlagMinimalData              ScriptFlag = "MINIMALDATA"
	FlagNullDummy                ScriptFlag = "NULLDUMMY"
)

// Policy is the struct of a policy (from policy quote response)
type Policy struct {
	AcceptNonStdOutputs           bool         `json:"acceptnonstdoutputs"`
	DataCarrier                   bool         `json:"datacarrier"`
	DataCarrierSize               uint32       `json:"datacarriersize"`
	DustLimitFactor               uint32       `json:"dustlimitfactor"`
	DustRelayFee                  uint32       `json:"dustrelayfee"`
	LimitAncestorCount            uint32       `json:"limitancestorcount"`
	LimitCpfpGroupMembersCount    uint32       `json:"limitcpfpgroupmemberscount"`
	MaxNonStdTxValidationDuration uint32       `json:"maxnonstdtxvalidationduration"`
	MaxScriptNumLengthPolicy      uint32       `json:"maxscriptnumlengthpolicy"`
	MaxScriptSizePolicy           uint32       `json:"maxscriptsizepolicy"`
	MaxStackMemoryUsagePolicy     uint64       `json:"maxstackmemoryusagepolicy"`
	MaxStdTxValidationDuration    uint32       `json:"maxstdtxvalidationduration"`
	MaxTxSizePolicy               uint32       `json:"maxtxsizepolicy"`
	SkipScriptFlags               []ScriptFlag `json:"skipscriptflags"`
}

// PolicyCallback is the callback address
type PolicyCallback struct {
	IPAddress string `json:"ipAddress"`
}

// PolicyQuote will fire a Merchant API request to retrieve the policy from a given miner
//
// This endpoint is used to get the different policies quoted by a miner.
// It returns a JSONEnvelope with a payload that contains the policies used by a specific BSV miner.
// The purpose of the envelope is to ensure strict consistency in the message content for
// the purpose of signing responses. This is a superset of the fee quote service, as it also
// includes information on DSNT IP addresses and miner policies.
//
// Specs: https://github.com/bitcoin-sv-specs/brfc-merchantapi#1-get-policy-quote
func (c *Client) PolicyQuote(ctx context.Context, miner *Miner) (*PolicyQuoteResponse, error) {

	// Make sure we have a valid miner
	if miner == nil {
		return nil, errors.New("miner was nil")
	}

	// Make the HTTP request
	result := getQuote(ctx, c, miner, routePolicyQuote)
	if result.Response.Error != nil {
		return nil, result.Response.Error
	}

	// Parse the response
	response, err := result.parsePolicyQuote()
	if err != nil {
		return nil, err
	}

	// Valid?
	if response.Quote == nil || len(response.Quote.Fees) == 0 {
		return nil, errors.New("failed getting policy from: " + miner.Name)
	}

	// Return the fully parsed response
	return &response, nil
}

// parsePolicyQuote will convert the HTTP response into a struct and also unmarshal the payload JSON data
func (i *internalResult) parsePolicyQuote() (response PolicyQuoteResponse, err error) {

	// Process the initial response payload
	if err = response.process(i.Miner, i.Response.BodyContents); err != nil {
		return
	}

	// If we have a valid payload
	if len(response.Payload) > 0 {
		if err = json.Unmarshal([]byte(response.Payload), &response.Quote); err != nil {
			return
		}
		if response.Quote != nil &&
			len(response.Quote.Fees) > 0 &&
			len(response.Quote.Fees[0].FeeType) == 0 { // This is an issue because go-bt json field is stripping the types
			response.Quote.Fees[0].FeeType = FeeTypeStandard
			response.Quote.Fees[1].FeeType = FeeTypeData
		}
	}
	return
}
