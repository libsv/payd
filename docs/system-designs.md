# System Designs

## Basic HTTP/REST Model

This is the basic P4 (P2P Payment Protocol) model, previously known as BIP270, where the payment flow is done peer-to-peer instead of peer-to-blockchain-to-peer.

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

Actor Customer
Entity Merchant_P4
Actor Merchant
Entity mAPI

note over Customer: customer wants to buy something from merchant

autonumber
Customer-->Merchant: create basket of goods

group Socket setup
Merchant --> Merchant_P4: connect to socket
Merchant_P4 --> Merchant: get socket channel id + info
end

Merchant -> Merchant: create purchase order (PO)
Merchant -> Merchant: display bitcoin URI QR code

Customer --> Merchant: attempt to purchase goods (scan qr code)
Customer -> Merchant_P4 ++: fetch PaymentRequest

group Socket Communications
Merchant_P4 --> Merchant: fetch PaymentRequest
Merchant --> Merchant_P4: PaymentRequest
end
return PaymentRequest

Customer -> Customer: Build Payment
Customer -> Merchant_P4 ++: Payment
group Socket Communications
Merchant_P4 --> Merchant: Payment
end

Merchant -> Merchant: validate payment, store tx
Merchant -> mAPI: broadcast tx
mAPI -> Merchant: tx response

group Socket Communications
Merchant --> Merchant_P4: PaymentAck
end
return PaymentAck 
@enduml
```

## Websocket Model (using a P4 server as a proxy)

This setup just uses websockets all through (on the customer/sender side as well). This is a new setup that most other wallets in the ecosystem are not used to or have not seen before. Basically, the customer and merchant communicate with each other using websockets for the entire P4 flow with the P4 server acting as a proxy between them. The reason why the P4 server is needed is the same reasoning as above: because the merchant will not always be externally accessible over the internet.

```plantuml
@startuml
Actor Customer
Entity Merchant_P4
Actor Merchant
Entity mAPI 

skinparam responseMessageBelowArrow true

note over Customer: customer wants to buy something from merchant

autonumber
Customer-->Merchant: create basket of goods

group Socket setup (merchant)
Merchant --> Merchant_P4: connect to socket (join channel)
Merchant_P4 --> Merchant: get socket channel id/info (join channel status)
end

Merchant -> Merchant: create purchase order (PO)
Merchant -> Merchant: display bitcoin URI QR code (with ws:// payment URL)

Customer --> Merchant: attempt to purchase goods (scan qr code)

group Socket setup (customer)
Customer --> Merchant_P4: join channel 
Merchant_P4 --> Customer: join channel status 
end

group Socket Communication (payment)
Customer -> Merchant_P4: paymentrequest.create
Merchant_P4 -> Merchant: paymentrequest.create
Merchant -> Merchant: build payment request
Merchant --> Merchant_P4: paymentrequest.response
Merchant_P4 --> Customer: paymentrequest.response
Customer -> Customer: build and fund transaction
Customer -> Merchant_P4: payment
Merchant_P4 -> Merchant: payment
Merchant -> Merchant: validate payment, store tx
Merchant -> mAPI: broadcast transaction
mAPI -> Merchant: tx response
Merchant --> Merchant_P4: payment.ack
Merchant_P4 --> Customer: payment.ack
end

mAPI --> Merchant_P4: http proof call back

group Socket Communications (merkle proof)
Merchant_P4 --> Customer: merkle proof
Merchant_P4 --> Merchant: merkle proof
Merchant_P4 -> Customer: channel.expired
note right: after a specific time period has elapsed
Merchant_P4 -> Merchant: channel.expired
end
@enduml
```

### Channel setup

Each invoice payment flow would be achieved with a unique web socket 'channel', i.e. communication will occur on a common channel, setup by the merchant and communicated to the payer. This channel could be secured with a token or other mechanism but to start with it will be open.

All comms will happen in real time until the point the channel is killed by the merchant (usually once merkle proofs have been sent).