package main

import (
	"fmt"

	deco "github.com/RodolfoBonis/deco"
)

func main() {
	fmt.Println("🔒 Exemplo: Proteção dos endpoints internos do gin-decorators")

	// Exemplo 1: Proteção padrão (apenas localhost)
	fmt.Println("✅ Executando com proteção padrão (localhost apenas)")
	r := deco.Default() // Aplica automaticamente AllowLocalhostOnly()

	fmt.Println("📋 Endpoints protegidos:")
	fmt.Println("   - /decorators/docs (documentação)")
	fmt.Println("   - /decorators/docs.json (documentação JSON)")
	fmt.Println("   - /decorators/openapi.json (especificação OpenAPI)")
	fmt.Println("   - /decorators/openapi.yaml (especificação OpenAPI YAML)")
	fmt.Println("   - /decorators/swagger-ui (interface Swagger)")
	fmt.Println("   - /decorators/swagger (redirecionamento Swagger)")
	fmt.Println("\n🔒 Apenas localhost (127.0.0.1) pode acessar esses endpoints")
	fmt.Println("🌐 Teste acessando: http://localhost:8080/decorators/docs")
	fmt.Println("🚫 Tentativas de acesso externo serão bloqueadas e logadas")

	r.Run(":8080")
}

// Exemplo de como usar em produção
func productionExample() {
	// Configuração de segurança para produção
	prodSecurity := &deco.SecurityConfig{
		// Permitir apenas redes da empresa
		AllowedNetworks: []string{"10.0.0.0/8", "172.16.0.0/12"},
		// Permitir localhost para desenvolvimento
		AllowLocalhost: true,
		// IPs específicos de servidores de monitoramento
		AllowedIPs: []string{"192.168.1.100", "192.168.1.101"},
		// Mensagem de erro personalizada
		ErrorMessage: "Acesso negado: Endpoints de documentação restritos à rede corporativa",
		// Log de tentativas bloqueadas
		LogBlockedAttempts: true,
	}

	// Criar engine com segurança
	r := deco.DefaultWithSecurity(prodSecurity)

	// Adicionar suas rotas de aplicação
	// r.GET("/api/users", handlers.GetUsers)
	// r.POST("/api/users", handlers.CreateUser)

	r.Run(":8080")
}

// Exemplo de como usar em desenvolvimento
func developmentExample() {
	// Configuração de segurança para desenvolvimento
	devSecurity := &deco.SecurityConfig{
		// Permitir redes privadas (VPN, Docker, etc.)
		AllowPrivateNetworks: true,
		// Permitir localhost
		AllowLocalhost: true,
		// Mensagem de erro amigável
		ErrorMessage: "Acesso negado: Endpoints de documentação restritos ao ambiente de desenvolvimento",
		// Log de tentativas bloqueadas
		LogBlockedAttempts: true,
	}

	// Criar engine com segurança
	r := deco.DefaultWithSecurity(devSecurity)

	// Adicionar suas rotas de aplicação
	// r.GET("/api/users", handlers.GetUsers)
	// r.POST("/api/users", handlers.CreateUser)

	r.Run(":8080")
}
