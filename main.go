package main

import (
	"backend_golang/handlers/auth"
	"backend_golang/handlers/users"
	"fmt"
	"net/http"
)

func main() {
	serv := http.NewServeMux()

	serv.HandleFunc("/auth/register", auth.Register)
	serv.HandleFunc("/auth/login", auth.Login)
	serv.HandleFunc("/auth/logout", auth.Logout)
	serv.HandleFunc("/auth/refresh", auth.RefreshSession)
	serv.HandleFunc("/auth/newPassword", auth.RefreshPassword)
	serv.HandleFunc("/auth/getBySession", auth.GetBySession)


	serv.HandleFunc("/users/getAll", users.GetAll)
	serv.HandleFunc("/users/getById", users.GetByID)


	fmt.Println("✅ Server started: http://localhost:8080")
	fmt.Println("📌 Endpoints for Postman:")

	
	fmt.Println("  USERS  ")
	fmt.Println("  GET  http://localhost:8080/users/getAll")
	fmt.Println("  GET http://localhost:8080/users/getById")


	fmt.Println("  AUTH  ")
	fmt.Println("  POST http://localhost:8080/auth/register")
	fmt.Println("  POST http://localhost:8080/auth/login")
	fmt.Println("  POST http://localhost:8080/auth/logout")
	fmt.Println("  POST http://localhost:8080/auth/refresh")
	fmt.Println("  POST http://localhost:8080/auth/newPassword")
	fmt.Println("  POST http://localhost:8080/auth/getBySession")

	http.ListenAndServe(":8080", serv)
}
