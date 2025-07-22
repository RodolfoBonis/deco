# Guia de Uso do Deco

## Visão Geral

O Deco é uma ferramenta poderosa para decoradores em Go que simplifica o desenvolvimento de APIs RESTful com recursos avançados como cache, rate limiting, validação, telemetria e muito mais.

## Instalação

```bash
go install github.com/RodolfoBonis/deco/cmd/deco@latest
```

## Configuração Inicial

### 1. Inicializar o Projeto

```bash
# No diretório do seu projeto
deco init
```

Isso criará:
- `decorators.go` - Arquivo principal de decoradores
- `.deco/config.yaml` - Configuração do projeto
- `.gitignore` - Configuração do Git

### 2. Configuração Básica

```yaml
# .deco/config.yaml
project:
  name: "meu-projeto"
  version: "1.0.0"
  description: "API REST com decoradores"

patterns:
  - "**/*.go"
  - "!**/*_test.go"
  - "!vendor/**"

output:
  directory: "./generated"
  format: "go"

middlewares:
  cache:
    enabled: true
    type: "memory"
    ttl: "5m"
  
  rate_limiting:
    enabled: true
    limit: 100
    window: "1m"
  
  validation:
    enabled: true
    fail_fast: false
  
  telemetry:
    enabled: true
    service_name: "meu-projeto"
    endpoint: "http://localhost:4318"
```

## Decoradores Disponíveis

### 1. Cache (@Cache)

Armazena respostas em cache para melhorar performance.

```go
// @Cache(ttl=5m, key=user_id)
func GetUser(c *gin.Context) {
    userID := c.Param("id")
    // ... lógica do handler
}
```

**Opções:**
- `ttl`: Tempo de vida do cache (ex: "5m", "1h")
- `key`: Chave personalizada para o cache
- `type`: Tipo de cache ("memory", "redis")

### 2. Rate Limiting (@RateLimit)

Controla a taxa de requisições por cliente.

```go
// @RateLimit(limit=100, window=1m, key=ip)
func CreateUser(c *gin.Context) {
    // ... lógica do handler
}
```

**Opções:**
- `limit`: Número máximo de requisições
- `window`: Janela de tempo (ex: "1m", "1h")
- `key`: Chave para identificação (ex: "ip", "user_id")

### 3. Validação (@Validate)

Valida dados de entrada automaticamente.

```go
type User struct {
    Name  string `json:"name" validate:"required,min=2"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"required,gte=18"`
}

// @Validate(schema=User)
func CreateUser(c *gin.Context) {
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    // ... lógica do handler
}
```

### 4. Autenticação (@Auth)

Protege endpoints com autenticação.

```go
// @Auth(required=true, roles=admin,user)
func AdminEndpoint(c *gin.Context) {
    // ... lógica do handler
}
```

**Opções:**
- `required`: Se a autenticação é obrigatória
- `roles`: Lista de roles permitidos

### 5. Telemetria (@Trace)

Adiciona rastreamento automático.

```go
// @Trace(operation=create_user, attributes=user_id,email)
func CreateUser(c *gin.Context) {
    // ... lógica do handler
}
```

### 6. Proxy (@Proxy)

Configura proxy reverso para outros serviços.

```go
// @Proxy(target=http://api-service:8080, timeout=10s, retries=3)
func ProxyToService(c *gin.Context) {
    // O middleware de proxy cuida de tudo automaticamente
}
```

**Opções:**
- `target`: URL do serviço de destino
- `timeout`: Timeout da requisição
- `retries`: Número de tentativas
- `circuit_breaker`: Configuração do circuit breaker

### 7. WebSocket (@WebSocket)

Configura endpoints WebSocket.

```go
// @WebSocket(path=/ws, groups=chat,notifications)
func WebSocketHandler(c *gin.Context) {
    // O middleware WebSocket cuida da conexão
}
```

## Exemplos Práticos

### API REST Completa

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/RodolfoBonis/deco/pkg/decorators"
)

type User struct {
    ID    string `json:"id" validate:"required"`
    Name  string `json:"name" validate:"required,min=2"`
    Email string `json:"email" validate:"required,email"`
}

