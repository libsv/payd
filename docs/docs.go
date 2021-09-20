// Package docs GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "https://github.com/libsv/payd/blob/master/CODE_OF_CONDUCT.md",
        "contact": {},
        "license": {
            "name": "ISC",
            "url": "https://github.com/libsv/payd/blob/master/LICENSE"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/balance": {
            "get": {
                "description": "Returns current balance, which is a sum of unspent txos",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Balance"
                ],
                "summary": "Balance",
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            }
        },
        "/invoices": {
            "get": {
                "description": "Returns all invoices currently stored",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Invoices"
                ],
                "summary": "Invoices",
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            },
            "post": {
                "description": "Creates an invoice with payment id and satoshis",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Invoices"
                ],
                "summary": "Create invoice",
                "parameters": [
                    {
                        "description": "PaymentReference and Satoshis",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/gopayd.InvoiceCreate"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": ""
                    }
                }
            },
            "delete": {
                "description": "Delete",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Invoices"
                ],
                "summary": "Delete invoice",
                "parameters": [
                    {
                        "type": "string",
                        "description": "PaymentReference",
                        "name": "PaymentReference",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            }
        },
        "/invoices/{paymentID}": {
            "get": {
                "description": "Returns invoice by payment id if exists",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Invoices"
                ],
                "summary": "Invoices",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Payment ID",
                        "name": "paymentID",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            }
        },
        "/payment/{paymentID}": {
            "get": {
                "description": "Creates a payment request based on a payment id (the identifier for an invoice).",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Payment"
                ],
                "summary": "Create payment request",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Payment ID",
                        "name": "paymentID",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            },
            "post": {
                "description": "Creates a payment based on a payment id (the identifier for an invoice).",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Payment"
                ],
                "summary": "Create payment",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Payment ID",
                        "name": "paymentID",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "payment message used in BIP270",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/gopayd.CreatePayment"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": ""
                    }
                }
            }
        },
        "/proofs/{txid}": {
            "post": {
                "description": "Creates a json envelope proof",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Proofs"
                ],
                "summary": "Create proof",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Transaction ID",
                        "name": "txid",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "JSON Envelope",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/envelope.JSONEnvelope"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": ""
                    }
                }
            }
        },
        "/txstatus/{txid}": {
            "get": {
                "description": "Returns status of transaction",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "TxStatus"
                ],
                "summary": "Transaction Status",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Transaction ID",
                        "name": "txid",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            }
        }
    },
    "definitions": {
        "bc.MapiCallback": {
            "type": "object",
            "properties": {
                "apiVersion": {
                    "type": "string"
                },
                "blockHash": {
                    "type": "string"
                },
                "blockHeight": {
                    "type": "integer"
                },
                "callbackPayload": {
                    "type": "string"
                },
                "callbackReason": {
                    "type": "string"
                },
                "callbackTxId": {
                    "type": "string"
                },
                "minerId": {
                    "type": "string"
                },
                "timestamp": {
                    "type": "string"
                }
            }
        },
        "bc.MerkleProof": {
            "type": "object",
            "properties": {
                "composite": {
                    "type": "boolean"
                },
                "index": {
                    "type": "integer"
                },
                "nodes": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "proofType": {
                    "type": "string"
                },
                "target": {
                    "type": "string"
                },
                "targetType": {
                    "type": "string"
                },
                "txOrId": {
                    "type": "string"
                }
            }
        },
        "envelope.JSONEnvelope": {
            "type": "object",
            "properties": {
                "encoding": {
                    "type": "string"
                },
                "mimetype": {
                    "type": "string"
                },
                "payload": {
                    "type": "string"
                },
                "publicKey": {
                    "type": "string"
                },
                "signature": {
                    "type": "string"
                }
            }
        },
        "gopayd.CreatePayment": {
            "type": "object",
            "properties": {
                "memo": {
                    "description": "Memo is a plain-text note from the customer to the payment host.",
                    "type": "string"
                },
                "merchantData": {
                    "description": "MerchantData is copied from PaymentDetails.merchantData.\nPayment hosts may use invoice numbers or any other data they require to match Payments to PaymentRequests.\nNote that malicious clients may modify the merchantData, so should be authenticated\nin some way (for example, signed with a payment host-only key).\nMaximum length is 10000 characters.",
                    "$ref": "#/definitions/gopayd.MerchantData"
                },
                "refundTo": {
                    "description": "RefundTo is a paymail to send a refund to should a refund be necessary.\nMaximum length is 100 characters",
                    "type": "string"
                },
                "spvEnvelope": {
                    "description": "SPVEnvelope which contains the details of previous transaction and Merkle proof of each input UTXO.\nSee https://tsc.bitcoinassociation.net/standards/spv-envelope/",
                    "$ref": "#/definitions/spv.Envelope"
                },
                "transaction": {
                    "description": "Transaction is a valid, signed Bitcoin transaction that fully\npays the PaymentRequest.\nThe transaction is hex-encoded and must NOT be prefixed with \"0x\".",
                    "type": "string"
                }
            }
        },
        "gopayd.InvoiceCreate": {
            "type": "object",
            "properties": {
                "satoshis": {
                    "type": "integer"
                }
            }
        },
        "gopayd.MerchantData": {
            "type": "object",
            "properties": {
                "address": {
                    "description": "Address is the merchants store / head office address.",
                    "type": "string"
                },
                "avatar": {
                    "description": "AvatarURL displays a canonical url to a merchants avatar.",
                    "type": "string"
                },
                "email": {
                    "description": "Email can be sued to contact the merchant about this transaction.",
                    "type": "string"
                },
                "extendedData": {
                    "description": "ExtendedData can be supplied if the merchant wishes to send some arbitrary data back to the wallet.",
                    "type": "object",
                    "additionalProperties": true
                },
                "name": {
                    "description": "MerchantName is a human readable string identifying the merchant.",
                    "type": "string"
                },
                "paymentReference": {
                    "description": "PaymentReference can be sent to link this request with a specific payment id.",
                    "type": "string"
                }
            }
        },
        "spv.Envelope": {
            "type": "object",
            "properties": {
                "mapiResponses": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/bc.MapiCallback"
                    }
                },
                "parents": {
                    "type": "object",
                    "additionalProperties": {
                        "$ref": "#/definitions/spv.Envelope"
                    }
                },
                "proof": {
                    "$ref": "#/definitions/bc.MerkleProof"
                },
                "rawTx": {
                    "type": "string"
                },
                "txid": {
                    "type": "string"
                }
            }
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "0.0.1",
	Host:        "localhost:8443",
	BasePath:    "/api/v1",
	Schemes:     []string{},
	Title:       "Payd",
	Description: "Payd is a txo and key manager, with a common interface that can be implemented by wallets.",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
		"escape": func(v interface{}) string {
			// escape tabs
			str := strings.Replace(v.(string), "\t", "\\t", -1)
			// replace " with \", and if that results in \\", replace that with \\\"
			str = strings.Replace(str, "\"", "\\\"", -1)
			return strings.Replace(str, "\\\\\"", "\\\\\\\"", -1)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}
