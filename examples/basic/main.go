package main

import (
	deco "github.com/RodolfoBonis/deco"
	_ "github.com/RodolfoBonis/deco/examples/basic/.deco"
)

func main() {
	// O framework automatically carrega todas as rotas dos handlers
	r := deco.Default()

	// Servidor rodando na porta 8080
	// Acesse:
	// - http://localhost:8080/api/health
	// - http://localhost:8080/api/users
	// - http://localhost:8080/decorators/docs (documentação)
	// - http://localhost:8080/demo/websocket (teste WebSocket)
	r.Run(":8080")
}
