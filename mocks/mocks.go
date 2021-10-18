package mocks

//go:generate moq -pkg mocks -out destination_service.go ../ DestinationsService
//go:generate moq -pkg mocks -out timestamp_service.go ../ TimestampService
//go:generate moq -pkg mocks -out private_key_service.go ../ PrivateKeyService
//go:generate moq -pkg mocks -out seed_service.go ../ SeedService

//go:generate moq -pkg mocks -out transacter.go ../ Transacter
//go:generate moq -pkg mocks -out fee_reader.go ../ FeeReader
//go:generate moq -pkg mocks -out owner_store.go ../ OwnerStore
//go:generate moq -pkg mocks -out proofs_writer.go ../ ProofsWriter
//go:generate moq -pkg mocks -out tx_writer.go ../ TransactionWriter
//go:generate moq -pkg mocks -out broadcast_writer.go ../ BroadcastWriter
//go:generate moq -pkg mocks -out derivation_reader.go ../ DerivationReader
//go:generate moq -pkg mocks -out proof_callback_writer.go ../ ProofCallbackWriter
//go:generate moq -pkg mocks -out invoice_reader_writer.go ../ InvoiceReaderWriter
//go:generate moq -pkg mocks -out private_key_reader_writer.go ../ PrivateKeyReaderWriter
//go:generate moq -pkg mocks -out destination_reader_writer.go ../ DestinationsReaderWriter

// third party

//go:generate moq -pkg mocks -out payment_verifier.go ../vendor/github.com/libsv/go-bc/spv PaymentVerifier
