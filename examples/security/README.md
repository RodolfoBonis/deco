# 🔒 Proteção de Endpoints Internos

Este exemplo demonstra como proteger automaticamente os endpoints internos do gin-decorators (documentação, Swagger, etc.) para que só sejam acessíveis de redes locais/VPN.

## 🎯 Problema

Por padrão, os endpoints internos do gin-decorators são acessíveis publicamente:

- `/decorators/docs` - Documentação das rotas
- `/decorators/docs.json` - Documentação em JSON
- `/decorators/openapi.json` - Especificação OpenAPI
- `/decorators/openapi.yaml` - Especificação OpenAPI YAML
- `/decorators/swagger-ui` - Interface Swagger
- `/decorators/swagger` - Redirecionamento Swagger

Isso pode representar um risco de segurança em produção, pois expõe informações sobre a estrutura da API.

## ✅ Solução

O gin-decorators agora aplica automaticamente proteção de segurança nesses endpoints usando a função `Default()` ou `DefaultWithSecurity()`.

## 🚀 Como Usar

### 1. Proteção Padrão (Recomendado)

```go
package main

import deco "github.com/RodolfoBonis/deco"

func main() {
    // Aplica automaticamente proteção localhost-only
    r := deco.Default()
    r.Run(":8080")
}
```

**Resultado**: Apenas localhost (127.0.0.1) pode acessar os endpoints internos.

### 2. Proteção com Redes Privadas

```go
package main

import deco "github.com/RodolfoBonis/deco"

func main() {
    securityConfig := deco.DefaultSecurityConfig()
    securityConfig.AllowPrivateNetworks = true
    
    r := deco.DefaultWithSecurity(securityConfig)
    r.Run(":8080")
}
```

**Resultado**: Localhost + redes privadas (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16) podem acessar.

### 3. Proteção com Redes Específicas

```go
package main

import deco "github.com/RodolfoBonis/deco"

func main() {
    customSecurity := &deco.SecurityConfig{
        AllowedNetworks: []string{"192.168.1.0/24", "10.0.0.0/8"},
        AllowLocalhost:  true,
        ErrorMessage:    "Acesso negado: Endpoints restritos a redes autorizadas",
        LogBlockedAttempts: true,
    }
    
    r := deco.DefaultWithSecurity(customSecurity)
    r.Run(":8080")
}
```

### 4. Proteção com IPs Específicos

```go
package main

import deco "github.com/RodolfoBonis/deco"

func main() {
    ipSecurity := &deco.SecurityConfig{
        AllowedIPs: []string{"192.168.1.100", "10.0.0.50", "127.0.0.1"},
        ErrorMessage: "Acesso negado: Endpoints restritos a IPs autorizados",
        LogBlockedAttempts: true,
    }
    
    r := deco.DefaultWithSecurity(ipSecurity)
    r.Run(":8080")
}
```

## 🔧 Configurações Disponíveis

### SecurityConfig

```go
type SecurityConfig struct {
    // Redes permitidas em notação CIDR
    AllowedNetworks []string
    
    // IPs individuais permitidos
    AllowedIPs []string
    
    // Hostnames/domínios permitidos
    AllowedHosts []string
    
    // Permitir localhost/127.0.0.1
    AllowLocalhost bool
    
    // Permitir redes privadas (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16)
    AllowPrivateNetworks bool
    
    // Mensagem de erro personalizada
    ErrorMessage string
    
    // Log de tentativas bloqueadas
    LogBlockedAttempts bool
}
```

## 📋 Exemplos de Uso

### Desenvolvimento

```go
devSecurity := &deco.SecurityConfig{
    AllowPrivateNetworks: true,  // VPN, Docker, etc.
    AllowLocalhost: true,        // Desenvolvimento local
    ErrorMessage: "Acesso negado: Endpoints restritos ao ambiente de desenvolvimento",
    LogBlockedAttempts: true,
}
```

### Produção

```go
prodSecurity := &deco.SecurityConfig{
    AllowedNetworks: []string{"10.0.0.0/8", "172.16.0.0/12"}, // Redes da empresa
    AllowLocalhost: true,                                        // Para debugging
    AllowedIPs: []string{"192.168.1.100", "192.168.1.101"},     // Servidores de monitoramento
    ErrorMessage: "Acesso negado: Endpoints restritos à rede corporativa",
    LogBlockedAttempts: true,
}
```

## 🔍 Logs de Segurança

Quando `LogBlockedAttempts` está ativado, tentativas de acesso bloqueadas são logadas:

```
🔒 SECURITY: Blocked access to internal endpoint from 203.0.113.1 (Host: example.com, Path: /decorators/docs)
```

## 🚫 Resposta de Erro

Quando o acesso é bloqueado, o servidor retorna:

```json
{
    "error": "access_denied",
    "message": "Acesso negado: Este endpoint é restrito a redes internas",
    "details": "This endpoint is restricted to internal networks only"
}
```

## 🧪 Testando

1. Execute o exemplo:
   ```bash
   cd examples/security
   go run protect_internal_endpoints.go
   ```

2. Teste acesso local (deve funcionar):
   ```bash
   curl http://localhost:8080/decorators/docs
   ```

3. Teste acesso externo (deve ser bloqueado):
   ```bash
   curl -H "X-Forwarded-For: 203.0.113.1" http://localhost:8080/decorators/docs
   ```

## ⚠️ Importante

- **Desenvolvimento**: Use `AllowPrivateNetworks: true` para permitir VPN, Docker, etc.
- **Produção**: Configure redes específicas da empresa
- **Sempre**: Mantenha `LogBlockedAttempts: true` para monitoramento
- **Nunca**: Deixe endpoints internos abertos publicamente em produção 