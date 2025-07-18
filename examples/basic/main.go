package main

import (
	deco "github.com/RodolfoBonis/deco"
	_ "github.com/RodolfoBonis/deco/examples/basic/.deco"
)

func main() {
	// O framework automatically carrega todas as rotas dos handlers
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

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
