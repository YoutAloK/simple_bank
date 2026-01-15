// Package users for get info
// from db
package users

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"backend_golang/database"
	"backend_golang/types"
)

func GetAll(c *gin.Context) {
	rows, err := database.DB.Query("SELECT id, name, surname, phone_number, balance FROM users ORDER BY id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "Ошибка базы данных: " + err.Error(),
			Error:   "DATABASE_ERROR",
		})
		return
	}
	defer rows.Close()

	users := make([]types.UserResponse, 0)

	for rows.Next() {
		var user types.UserResponse
		if err := rows.Scan(&user.ID, &user.Name, &user.Surname, &user.PhoneNumber, &user.Balance); err != nil {
			c.JSON(http.StatusInternalServerError, types.Response{
				Success: false,
				Message: "Ошибка сканирования пользователя: " + err.Error(),
				Error:   "SCAN_ERROR",
			})
			return
		}
		users = append(users, user)
	}

	if len(users) == 0 {
		c.JSON(http.StatusOK, types.Response{
			Success: true,
			Message: "Нет пользователей в базе данных",
			Data:    []types.UserResponse{},
		})
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Success: true,
		Message: fmt.Sprintf("Найдено %d пользователей", len(users)),
		Data:    users,
	})
}

func GetByID(c *gin.Context) {
	idUser := c.Param("id")

	id, err := strconv.Atoi(idUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.Response{
			Success: false,
			Message: "Неверный формат параметра 'id'",
			Error:   "INVALID_ID",
		})
		return
	}

	var user types.UserResponse
	err = database.DB.QueryRow(
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
			c.JSON(http.StatusNotFound, types.Response{
				Success: false,
				Message: "Пользователь не найден",
				Error:   "USER_NOT_FOUND",
			})
		} else {
			c.JSON(http.StatusInternalServerError, types.Response{
				Success: false,
				Message: "Ошибка базы данных: " + err.Error(),
				Error:   "DATABASE_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Success: true,
		Message: "Пользователь найден",
		Data:    user,
	})
}