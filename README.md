# Order Book Manager (OBM)
The *OBM* aggregates different exchanges and exposes them through a single API
It provides two binaries:`obm` and `obm-ui`. 

# Installation
## Building from source
```
go install github.com/ParadigmFoundation/go-obm/cmd/obm

# optional
go install github.com/ParadigmFoundation/go-obm/cmd/obm-ui
```

# Usage
*OBM* currently supports _coinbase_, _gemini_ and _binance_. You can subscribe to one or more exchange using the flag `--exchange` with the syntax `--exachange=<name>:<symbol>[,symbol|...]`
For example to subscribe to the pair `BTC/USD` on  _coinbase_  and `BTC/USDT` and `ETH/USDT` on _binance_ use:
```bash
obm --exchange=coinbase:BTC/USD --exchange=binance:ETH/USDT,BTC/USDT
```
Once the `obm` service is running, you can check it's working running `obm-ui`
