package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// HealthCheck verifica o status da API.
// @Route("GET", "/api/health")
// @Summary("Health Check")
// @Description("Verifica o status da API e retorna informações básicas do serviço")
// @Tag("health")
// @Response(200, description="API está funcionando")
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":   "gin-decorators-example",
		"status":    "ok",
		"timestamp": 1234567890,
	})
}

// ListUsers retorna a lista de usuários.
// @Route("GET", "/api/users")
// @Summary("List Users")
// @Description("Retorna uma lista paginada de usuários do sistema")
// @Tag("users")
// @Param(name="page", type="int", location="query", required=false, description="Número da página")
// @Param(name="limit", type="int", location="query", required=false, description="Quantidade de itens por página")
// @Response(200, description="Lista de usuários")
func ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	users := []gin.H{
		{"id": 1, "name": "John Silva", "email": "john@example.com"},
		{"id": 2, "name": "Mary Santos", "email": "mary@example.com"},
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": len(users),
		},
	})
}

// CreateUser cria um novo usuário.
// @Route("POST", "/api/users")
// @Summary("Create User")
// @Description("Cria um novo usuário no sistema")
// @Tag("users")
// @ValidateJSON()
// @RequestBody(type="object", description="Dados do usuário a ser criado")
// @Response(201, description="Usuário criado com sucesso")
// @Response(400, description="Dados inválidos")
func CreateUser(c *gin.Context) {
	var body struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newUser := gin.H{
		"id":    3,
		"name":  body.Name,
		"email": body.Email,
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    newUser,
	})
}

// GetUser retorna um usuário pelo ID.
// @Route("GET", "/api/users/:id")
// @Summary("Get User")
// @Description("Retorna um usuário específico pelo ID")
// @Tag("users")
// @Param(name="id", type="int", location="path", required=true, description="ID do usuário")
// @Response(200, description="Usuário encontrado")
// @Response(400, description="ID inválido")
// @Response(404, description="Usuário não encontrado")
func GetUser(c *gin.Context) {
	id := c.Param("id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	user := gin.H{
		"id":    userID,
		"name":  "John Silva",
		"email": "john@example.com",
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}
