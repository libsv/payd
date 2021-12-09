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
        "/api/v1/balance": {
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
        "/api/v1/destinations/{invoiceID}": {
            "get": {
                "description": "Given an invoiceID, a set of outputs and fees will be returned, if not found a 404 is returned.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Destinations",
                    "Receive"
                ],
                "summary": "Given an invoiceID, a set of outputs and fees will be returned, if not found a 404 is returned.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Invoice ID",
                        "name": "invoiceID",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": ""
                    },
                    "404": {
                        "description": "returned if the invoiceID has not been found",
                        "schema": {
                            "$ref": "#/definitions/payd.ClientError"
                        }
                    }
                }
            }
        },
        "/api/v1/invoices": {
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
                "description": "Creates an invoices with invoiceID and satoshis",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Invoices"
                ],
                "summary": "InvoiceCreate invoices",
                "parameters": [
                    {
                        "description": "Reference and Satoshis",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/payd.InvoiceCreate"
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
        "/api/v1/invoices/{invoiceID}": {
            "get": {
                "description": "Returns invoices by invoices id if exists",
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
                        "description": "Invoice ID",
                        "name": "invoiceID",
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
            "delete": {
                "description": "InvoiceDelete",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Invoices"
                ],
                "summary": "InvoiceDelete invoices",
                "parameters": [
                    {
                        "type": "string",
                        "description": "invoiceID we want to remove",
                        "name": "invoiceID",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": ""
                    },
                    "404": {
                        "description": "returned if the paymentID has not been found",
                        "schema": {
                            "$ref": "#/definitions/payd.ClientError"
                        }
                    }
                }
            }
        },
        "/api/v1/owner": {
            "get": {
                "description": "Returns information about the wallet owner",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Users"
                ],
                "summary": "Wallet owner information.",
                "responses": {
                    "200": {
                        "description": "Current wallet owner",
                        "schema": {
                            "$ref": "#/definitions/payd.User"
                        }
                    }
                }
            }
        },
        "/api/v1/pay": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Pay"
                ],
                "summary": "Make a payment",
                "parameters": [
                    {
                        "description": "Pay to url",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/payd.PayRequest"
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
        "/api/v1/payment/{paymentID}": {
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
                "summary": "Request to pay an invoice and receive back outputs to use when constructing the payment transaction",
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
                    "201": {
                        "description": "contains outputs, merchant data and expiry information, used by the payee to construct a transaction",
                        "schema": {
                            "$ref": "#/definitions/payd.PaymentRequestResponse"
                        }
                    },
                    "400": {
                        "description": "returned if the user input is invalid, usually an issue with the paymentID",
                        "schema": {
                            "$ref": "#/definitions/payd.ClientError"
                        }
                    },
                    "404": {
                        "description": "returned if the paymentID has not been found",
                        "schema": {
                            "$ref": "#/definitions/payd.ClientError"
                        }
                    },
                    "500": {
                        "description": "returned if there is an unexpected internal error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/api/v1/payments/{invoiceID}": {
            "post": {
                "description": "Given an invoiceID, and an spvEnvelope, we will validate the payment and inputs used are valid and that it covers the invoice.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Payments"
                ],
                "summary": "Validate and store a payment.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Invoice ID",
                        "name": "invoiceID",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": ""
                    },
                    "400": {
                        "description": "returned if the invoiceID is empty or payment isn't valid",
                        "schema": {
                            "$ref": "#/definitions/payd.ClientError"
                        }
                    },
                    "404": {
                        "description": "returned if the invoiceID has not been found",
                        "schema": {
                            "$ref": "#/definitions/payd.ClientError"
                        }
                    }
                }
            }
        },
        "/api/v1/proofs/{txid}": {
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
                "summary": "InvoiceCreate proof",
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
        }
    },
    "definitions": {
        "bt.FeeQuote": {
            "type": "object"
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
        "payd.ClientError": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "string",
                    "example": "N01"
                },
                "id": {
                    "type": "string",
                    "example": "e97970bf-2a88-4bc8-90e6-2f597a80b93d"
                },
                "message": {
                    "type": "string",
                    "example": "unable to find foo when loading bar"
                },
                "title": {
                    "type": "string",
                    "example": "not found"
                }
            }
        },
        "payd.InvoiceCreate": {
            "type": "object",
            "properties": {
                "description": {
                    "description": "Description is an optional text field that can have some further info\nlike 'invoice for oranges'.\nMaxLength is 1024 characters.",
                    "type": "string"
                },
                "expiresAt": {
                    "description": "ExpiresAt is an optional param that can be passed to set an expiration\ndate on an invoice, after which, payments will not be accepted.",
                    "type": "string"
                },
                "reference": {
                    "description": "Reference is an identifier that can be used to link the\npayd invoice with an external system.\nMaxLength is 32 characters.",
                    "type": "string"
                },
                "satoshis": {
                    "description": "Satoshis is the total amount this invoice is to pay.",
                    "type": "integer"
                }
            }
        },
        "payd.P4Destination": {
            "type": "object",
            "properties": {
                "outputs": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/payd.P4Output"
                    }
                }
            }
        },
        "payd.P4Output": {
            "type": "object",
            "properties": {
                "amount": {
                    "type": "integer"
                },
                "description": {
                    "type": "string"
                },
                "script": {
                    "type": "string"
                }
            }
        },
        "payd.PayRequest": {
            "type": "object",
            "properties": {
                "payToURL": {
                    "type": "string"
                }
            }
        },
        "payd.PaymentRequestResponse": {
            "type": "object",
            "properties": {
                "creationTimestamp": {
                    "type": "string"
                },
                "destinations": {
                    "$ref": "#/definitions/payd.P4Destination"
                },
                "expirationTimestamp": {
                    "type": "string"
                },
                "fees": {
                    "$ref": "#/definitions/bt.FeeQuote"
                },
                "memo": {
                    "type": "string"
                },
                "merchantData": {
                    "$ref": "#/definitions/payd.User"
                },
                "network": {
                    "type": "string"
                },
                "paymentURL": {
                    "type": "string"
                },
                "spvRequired": {
                    "type": "boolean",
                    "example": true
                }
            }
        },
        "payd.User": {
            "type": "object",
            "properties": {
                "address": {
                    "type": "string"
                },
                "avatar": {
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "extendedData": {
                    "type": "object",
                    "additionalProperties": true
                },
                "id": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                },
                "phoneNumber": {
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
	BasePath:    "/api",
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
	swag.Register("swagger", &s{})
}
