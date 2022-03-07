# System Designs

## Basic HTTP/REST Model

This is the basic DPP (Direct Payment Protocol) model, previously known as BIP270, where the payment flow is done peer-to-peer instead of peer-to-blockchain-to-peer.

> Please note that the merchant should use TLS/HTTPS when exposing their Bitcoin URI in order for the customer to be secure against a MITM (man in the middle) attack.

```plantuml
@startuml

hide footbox
title P2P Payment Protocol Flow

actor Customer
actor Merchant
box "Node/Miner"
    boundary MerchantAPI
    control bitcoind
end box

note over Customer: Customer wants to buy something from Merchant

autonumber
Customer-->Merchant: create basket of goods
Merchant->Merchant: create invoice (X satoshis, etc.)
Merchant->Merchant: display invoice [pay:url] \n(QR code or Bitcoin URI)

Merchant-->Customer: obtain payment protocol URL \n(scan QR code or click link)

group Payment Procotol - BIP (2)70
  Customer->Merchant: fetch Payment Request

  ' group if no un-expired quote available
  Merchant-->MerchantAPI ++: GET policyQuote() \n(if no un-expired quote available)
  autonumber stop
  MerchantAPI -\ bitcoind
  MerchantAPI \- bitcoind
  autonumber 7
  return policyQuote Response
  ' end

  Merchant->Merchant: construct Payment Request arguments \n(destinations, fees, etc.)

  Merchant->Customer: Payment Request

  Customer->Customer: Authorize? (click "OK")

  Customer->Customer: Construct Payment message

  Customer->Merchant: Payment

  Merchant->Merchant: Validate Payment Tx

  alt Invalid Tx

    Merchant->Customer: reject payment (error)

  else Valid Tx
    
    autonumber 14
    Merchant->MerchantAPI ++: POST submitTransaction(Tx)
    autonumber stop
    MerchantAPI -\ bitcoind
    MerchantAPI \- bitcoind
    autonumber 15
    return transactionResponse

    Merchant->Customer: Payment ACK

    MerchantAPI->Merchant: Merkle Proof
    MerchantAPI->Customer: Merkle Proof

  end

end
@enduml
```

## Advanced REST/HTTP Model (using REST API externally + websockets internally) 

There exists a limitation with the above basic REST/HTTP model: namely that the merchant/receiver must be externally accessible on the internet. For example if you run your wallet on you laptop connected to the internet through your home WiFi router connection, you won't be able to receive payments as detailed above. This is because you will likely be using IPv4, where there aren't enough addresses for everyone in the world (unlike IPv6), so your laptop's IP address will probably be an internal/private IP address that gets packets forwarded to it by the router using NAT (network address translation). Unless your computer is using a public IP address that is publicly accessible, you won't (easily) be able to do this. Another issue is that connecting to a base IPv4 address is not secure against MITM (man in the middle) attacks since there is no authentication involved. To combat this you will need TLS/HTTPS for authentication as mentioned above.

This setup is little more advanced than the basic flow above. Externally (from the sender/customer's point of view) it looks identical to the basic flow, however internally (under the hood, between the merchant's wallet and the merchant's P4 server) websockets are used. Here the merchant's P4 server is hosted somewhere externally accessible on the internet (using TLS/HTTPS) while the merchant can run their wallet anywhere. When the merchant's wallet starts up, it will connect to the P4 server and create a websocket connection to it. Then, it will use that socket channel id when exposing/displaying its Bitcoin URI so that when the customer hits the P4 server, the P4 server will know where to send a message over the websocket and then get a message back from the merchant's wallet with the `PaymentRequest` and then the P4 server will respond to the REST API call with that.

```plantuml
@startuml 

Actor Sender

box Receiver Proxy
Participant PCS
Participant DP3
end box

box Receiver
Participant Wallet
Participant BHC
end box

Entity mAPI

note over Sender: customer wants to buy something from merchant

autonumber
Sender-->Wallet: create basket of goods

group Socket setup
Wallet --> DP3: connect to socket
DP3 --> Wallet: get socket channel id + info
end

Wallet -> Wallet: create purchase order (PO)
Wallet -> Wallet: display bitcoin URI QR code

Sender --> Wallet: attempt to purchase goods (scan qr code)
Sender -> DP3 ++: GET Payment Terms

group Socket Communications
DP3 --> Wallet: Payment Terms
Wallet --> DP3: PaymentRequest
end
return PaymentRequest

Sender -> Sender: Build Payment
Sender -> DP3 ++: POST Payment
group Socket Communications
DP3 --> Wallet: Payment
end

Wallet -> Wallet: validate payment, store tx

Wallet -> PCS: Create channel
PCS -> Wallet: Channel ID

Wallet -> mAPI: broadcast tx
mAPI -> Wallet: tx response

group Socket Communications
Wallet --> DP3: PaymentAck
end
return PaymentAck 

Sender -> PCS: Subscribe to channel
Wallet -> PCS: Subscribe to channel
mAPI -> PCS: Merkle Proof
PCS -> Sender: Merkle Proof Notification
PCS -> Wallet: Merkle Proof Notification

Wallet <-> BHC: SPV Check

@enduml
```
