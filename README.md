# Bitcoin Core - Atividades

Projeto em Go com interfaces web para interagir com um node Bitcoin Core local em `regtest`.

- `cmd/atividade1`: dashboard simples de mempool e blockchain.
- `cmd/atividade2`: dashboard de eventos ZMQ (`hashblock` e `hashtx`).
- `cmd/atividade3`: dashboard de wallets, envio e acompanhamento de transacoes.

## Pre-requisitos

Instale:

- Go `1.26.1` ou compativel com o `go.mod`.
- Bitcoin Core instalado no host, com `bitcoind` e `bitcoin-cli` disponiveis no `PATH`.
- Docker e Docker Compose, caso queira executar as atividades em containers.
- Dependencias de ZMQ do sistema, necessarias para a atividade 2.

No Ubuntu/Debian:

```bash
sudo apt update
sudo apt install -y build-essential curl libzmq3-dev pkg-config
```

Confira as instalacoes:

```bash
go version
bitcoind -version
bitcoin-cli -version
```

## Instalar Bitcoin Core no Ubuntu Server

Baixe o Bitcoin Core pelo site oficial:

```text
https://bitcoincore.org/en/download/
```

No Ubuntu Server, escolha o arquivo Linux correspondente a arquitetura da maquina. Para a maioria dos notebooks e PCs, use `x86_64-linux-gnu.tar.gz`.

Exemplo usando uma versao baixada manualmente:

```bash
tar -xzf bitcoin-*-x86_64-linux-gnu.tar.gz
sudo install -m 0755 -o root -g root -t /usr/local/bin bitcoin-*/bin/*
```

Valide:

```bash
bitcoind -version
bitcoin-cli -version
```

## Configurar o Bitcoin Core em regtest

Crie o diretorio de configuracao:

```bash
mkdir -p ~/.bitcoin
```

Crie ou edite o arquivo `~/.bitcoin/bitcoin.conf`:

```ini
regtest=1
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

Essas opcoes permitem que os containers das atividades acessem o Bitcoin Core rodando no host pela rede do Docker. Para uso local fora do Docker, os mesmos comandos continuam funcionando.

Inicie o node:

```bash
bitcoind -regtest
```

Em outro terminal, valide:

```bash
bitcoin-cli -regtest getblockchaininfo
```

Para parar o node:

```bash
bitcoin-cli -regtest stop
```

## Preparar wallet e saldo

Crie uma wallet:

```bash
bitcoin-cli -regtest createwallet teste
```

Gere um endereco da wallet:

```bash
bitcoin-cli -regtest -rpcwallet=teste getnewaddress
```

Minere blocos para gerar saldo em regtest:

```bash
ADDR=$(bitcoin-cli -regtest -rpcwallet=teste getnewaddress)
bitcoin-cli -regtest generatetoaddress 101 "$ADDR"
```

Confira saldo:

```bash
bitcoin-cli -regtest -rpcwallet=teste getbalance
```

## Executar o projeto

Baixe as dependencias Go:

```bash
go mod download
```

Rode a atividade desejada a partir da raiz do repositorio.

Atividade 1:

```bash
go run ./cmd/atividade1
```

Acesse:

```text
http://localhost:8080
```

Atividade 2:

```bash
go run ./cmd/atividade2
```

Acesse:

```text
http://localhost:8081
```

Atividade 3:

```bash
go run ./cmd/atividade3
```

Acesse:

```text
http://localhost:8082
```

## Executar com Docker Compose

O `docker-compose.yml` sobe apenas as aplicacoes do projeto. O Bitcoin Core deve estar instalado e rodando no host Ubuntu Server.

- `atividade1`: dashboard de mempool/blockchain em `http://localhost:8080`.
- `atividade2`: dashboard ZMQ em `http://localhost:8081`.
- `atividade3`: dashboard de wallets em `http://localhost:8082`.
- `cloudflared`: opcional, para expor os dashboards via Cloudflare Tunnel.

Antes de subir o compose, deixe o Bitcoin Core rodando no host:

```bash
bitcoind -regtest
```

Suba os servicos principais:

```bash
docker compose up --build atividade1 atividade2 atividade3
```

Em outro terminal, crie a wallet e gere saldo usando o `bitcoin-cli` do host:

```bash
bitcoin-cli -regtest createwallet teste
bitcoin-cli -regtest -rpcwallet=teste getnewaddress
```

Minere 101 blocos para liberar saldo:

```bash
ADDR=$(bitcoin-cli -regtest -rpcwallet=teste getnewaddress)
bitcoin-cli -regtest generatetoaddress 101 "$ADDR"
```

Para confirmar transacoes no regtest:

```bash
ADDR=$(bitcoin-cli -regtest -rpcwallet=teste getnewaddress)
bitcoin-cli -regtest generatetoaddress 1 "$ADDR"
```

Para parar:

```bash
docker compose down
```

Para apagar o historico local persistido pela atividade 3:

```bash
docker compose down -v
```

Esse comando nao apaga a blockchain nem as wallets do Bitcoin Core instalado no host. Para parar o node manualmente:

```bash
bitcoin-cli -regtest stop
```

### Cloudflare Tunnel

Crie um tunnel nomeado no painel da Cloudflare e configure os public hostnames apontando para os servicos internos do compose:

- atividade 2: `http://atividade2:8081`
- atividade 3: `http://atividade3:8082`
- atividade 1: `http://atividade1:8080`

Depois copie o arquivo de exemplo e preencha o token:

```bash
cp .env.example .env
```

Suba com o profile do tunnel:

```bash
docker compose --profile tunnel up --build
```

## Usar a atividade 3

A interface deixa claro que a transacao sera criada e assinada no contexto da wallet selecionada.

Fluxo recomendado:

1. Abra `http://localhost:8082`.
2. Selecione a wallet desejada.
3. Informe o endereco de destino.
4. Informe o valor em BTC.
5. Clique em `Criar, assinar e transmitir`.
6. No regtest, mine 1 bloco para confirmar a transacao.

Comando para minerar 1 bloco:

```bash
ADDR=$(bitcoin-cli -regtest -rpcwallet=teste getnewaddress)
bitcoin-cli -regtest generatetoaddress 1 "$ADDR"
```

A lista de transacoes enviadas mostra o `txid`, status, confirmacoes e `wallet: nome_da_wallet`.

O backend tambem grava um indice local em `cmd/atividade3/tracked_transactions.json` com `txid`, `wallet` e `sent_at`. Esse arquivo nao substitui a blockchain; ele apenas lembra quais transacoes foram enviadas pela interface depois que o servidor Go reinicia.

## Comandos uteis

Listar wallets carregadas:

```bash
bitcoin-cli -regtest listwallets
```

Listar wallets disponiveis no diretorio:

```bash
bitcoin-cli -regtest listwalletdir
```

Carregar uma wallet:

```bash
bitcoin-cli -regtest loadwallet teste
```

Gerar um endereco novo:

```bash
bitcoin-cli -regtest -rpcwallet=teste getnewaddress
```

Enviar direto pelo `bitcoin-cli`:

```bash
bitcoin-cli -regtest -rpcwallet=teste -named sendtoaddress \
  address="bcrt1..." \
  amount=1 \
  fee_rate=1
```

Ver detalhes de uma transacao:

```bash
bitcoin-cli -regtest -rpcwallet=teste gettransaction TXID
```

Ver mempool:

```bash
bitcoin-cli -regtest getrawmempool
```

## Testes

Rode:

```bash
go test ./...
```
