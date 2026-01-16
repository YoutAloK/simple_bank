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

func UpdateProfile(c *gin.Context) {
	userIDParam := c.Param("id")

	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.Response{
			Success: false,
			Message: "Неверный формат параметра 'id'",
			Error:   "INVALID_ID",
		})
		return
	}

	var updateData struct {
		Name        *string `json:"name"`
		Surname     *string `json:"surname"`
		PhoneNumber *string `json:"phone_number"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, types.Response{
			Success: false,
			Message: "Неверный формат данных: " + err.Error(),
			Error:   "INVALID_JSON",
		})
		return
	}

	if updateData.Name == nil && updateData.Surname == nil && updateData.PhoneNumber == nil {
		c.JSON(http.StatusBadRequest, types.Response{
			Success: false,
			Message: "Необходимо указать хотя бы одно поле для обновления",
			Error:   "NO_FIELDS_TO_UPDATE",
		})
		return
	}

	var exists bool
	err = database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", userID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "Ошибка базы данных: " + err.Error(),
			Error:   "DATABASE_ERROR",
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, types.Response{
			Success: false,
			Message: "Пользователь не найден",
			Error:   "USER_NOT_FOUND",
		})
		return
	}

	query := "UPDATE users SET "
	args := []interface{}{}
	updates := []string{}

	if updateData.Name != nil {
		updates = append(updates, "name = ?")
		args = append(args, *updateData.Name)
	}
	if updateData.Surname != nil {
		updates = append(updates, "surname = ?")
		args = append(args, *updateData.Surname)
	}
	if updateData.PhoneNumber != nil {
		updates = append(updates, "phone_number = ?")
		args = append(args, *updateData.PhoneNumber)
	}

	query += updates[0]
	for i := 1; i < len(updates); i++ {
		query += ", " + updates[i]
	}
	query += " WHERE id = ?"
	args = append(args, userID)

	result, err := database.DB.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "Не удалось обновить профиль: " + err.Error(),
			Error:   "UPDATE_ERROR",
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "Ошибка при проверке результата",
			Error:   "RESULT_CHECK_ERROR",
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, types.Response{
			Success: false,
			Message: "Пользователь не был обновлен",
			Error:   "NOT_UPDATED",
		})
		return
	}

	var user types.UserResponse
	err = database.DB.QueryRow(
		"SELECT id, name, surname, phone_number, balance FROM users WHERE id = ?",
		userID,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Surname,
		&user.PhoneNumber,
		&user.Balance,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "Ошибка получения обновленных данных: " + err.Error(),
			Error:   "FETCH_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Success: true,
		Message: "Профиль успешно обновлен",
		Data:    user,
	})
}

func UserDelete(c *gin.Context) {
	userIDParam := c.Param("id")

	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.Response{
			Success: false,
			Message: "Неверный формат параметра 'id'",
			Error:   "INVALID_ID",
		})
		return
	}

	var exists bool
	err = database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", userID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "Ошибка базы данных: " + err.Error(),
			Error:   "DATABASE_ERROR",
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, types.Response{
			Success: false,
			Message: "Пользователь не найден",
			Error:   "USER_NOT_FOUND",
		})
		return
	}

	result, err := database.DB.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "Не удалось удалить пользователя: " + err.Error(),
			Error:   "USER_DELETE_ERROR",
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "Ошибка при проверке результата",
			Error:   "RESULT_CHECK_ERROR",
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, types.Response{
			Success: false,
			Message: "Пользователь не был удален",
			Error:   "USER_NOT_DELETED",
		})
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Success: true,
		Message: "Пользователь успешно удален",
		Data: map[string]interface{}{
			"deleted_id": userID,
		},
	})
}
