# Bitcoin Core - Atividades

Interfaces web em Go para interagir com um Bitcoin Core local.

- `atividade1`: dashboard de blockchain/mempool.
- `atividade2`: eventos ZMQ de blocos e transacoes.
- `atividade3`: wallets, envio e acompanhamento de transacoes.

## Requisitos

No Ubuntu Server:

```bash
sudo apt update
sudo apt install -y build-essential curl git libzmq3-dev pkg-config
```

Instale tambem:

- Go compativel com o `go.mod`.
- Docker e Docker Compose.
- Bitcoin Core com `bitcoind` e `bitcoin-cli` no `PATH`.

Instalacao manual do Bitcoin Core:

```bash
tar -xzf bitcoin-*-x86_64-linux-gnu.tar.gz
sudo install -m 0755 -o root -g root -t /usr/local/bin bitcoin-*/bin/*
bitcoind -version
bitcoin-cli -version
```

## Bitcoin Core

Crie o arquivo:

```bash
mkdir -p ~/.bitcoin
nano ~/.bitcoin/bitcoin.conf
```

Config para `signet`:

```ini
[signet]
server=1
fallbackfee=0.00001
rpcbind=0.0.0.0
rpcallowip=127.0.0.1
rpcallowip=172.16.0.0/12
rpcuser=bitcoin
rpcpassword=bitcoin
zmqpubhashblock=tcp://0.0.0.0:28332
zmqpubhashtx=tcp://0.0.0.0:28333
```

Suba o node:

```bash
bitcoind -signet
bitcoin-cli -signet getblockchaininfo
```

Wallet em signet:

```bash
bitcoin-cli -signet createwallet teste
bitcoin-cli -signet -rpcwallet=teste getnewaddress
bitcoin-cli -signet -rpcwallet=teste getbalance
```

Em signet, use faucet para receber moedas. Nao da para minerar blocos localmente como no regtest.

Config opcional para `regtest` no mesmo `bitcoin.conf`:

```ini
[regtest]
server=1
fallbackfee=0.00001
rpcbind=0.0.0.0
rpcallowip=127.0.0.1
rpcallowip=172.16.0.0/12
rpcuser=bitcoin
rpcpassword=bitcoin
zmqpubhashblock=tcp://0.0.0.0:28332
zmqpubhashtx=tcp://0.0.0.0:28333
```

## Rodar com Docker Compose

O compose sobe apenas as atividades. O Bitcoin Core roda fora do Docker, no host.

Signet:

```bash
BITCOIN_NETWORK=signet BITCOIN_RPC_PORT=38332 docker compose up --build atividade2 atividade3
```

Regtest:

```bash
docker compose up --build atividade1 atividade2 atividade3
```

Acessos:

```text
http://localhost:8080  atividade1
http://localhost:8081  atividade2
http://localhost:8082  atividade3
```

De outro PC na mesma rede, troque `localhost` pelo IP do servidor:

```bash
hostname -I
```

## Rodar sem Docker

```bash
go mod download
go run ./cmd/atividade1
go run ./cmd/atividade2
go run ./cmd/atividade3
```

Para rodar em signet sem Docker:

```bash
BITCOIN_NETWORK=signet BITCOIN_RPC_PORT=38332 go run ./cmd/atividade3
```

## Cloudflare Tunnel

Configure os public hostnames apontando para os servicos do compose:

```text
atividade1 -> http://atividade1:8080
atividade2 -> http://atividade2:8081
atividade3 -> http://atividade3:8082
```

Depois:

```bash
cp .env.example .env
docker compose --profile tunnel up --build
```

## Comandos uteis

```bash
bitcoin-cli -signet listwallets
bitcoin-cli -signet listwalletdir
bitcoin-cli -signet loadwallet teste
bitcoin-cli -signet -rpcwallet=teste getnewaddress
bitcoin-cli -signet getrawmempool
bitcoin-cli -signet -rpcwallet=teste gettransaction TXID
```

Enviar transacao:

```bash
bitcoin-cli -signet -rpcwallet=teste -named sendtoaddress \
  address="tb1..." \
  amount=0.001 \
  fee_rate=1
```

Parar:

```bash
bitcoin-cli -signet stop
docker compose down
```

## Testes

```bash
go test ./...
```

