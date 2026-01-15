// Package users for get info
// from db
package users

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	DataBase "backend_golang/dataBase"
	"backend_golang/methods"
	"backend_golang/types"

	_ "github.com/go-sql-driver/mysql"
)

func GetAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methods.SendJSONErrorResponse(w, "Только GET запросы", "METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
		return
	}

	rows, err := DataBase.DB.Query("SELECT id, name, surname, phone_number, balance FROM users ORDER BY id")
	if err != nil {
		methods.SendJSONErrorResponse(w, "Ошибка базы данных: "+err.Error(), "DATABASE_ERROR", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := make([]types.UserResponse, 0)

	for rows.Next() {
		var user types.UserResponse
		if err := rows.Scan(&user.ID, &user.Name, &user.Surname, &user.PhoneNumber, &user.Balance); err != nil {
			methods.SendJSONErrorResponse(w, "Ошибка сканирования пользователя: "+err.Error(), "SCAN_ERROR", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	if len(users) == 0 {
		methods.SendJSONSuccessResponseWithData(w, "Нет пользователей в базе данных", []types.UserResponse{}, http.StatusOK)
		return
	}

	methods.SendJSONSuccessResponseWithData(w, fmt.Sprintf("Найдено %d пользователей", len(users)), users, http.StatusOK)
}

func GetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methods.SendJSONErrorResponse(w, "Только GET запросы", "METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
		return
	}

	idUser := r.URL.Query().Get("id")
	if idUser == "" {
		methods.SendJSONErrorResponse(w, "Параметр 'id' обязателен", "MISSING_FIELDS", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idUser)
	if err != nil {
		methods.SendJSONErrorResponse(w, "Неверный формат параметра 'id'", "INVALID_ID", http.StatusBadRequest)
		return
	}

	var user types.UserResponse
	err = DataBase.DB.QueryRow(
		"SELECT id, name, surname, phone_number, balance FROM users WHERE id = ?",
		id,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Surname,
		&user.PhoneNumber,
		&user.Balance,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			methods.SendJSONErrorResponse(w, "Пользователь не найден", "USER_NOT_FOUND", http.StatusNotFound)
		} else {
			methods.SendJSONErrorResponse(w, "Ошибка базы данных: "+err.Error(), "DATABASE_ERROR", http.StatusInternalServerError)
		}
		return
	}

	methods.SendJSONSuccessResponseWithData(w, "Пользователь найден", user, http.StatusOK)
}
