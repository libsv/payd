package mocks

//go:generate moq -pkg mocks -out destination_service.go ../ DestinationsService
//go:generate moq -pkg mocks -out timestamp_service.go ../ TimestampService
//go:generate moq -pkg mocks -out private_key_service.go ../ PrivateKeyService
//go:generate moq -pkg mocks -out envelope_service.go ../ EnvelopeService
//go:generate moq -pkg mocks -out seed_service.go ../ SeedService
//go:generate moq -pkg mocks -out peerchannels_service.go ../ PeerChannelsService
//go:generate moq -pkg mocks -out peerchannels_notify_service.go ../ PeerChannelsNotifyService

//go:generate moq -pkg mocks -out transacter.go ../ Transacter
//go:generate moq -pkg mocks -out fee_quote_reader.go ../ FeeQuoteReader
//go:generate moq -pkg mocks -out fee_quote_fetcher.go ../ FeeQuoteFetcher
//go:generate moq -pkg mocks -out txo_writer.go ../ TxoWriter
//go:generate moq -pkg mocks -out owner_store.go ../ OwnerStore
//go:generate moq -pkg mocks -out proofs_writer.go ../ ProofsWriter
//go:generate moq -pkg mocks -out tx_writer.go ../ TransactionWriter
//go:generate moq -pkg mocks -out broadcast_writer.go ../ BroadcastWriter
//go:generate moq -pkg mocks -out derivation_reader.go ../ DerivationReader
//go:generate moq -pkg mocks -out peerchannels_store.go ../ PeerChannelsStore
//go:generate moq -pkg mocks -out proof_callback_writer.go ../ ProofCallbackWriter
//go:generate moq -pkg mocks -out invoice_reader_writer.go ../ InvoiceReaderWriter
//go:generate moq -pkg mocks -out private_key_reader_writer.go ../ PrivateKeyReaderWriter
//go:generate moq -pkg mocks -out destination_reader_writer.go ../ DestinationsReaderWriter
//go:generate moq -pkg mocks -out dpp.go ../data/http DPP

// third party

//go:generate moq -pkg mocks -out payment_verifier.go ../vendor/github.com/libsv/go-bc/spv PaymentVerifier
//go:generate moq -pkg mocks -out envelope_creator.go ../vendor/github.com/libsv/go-bc/spv EnvelopeCreator
