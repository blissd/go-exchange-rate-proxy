# Golang Exchange Rate Proxy

The impeccable Jin Bal has produced a Scala solution to the Landoop
_Exchange Rate API Proxy_ exercise. See his work here -> https://github.com/jinbal/exchange-rates-api-proxy

# Exercise

Create a REST application with a single endpoint :

GET /api/convert
parameters:
```json
{
"fromCurrency": "GBP",
"toCurrency" : "EUR",
"amount" : 102.6
}
```

The return should be an object with the exchange rate between the "fromCurrency" to "toCurrency" and the amount converted to the second curency.

```json
{
"exchange" : 1.11,
"amount" : 113.886,
"original" : 102.6
}
```

The exchange rates should be loaded from https://developers.coinbase.com/api/v2?shell#exchange-rates rates change every 1 minute.

## How to Run

First run the server process.
```shell
go run server/main.go
```

The POST a JSON message with `curl`

```shell
curl -X POST -H "Content-Type: application/json" \
  -d'{"fromCurrency":"GBP", "toCurrency":"EUR", "amount":2.0}' \
  http://localhost:8080/api/convert

```
