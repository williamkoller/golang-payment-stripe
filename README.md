# API de Pagamentos com Stripe em Go

Uma API REST robusta e escalável para processamento de pagamentos usando Stripe, implementada em Go com Clean Architecture.

## 📋 Funcionalidades

- ✅ **Criação e Autorização de Pagamentos**: Criação de Payment Intents com autorização automática
- ✅ **Captura Manual**: Captura de fundos autorizados quando necessário
- ✅ **Cancelamento**: Cancelamento de autorizações não capturadas
- ✅ **Consulta de Pagamentos**: Busca detalhada de pagamentos por ID
- ✅ **Webhooks do Stripe**: Processamento automático de eventos do Stripe
- ✅ **Rate Limiting**: Proteção contra abuso com rate limiting configurável
- ✅ **Circuit Breaker**: Proteção contra falhas com padrão circuit breaker
- ✅ **Logging Estruturado**: Observabilidade completa com Zap
- ✅ **Validação Robusta**: Validação de entrada com go-playground/validator
- ✅ **Graceful Shutdown**: Encerramento gracioso do servidor

## 🏗️ Arquitetura

O projeto segue os princípios da **Clean Architecture**, com separação clara de responsabilidades:

```
├── cmd/api/                 # Ponto de entrada da aplicação
├── internal/
│   ├── domain/             # Entidades e regras de negócio
│   │   └── payment/        # Domínio de pagamentos
│   ├── app/                # Casos de uso e orquestração
│   │   ├── service/        # Serviços de aplicação
│   │   ├── saga/           # Padrão Saga para transações
│   │   └── ports/          # Interfaces/Contratos
│   └── infra/              # Infraestrutura e adapters
│       ├── http/           # HTTP handlers, middleware, routers
│       ├── stripe/         # Cliente Stripe
│       ├── repo/           # Repositórios de dados
│       ├── config/         # Configuração da aplicação
│       └── logger/         # Logging estruturado
└── pkg/                    # Utilitários compartilhados
    └── ulidx/              # Geração de ULIDs
```

## 🛠️ Tecnologias Utilizadas

- **Go 1.24.5** - Linguagem de programação
- **Gin Framework** - Framework HTTP rápido e minimalista
- **Stripe Go SDK** - Integração oficial com Stripe
- **Zap** - Logging estruturado de alta performance
- **Go Playground Validator** - Validação de estruturas
- **ULID** - Identificadores únicos ordenáveis
- **Circuit Breaker** - Padrão de resiliência
- **Rate Limiting** - Controle de taxa de requisições

## ⚙️ Configuração

### Variáveis de Ambiente

Crie um arquivo `.env` na raiz do projeto com as seguintes variáveis:

```bash
# Configuração da Aplicação
APP_ENV=dev
HTTP_PORT=8080
LOG_LEVEL=info

# Rate Limiting
RATE_LIMIT_RPS=10
RATE_LIMIT_BURST=20
REQUEST_TIMEOUT=15s

# Stripe Configuration
STRIPE_SECRET_KEY=sk_test_seu_secret_key_aqui
STRIPE_WEBHOOK_SECRET=whsec_seu_webhook_secret_aqui
STRIPE_ENABLE_TEST_PM=true
STRIPE_TEST_PAYMENT_METHOD=pm_card_visa

# Circuit Breaker
CB_MAX_REQUESTS=3
CB_INTERVAL=60s
CB_TIMEOUT=8s
```

### Configuração do Stripe

