#!/bin/bash

echo "🧪 Teste Específico do Fechamento Automático de Leilões"
echo "======================================================"

BASE_URL="http://localhost:8080"

# Cores para output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "\n${BLUE}📋 Passo 1: Criar um leilão para teste${NC}"
auction_data='{
    "product_name": "Test Auction for Auto Close",
    "category": "Test",
    "description": "This auction will be automatically closed",
    "condition": 1
}'

response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auction" \
    -H "Content-Type: application/json" \
    -d "$auction_data")

http_code=$(echo "$response" | tail -n1)
response_body=$(echo "$response" | head -n -1)

if [ "$http_code" -eq 201 ]; then
    echo -e "${GREEN}✅ Leilão criado com sucesso${NC}"
else
    echo -e "${RED}❌ Erro ao criar leilão (HTTP $http_code)${NC}"
    echo "Response: $response_body"
    exit 1
fi

echo -e "\n${BLUE}📋 Passo 2: Listar leilões ativos e extrair o ID${NC}"
auctions_response=$(curl -s "$BASE_URL/auction?status=0")
echo "Leilões ativos: $auctions_response"

# Extrair o ID do leilão criado (assume que é o último com nome 'Test Auction for Auto Close')
closed_auction_id=$(echo "$auctions_response" | grep -o '{[^}]*}' | grep 'Test Auction for Auto Close' | tail -1 | sed -E 's/.*"id":"([^"]+)".*/\1/')
if [ -z "$closed_auction_id" ]; then
    # fallback: pega o primeiro id
    closed_auction_id=$(echo "$auctions_response" | grep -o '"id":"[^"]\+"' | head -1 | cut -d '"' -f4)
fi

echo -e "${YELLOW}ID do leilão para testar fechamento: $closed_auction_id${NC}"

# Se não encontrou, aborta
test -z "$closed_auction_id" && { echo -e "${RED}❌ Não foi possível extrair o ID do leilão${NC}"; exit 1; }

echo -e "\n${BLUE}📋 Passo 3: Verificar logs da aplicação${NC}"
echo "Verificando se a goroutine de fechamento está rodando..."
docker-compose logs app --tail=20 | grep -E "auction|closer|routine|closed|expired" || echo "Nenhum log específico encontrado"

echo -e "\n${BLUE}📋 Passo 4: Aguardar o fechamento automático${NC}"
echo "Aguardando 25 segundos para o fechamento automático (AUCTION_INTERVAL=20s)..."
sleep 25

echo -e "\n${BLUE}📋 Passo 5: Verificar se o leilão foi fechado${NC}"
auctions_after_response=$(curl -s "$BASE_URL/auction?status=1")
echo "Leilões completados: $auctions_after_response"

# Extrair o ID do leilão fechado
completed_auction_id=$(echo "$auctions_after_response" | grep -o '{[^}]*}' | grep 'Test Auction for Auto Close' | tail -1 | sed -E 's/.*"id":"([^"]+)".*/\1/')
if [ -z "$completed_auction_id" ]; then
    completed_auction_id=$(echo "$auctions_after_response" | grep -o '"id":"[^"]\+"' | head -1 | cut -d '"' -f4)
fi

echo -e "${YELLOW}ID do leilão fechado: $completed_auction_id${NC}"

echo -e "\n${BLUE}📋 Passo 6: Verificar logs de fechamento${NC}"
docker-compose logs app --tail=50 | grep -E "auction|closer|routine|closed|expired" || echo "Nenhum log de fechamento encontrado"

echo -e "\n${BLUE}📋 Passo 7: Tentar fazer um lance em leilão fechado${NC}"
bid_data='{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "auction_id": "'$completed_auction_id'",
    "amount": 2000.00
}'

bid_response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/bid" \
    -H "Content-Type: application/json" \
    -d "$bid_data")

bid_http_code=$(echo "$bid_response" | tail -n1)
bid_response_body=$(echo "$bid_response" | head -n -1)

if [ "$bid_http_code" -eq 201 ]; then
    echo -e "${RED}❌ Lance aceito (ERRO: o leilão deveria estar fechado)${NC}"
else
    echo -e "${GREEN}✅ Lance rejeitado (leilão fechado corretamente)${NC}"
    echo "Response: $bid_response_body"
fi

echo -e "\n${GREEN}🎉 Teste concluído!${NC}"
echo -e "\n${YELLOW}📝 Para monitorar em tempo real:${NC}"
echo "docker-compose logs -f app" 