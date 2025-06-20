# Sistema de Leilões com Fechamento Automático

Este projeto implementa um sistema de leilões em Go com funcionalidade de fechamento automático baseado em tempo definido em variáveis de ambiente.

## Funcionalidades Implementadas

### ✅ Funcionalidades Existentes
- Criação de leilões (auctions)
- Sistema de lances (bids)
- Validação de status do leilão na criação de lances
- API REST com Gin framework
- Persistência em MongoDB

### 🆕 Nova Funcionalidade: Fechamento Automático
- **Goroutine de verificação**: Executa periodicamente para verificar leilões vencidos
- **Cálculo de tempo**: Baseado em variáveis de ambiente configuráveis
- **Fechamento automático**: Leilões são fechados automaticamente quando expiram
- **Logs detalhados**: Registro de todas as operações de fechamento
- **Graceful shutdown**: Parada segura da goroutine

## Arquitetura da Solução

### 1. Função de Cálculo de Tempo
- **Arquivo**: `internal/infra/database/auction/create_auction.go`
- **Funções**: `getAuctionInterval()` e `getAuctionCheckInterval()`
- **Configuração**: Via variáveis de ambiente `AUCTION_INTERVAL` e `AUCTION_CHECK_INTERVAL`

### 2. Goroutine de Fechamento Automático
- **Arquivo**: `internal/infra/database/auction/auction_closer.go`
- **Classe**: `AuctionCloser`
- **Funcionalidades**:
  - Execução periódica baseada em `AUCTION_CHECK_INTERVAL`
  - Busca de leilões ativos
  - Verificação de expiração baseada em `AUCTION_INTERVAL`
  - Atualização automática do status para `Completed`

### 3. Métodos de Repositório Adicionados
- **Arquivo**: `internal/infra/database/auction/create_auction.go`
- **Métodos**:
  - `UpdateAuctionStatus()`: Atualiza o status do leilão
  - `FindActiveAuctionsToCheck()`: Busca leilões ativos para verificação

### 4. Integração no Main
- **Arquivo**: `cmd/auction/main.go`
- **Funcionalidades**:
  - Inicialização do `AuctionCloser`
  - Início da goroutine na inicialização
  - Graceful shutdown com `signal.Notify`

## Configuração de Ambiente

Crie o arquivo `cmd/auction/.env` com as seguintes variáveis:

```env
# MongoDB Configuration
MONGODB_URL=mongodb://localhost:27017
MONGODB_DB=auction_db

# Auction Configuration
AUCTION_INTERVAL=5m
AUCTION_CHECK_INTERVAL=1m

# Bid Configuration
BATCH_INSERT_INTERVAL=3m
MAX_BATCH_SIZE=5
```

### Variáveis de Ambiente Explicadas

- **AUCTION_INTERVAL**: Duração do leilão (ex: `5m` = 5 minutos)
- **AUCTION_CHECK_INTERVAL**: Intervalo para verificar leilões vencidos (ex: `1m` = 1 minuto)
- **BATCH_INSERT_INTERVAL**: Intervalo para inserção em lote de lances
- **MAX_BATCH_SIZE**: Tamanho máximo do lote de lances

## Como Executar o Projeto

### 1. Usando Docker Compose (Recomendado)

```bash
# Construir e executar com Docker Compose
docker-compose up --build

# Executar em background
docker-compose up -d --build

# Parar os serviços
docker-compose down
```

### 2. Execução Local

#### Pré-requisitos
- Go 1.20+
- MongoDB rodando localmente

#### Passos
```bash
# 1. Instalar dependências
go mod tidy

# 2. Configurar variáveis de ambiente
cp cmd/auction/.env.example cmd/auction/.env
# Editar o arquivo .env conforme necessário

# 3. Executar o projeto
go run cmd/auction/main.go

# Ou compilar e executar
go build -o auction cmd/auction/main.go
./auction
```

## Testes

### Executar Todos os Testes
```bash
go test ./...
```

### Executar Testes Específicos
```bash
# Testes do fechamento automático
go test ./internal/infra/database/auction/ -v

# Testes com cobertura
go test ./internal/infra/database/auction/ -v -cover
```

### Testes Implementados
- `TestAuctionCloser_CheckAndCloseExpiredAuctions`: Valida o fechamento automático
- `TestAuctionCloser_StartAndStop`: Valida o ciclo de vida da goroutine

## Endpoints da API

### Leilões
- `GET /auction` - Listar leilões
- `GET /auction/:auctionId` - Buscar leilão por ID
- `POST /auction` - Criar novo leilão
- `GET /auction/winner/:auctionId` - Buscar vencedor do leilão

