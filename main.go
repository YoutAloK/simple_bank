package main

import (
	"backend_golang/handlers/auth"
	"backend_golang/handlers/users"
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", auth.Register)
		authGroup.POST("/login", auth.Login)
		authGroup.DELETE("/logout", auth.Logout)
		authGroup.PUT("/refresh", auth.RefreshSession)
		authGroup.PUT("/password", auth.RefreshPassword)
		authGroup.GET("/session", auth.GetBySession)
	}

	usersGroup := r.Group("/users")
	{
		usersGroup.GET("", users.GetAll)
		usersGroup.GET("/:id", users.GetByID)
	}

	fmt.Println("✅ Server started: http://localhost:8080")
	fmt.Println("📌 Endpoints for Postman:")

	fmt.Println("\n  USERS  ")
	fmt.Println("  GET    http://localhost:8080/users")
	fmt.Println("  GET    http://localhost:8080/users/:id")

	fmt.Println("\n  AUTH  ")
	fmt.Println("  POST   http://localhost:8080/auth/register")
	fmt.Println("  POST   http://localhost:8080/auth/login")
	fmt.Println("  DELETE http://localhost:8080/auth/logout")
	fmt.Println("  PUT    http://localhost:8080/auth/refresh")
	fmt.Println("  PUT    http://localhost:8080/auth/password")
	fmt.Println("  GET    http://localhost:8080/auth/session")

	r.Run(":8080")
}