1. **Criar conta no Stripe**: [https://stripe.com](https://stripe.com)
2. **Obter chaves de API**: Acesse o Dashboard → Developers → API keys
3. **Configurar webhook**:
   - URL: `https://seu-dominio.com/webhooks/stripe`
   - Eventos: `payment_intent.succeeded`, `payment_intent.payment_failed`
   - Copie o signing secret para `STRIPE_WEBHOOK_SECRET`

## 🚀 Instalação e Execução

### Pré-requisitos

- Go 1.24.5 ou superior
- Conta no Stripe (teste ou produção)

### Passos

1. **Clone o repositório**:

```bash
git clone https://github.com/williamkoller/golang-payment-stripe.git
cd golang-payment-stripe
```

2. **Instale as dependências**:

```bash
go mod download
```

3. **Configure as variáveis de ambiente**:

```bash
cp .env.example .env
# Edite o arquivo .env com suas configurações
```

4. **Execute a aplicação**:

```bash
go run cmd/api/main.go
```

A API estará disponível em `http://localhost:8080`

## 📡 Endpoints da API

### Base URL

```
http://localhost:8080/v1
```

### 1. Criar e Autorizar Pagamento

**POST** `/v1/payments`

Cria um novo pagamento e realiza a autorização automática.

```bash
curl -X POST http://localhost:8080/v1/payments \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 5500,
    "currency": "brl",
    "email": "cliente@example.com"
  }'
```

**Resposta:**

```json
{
  "id": "01HXYZ123ABC456DEF789GHI",
  "amount": 5500,
  "currency": "brl",
  "email": "cliente@example.com",
  "status": "authorized",
  "stripe_payment_intent_id": "pi_1ABC123def456GHI",
  "client_secret": "pi_1ABC123def456GHI_secret_xyz",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### 2. Capturar Pagamento

**POST** `/v1/payments/{id}/capture`

Captura os fundos de um pagamento autorizado.

```bash
curl -X POST http://localhost:8080/v1/payments/01HXYZ123ABC456DEF789GHI/capture
```

### 3. Cancelar Pagamento

**POST** `/v1/payments/{id}/cancel`

Cancela a autorização de um pagamento não capturado.

```bash
curl -X POST http://localhost:8080/v1/payments/01HXYZ123ABC456DEF789GHI/cancel
```

### 4. Consultar Pagamento

**GET** `/v1/payments/{id}`

Obtém os detalhes de um pagamento específico.

```bash
curl http://localhost:8080/v1/payments/01HXYZ123ABC456DEF789GHI
```

### 5. Webhook do Stripe

**POST** `/webhooks/stripe`

Endpoint para receber eventos do Stripe (configurado automaticamente).

## 💳 Fluxo de Pagamento

### 1. Autorização (Auth)

```
Cliente → API → Stripe → API → Cliente
```

- Valida dados de entrada
- Cria Payment Intent no Stripe
- Retorna client_secret para o frontend
- Status: `authorized`

### 2. Captura (Capture)

```
Merchant → API → Stripe → Webhook → API
```

- Captura fundos autorizados
- Atualiza status via webhook
- Status: `captured`

### 3. Cancelamento (Cancel)

```
Merchant → API → Stripe → API
```

- Cancela autorização não capturada
- Libera reserva no cartão do cliente
- Status: `canceled`

## 📊 Estados do Pagamento

| Status       | Descrição                                |
| ------------ | ---------------------------------------- |
| `created`    | Pagamento criado, aguardando autorização |
| `authorized` | Autorizado, aguardando captura           |
| `captured`   | Fundos capturados com sucesso            |
| `canceled`   | Autorização cancelada                    |
| `failed`     | Falha no processamento                   |
| `refunded`   | Pagamento reembolsado                    |

## 🔒 Segurança

### Rate Limiting

- **RPS**: 10 requisições por segundo (configurável)
- **Burst**: 20 requisições em rajada (configurável)

### Validação de Webhooks

- Verificação de assinatura do Stripe
- Proteção contra replay attacks
- Validação de timestamp

### Validação de Entrada

- Validação rigorosa de todos os parâmetros
- Sanitização de dados de entrada
- Tratamento seguro de erros

## 🔧 Padrões de Resiliência

### Circuit Breaker

- **Max Requests**: 3 falhas consecutivas
- **Interval**: 60 segundos
- **Timeout**: 8 segundos

### Retry e Timeout

- Timeout configurável para requisições
- Tratamento de erros com context
- Propagação adequada de cancelamentos

## 📈 Observabilidade

### Logging

- Logs estruturados em JSON
- Níveis configuráveis (debug, info, warn, error)
- Correlação de requests com trace IDs

### Métricas

- Logging de performance de requests
- Rastreamento de erros e falhas
- Monitoramento de circuit breaker

## 🧪 Testes

```bash
# Executar todos os testes
go test ./...

# Testes com coverage
go test -cover ./...

# Testes com verbose
go test -v ./...
```

## 🐛 Troubleshooting

### Problemas Comuns

1. **Erro de autenticação Stripe**

   - Verifique se `STRIPE_SECRET_KEY` está correto
   - Confirme se está usando a chave do ambiente correto (test/live)

2. **Webhook não funcionando**

   - Verifique se `STRIPE_WEBHOOK_SECRET` está correto
   - Confirme se a URL do webhook está acessível
   - Teste a conectividade: `curl -X POST sua-url/webhooks/stripe`

3. **Rate limit atingido**
   - Ajuste `RATE_LIMIT_RPS` e `RATE_LIMIT_BURST`
   - Implemente retry com backoff no cliente

---

Desenvolvido com ❤️ por [William Koller](https://github.com/williamkoller)