### Lances
- `POST /bid` - Criar novo lance
- `GET /bid/:auctionId` - Listar lances de um leilão

### Usuários
- `GET /user/:userId` - Buscar usuário por ID

## Exemplo de Uso

### 1. Criar um Leilão
```bash
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "iPhone 15",
    "category": "Electronics",
    "description": "iPhone 15 Pro Max 256GB",
    "condition": 1
  }'
```

### 2. Fazer um Lance
```bash
curl -X POST http://localhost:8080/bid \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-uuid",
    "auction_id": "auction-uuid",
    "amount": 1500.00
  }'
```

## Logs e Monitoramento

O sistema gera logs detalhados para:
- Início da goroutine de fechamento
- Leilões fechados automaticamente
- Erros durante o processo
- Estatísticas de processamento

### Exemplo de Logs
```
{"level":"info","message":"Auction closer routine started","check_interval":"1m0s","auction_interval":"5m0s"}
{"level":"info","message":"Auction closed automatically","auction_id":"abc-123","product_name":"iPhone 15","created_at":"2024-01-01T10:00:00Z","expired_at":"2024-01-01T10:05:00Z","closed_at":"2024-01-01T10:05:30Z"}
```

## Concorrência e Thread Safety

A solução implementa:
- **Mutex**: Para operações concorrentes no repositório
- **Channels**: Para comunicação entre goroutines
- **Context**: Para cancelamento e timeout
- **WaitGroup**: Para sincronização de goroutines

## Melhorias Implementadas

1. **Fechamento Automático**: Leilões são fechados automaticamente após o tempo configurado
2. **Configuração Flexível**: Intervalos configuráveis via variáveis de ambiente
3. **Logs Estruturados**: Usando zap logger com campos estruturados
4. **Graceful Shutdown**: Parada segura da aplicação
5. **Testes Unitários**: Cobertura de testes para a nova funcionalidade
6. **Documentação**: README completo com instruções de uso

## Estrutura de Arquivos Modificados

```
internal/
├── entity/auction_entity/
│   └── auction_entity.go          # Interface atualizada
├── infra/database/auction/
│   ├── create_auction.go          # Métodos de atualização
│   ├── auction_closer.go          # Nova goroutine
│   └── auction_closer_test.go     # Testes
└── cmd/auction/
    └── main.go                    # Integração da goroutine
```

## Considerações de Performance

- **Intervalo de Verificação**: Configurável para balancear performance e responsividade
- **Batch Processing**: Lances são processados em lotes
- **Índices MongoDB**: Recomenda-se criar índices para `status` e `timestamp`
- **Memory Usage**: Goroutine com controle de memória via context

## Troubleshooting

### Problemas Comuns

1. **Leilões não fecham automaticamente**
   - Verificar se `AUCTION_INTERVAL` está configurado
   - Verificar logs da goroutine

2. **Erro de conexão MongoDB**
   - Verificar se MongoDB está rodando
   - Verificar `MONGODB_URL` no arquivo .env

3. **Goroutine não inicia**
   - Verificar logs de inicialização
   - Verificar se não há erros de compilação

### Logs de Debug
Para mais detalhes, verifique os logs da aplicação que mostram:
- Início da goroutine
- Intervalos configurados
- Leilões processados
- Erros encontrados 

## 🧪 Testes Automatizados da Aplicação

Este projeto inclui dois scripts de teste para facilitar a validação do fluxo de leilão, lances e fechamento automático:

### 1. Teste Geral da Aplicação

O script [`test_app.sh`](./test_app.sh) executa uma bateria de testes básicos, incluindo:
- Criação de leilão
- Criação/consulta de usuário
- Realização de lances
- Listagem de lances e leilões
- Consulta de vencedor

**Como usar:**
```bash
bash test_app.sh
```

---

### 2. Teste Específico do Fechamento Automático de Leilões

O script [`test_auction_closer.sh`](./test_auction_closer.sh) valida o fluxo de:
- Criação de leilão
- Fechamento automático após o tempo configurado (`AUCTION_INTERVAL`)
- Tentativa de lance em leilão encerrado (deve ser rejeitado)

**Como usar:**
```bash
bash test_auction_closer.sh
```

---

> **Dica:**  
> Para acompanhar os logs do fechamento automático em tempo real, utilize:
> ```bash
> docker-compose logs -f app
> ```

Certifique-se de que a aplicação está rodando (por exemplo, via Docker Compose) antes de executar os scripts. 