// @Cache(ttl=10m, key=user_id)
// @RateLimit(limit=1000, window=1h)
// @Trace(operation=get_user)
func GetUser(c *gin.Context) {
    userID := c.Param("id")
    
    // Simular busca no banco
    user := &User{
        ID:    userID,
        Name:  "João Silva",
        Email: "joao@example.com",
    }
    
    c.JSON(200, user)
}

// @Validate(schema=User)
// @RateLimit(limit=100, window=1m)
// @Auth(required=true, roles=admin)
// @Trace(operation=create_user)
func CreateUser(c *gin.Context) {
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    // Simular criação
    user.ID = "generated-id"
    
    c.JSON(201, user)
}

func main() {
    r := gin.Default()
    
    // Aplicar decoradores automaticamente
    decorators.ApplyDecorators(r)
    
    r.GET("/users/:id", GetUser)
    r.POST("/users", CreateUser)
    
    r.Run(":8080")
}
```

### Configuração Avançada

```go
// Configuração personalizada de cache
// @Cache(ttl=1h, type=redis, key=user_id, endpoint=localhost:6379)
func GetUserProfile(c *gin.Context) {
    // ... lógica
}

// Rate limiting por IP com burst
// @RateLimit(limit=100, window=1m, burst=10, key=ip)
func SearchUsers(c *gin.Context) {
    // ... lógica
}

// Validação com custom validators
// @Validate(schema=User, custom=phone,cpf)
func UpdateUser(c *gin.Context) {
    // ... lógica
}

// Proxy com circuit breaker
// @Proxy(target=http://payment-service:8080, circuit_breaker=30s, failure_threshold=5)
func ProcessPayment(c *gin.Context) {
    // ... lógica
}
```

## Middlewares de Monitoramento

### Métricas

```go
// Habilitar métricas
r.Use(decorators.MetricsMiddleware(&decorators.MetricsConfig{
    Enabled: true,
    Path:    "/metrics",
}))

// Acessar métricas
// GET /metrics
```

### Health Check

```go
// Health check automático
r.GET("/health", decorators.HealthCheckHandler())
```

### Documentação OpenAPI

```go
// Gerar documentação automática
r.GET("/docs", decorators.DocsHandler(&decorators.DocsConfig{
    Title:       "Minha API",
    Description: "API REST com decoradores",
    Version:     "1.0.0",
}))
```

## Testes

### Executar Testes

```bash
# Executar todos os testes
make test

# Executar com cobertura
go test ./pkg/decorators -cover

# Executar testes específicos
go test ./pkg/decorators -run TestCacheMiddleware
```

### Cobertura Atual

- **Cobertura Total**: 61.5%
- **Testes Unitários**: 200+ testes
- **Testes de Integração**: Incluídos
- **Race Conditions**: Resolvidas

## Melhorias Implementadas

### 1. Race Conditions Resolvidas

- Implementado mutex thread-safe para `gin.SetMode()`
- Cache de proxy managers com sincronização adequada
- Telemetria com proteção contra acesso concorrente

### 2. Testes Melhorados

- Removidos `t.Parallel()` problemáticos
- Configurações únicas para evitar cache compartilhado
- Testes mais robustos e determinísticos

### 3. Cobertura Aumentada

- Adicionados testes para edge cases
- Cenários de erro cobertos
- Validação de estruturas de dados

### 4. Performance Otimizada

- Cache thread-safe
- Rate limiting otimizado
- Proxy com circuit breaker

## Troubleshooting

### Problemas Comuns

1. **Cache não funcionando**
   - Verificar configuração do Redis
   - Confirmar TTL configurado

2. **Rate limiting muito restritivo**
   - Ajustar `limit` e `window`
   - Verificar chave de identificação

3. **Validação falhando**
   - Verificar tags de validação
   - Confirmar schema correto

4. **Proxy não conectando**
   - Verificar URL do target
   - Confirmar timeout configurado

### Logs

```bash
# Habilitar logs detalhados
export DECO_LOG_LEVEL=verbose

# Executar com debug
deco generate --debug
```

## Próximos Passos

1. **Aumentar Cobertura**: Alcançar 80% de cobertura
2. **Testes de Performance**: Benchmarks automatizados
3. **Documentação**: Mais exemplos e casos de uso
4. **Integração**: CI/CD com testes automáticos

## Suporte

- **Issues**: [GitHub Issues](https://github.com/RodolfoBonis/deco/issues)
- **Documentação**: [docs/](docs/)
- **Exemplos**: [examples/](examples/)
