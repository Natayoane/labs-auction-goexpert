#!/bin/bash

echo "🚀 Testando Sistema de Leilões com Fechamento Automático"
echo "=================================================="

# URL base da aplicação
BASE_URL="http://localhost:8080"

# Cores para output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Função para fazer requisições e mostrar resultado
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    echo -e "\n${YELLOW}📋 $description${NC}"
    echo "Endpoint: $method $endpoint"
    
    if [ -n "$data" ]; then
        echo "Data: $data"
        response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint")
    fi
    
    # Separar response body e status code
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "${GREEN}✅ Sucesso (HTTP $http_code)${NC}"
        echo "Response: $response_body"
    else
        echo -e "${RED}❌ Erro (HTTP $http_code)${NC}"
        echo "Response: $response_body"
    fi
}

# Aguardar um pouco para garantir que a aplicação está pronta
echo "⏳ Aguardando aplicação inicializar..."
sleep 5

# Teste 1: Verificar se a aplicação está respondendo
echo -e "\n${YELLOW}🔍 Teste 1: Verificar se a aplicação está respondendo${NC}"
if curl -s "$BASE_URL/auction" > /dev/null; then
    echo -e "${GREEN}✅ Aplicação está respondendo${NC}"
else
    echo -e "${RED}❌ Aplicação não está respondendo${NC}"
    exit 1
fi

# Teste 2: Criar um leilão
echo -e "\n${YELLOW}🔍 Teste 2: Criar um leilão${NC}"
auction_data='{
    "product_name": "iPhone 15 Pro Max",
    "category": "Electronics",
    "description": "iPhone 15 Pro Max 256GB Titanium",
    "condition": 1
}'

test_endpoint "POST" "/auction" "$auction_data" "Criando leilão do iPhone 15"

# Extrair o ID do leilão criado (assumindo que retorna um ID)
# Como a API não retorna o ID no POST, vamos listar os leilões para pegar o primeiro
echo -e "\n${YELLOW}📋 Listando leilões para obter IDs${NC}"
auctions_response=$(curl -s "$BASE_URL/auction")
echo "Leilões disponíveis: $auctions_response"

# Teste 3: Criar um usuário (se necessário)
echo -e "\n${YELLOW}🔍 Teste 3: Verificar usuário${NC}"
# Assumindo que existe um usuário com ID conhecido ou criando um
user_id="550e8400-e29b-41d4-a716-446655440000"
test_endpoint "GET" "/user/$user_id" "" "Buscando usuário"

# Teste 4: Fazer um lance (se tivermos um auction_id válido)
echo -e "\n${YELLOW}🔍 Teste 4: Fazer um lance${NC}"
# Vamos usar um ID de exemplo - em produção você extrairia do response anterior
auction_id="550e8400-e29b-41d4-a716-446655440001"
bid_data='{
    "user_id": "'$user_id'",
    "auction_id": "'$auction_id'",
    "amount": 1500.00
}'

test_endpoint "POST" "/bid" "$bid_data" "Fazendo lance no leilão"

# Teste 5: Listar lances de um leilão
echo -e "\n${YELLOW}🔍 Teste 5: Listar lances${NC}"
test_endpoint "GET" "/bid/$auction_id" "" "Listando lances do leilão"

# Teste 6: Buscar vencedor do leilão
echo -e "\n${YELLOW}🔍 Teste 6: Buscar vencedor${NC}"
test_endpoint "GET" "/auction/winner/$auction_id" "" "Buscando vencedor do leilão"

# Teste 7: Listar leilões por status
echo -e "\n${YELLOW}🔍 Teste 7: Listar leilões ativos${NC}"
test_endpoint "GET" "/auction?status=0" "" "Listando leilões ativos"

# Teste 8: Listar leilões por categoria
echo -e "\n${YELLOW}🔍 Teste 8: Listar leilões por categoria${NC}"
test_endpoint "GET" "/auction?category=Electronics" "" "Listando leilões da categoria Electronics"

echo -e "\n${GREEN}🎉 Testes concluídos!${NC}"
echo -e "\n${YELLOW}📝 Próximos passos para testar o fechamento automático:${NC}"
echo "1. Aguarde o tempo configurado em AUCTION_INTERVAL (padrão: 5 minutos)"
echo "2. Verifique os logs do container para ver o fechamento automático:"
echo "   docker-compose logs app"
echo "3. Após o fechamento, teste novamente fazer um lance para verificar que foi rejeitado"
echo "4. Verifique o status do leilão mudou para 'Completed' (status=1)"

echo -e "\n${YELLOW}🔍 Para ver os logs em tempo real:${NC}"
echo "docker-compose logs -f app" 