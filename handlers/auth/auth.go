// Package auth предоставляет функции для регистрации пользователей
// в банковской системе.
package auth

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"backend_golang/database"
	"backend_golang/methods"
	"backend_golang/types"
)

func Register(c *gin.Context) {
	name := c.PostForm("name")
	surname := c.PostForm("surname")
	phoneNumber := c.PostForm("phone_number")
	password := c.PostForm("password")
	balanceStr := c.PostForm("balance")

	if name == "" || surname == "" || phoneNumber == "" || password == "" {
		c.JSON(http.StatusBadRequest, types.Response{
			Success: false,
			Message: "Имя, фамилия, телефон и пароль обязательны",
			Error:   "MISSING_FIELDS",
		})
		return
	}

	var existingID int64
	err := database.DB.QueryRow(
		"SELECT id FROM users WHERE phone_number = ?",
		phoneNumber,
	).Scan(&existingID)

	if err == nil {
		c.JSON(http.StatusConflict, types.Response{
			Success: false,
			Message: "Пользователь с таким телефоном уже существует",
			Error:   "USER_EXISTS",
		})
		return
	} else if err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "Ошибка при проверке пользователя: " + err.Error(),
			Error:   "DATABASE_ERROR",
		})
		return
	}

	var balance float64
	if balanceStr != "" {
		if bal, err := strconv.ParseFloat(balanceStr, 64); err == nil && bal >= 0 {
			balance = bal
		} else {
			c.JSON(http.StatusBadRequest, types.Response{
				Success: false,
				Message: "Неверный формат баланса",
				Error:   "INVALID_BALANCE",
			})
			return
		}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), types.BcryptSalt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "Ошибка при хешировании пароля: " + err.Error(),
			Error:   "HASH_ERROR",
		})
		return
	}

	session := methods.GenerateSecureSession(types.DefaultSession)

	query := `
        INSERT INTO users 
        (name, surname, phone_number, balance, password_hash, session) 
        VALUES (?, ?, ?, ?, ?, ?)
    `

	result, err := database.DB.Exec(
		query,
		name,
		surname,
		phoneNumber,
		balance,
		string(passwordHash),
		session,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "Ошибка при регистрации пользователя: " + err.Error(),
			Error:   "REGISTRATION_ERROR",
		})
		return
	}

	userID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "Ошибка при получении ID пользователя: " + err.Error(),
			Error:   "ID_ERROR",
		})
		return
	}

	c.JSON(http.StatusCreated, types.Response{
		Success: true,
		Message: "Пользователь успешно зарегистрирован",
		Data: map[string]interface{}{
			"user_id": userID,
			"session": session,
			"balance": balance,
		},
	})
}

func Login(c *gin.Context) {
	phoneNumber := c.PostForm("phone_number")
	password := c.PostForm("password")

	if phoneNumber == "" || password == "" {
		c.JSON(http.StatusBadRequest, types.Response{
			Success: false,
			Message: "Телефон и пароль обязательны",
			Error:   "MISSING_FIELDS",
		})
		return
	}

	var passwordHash string
	err := database.DB.QueryRow(
		"SELECT password_hash FROM users WHERE phone_number = ?",
		phoneNumber,
	).Scan(&passwordHash)

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

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			c.JSON(http.StatusBadRequest, types.Response{
				Success: false,
				Message: "Неправильный пароль",
				Error:   "INVALID_PASSWORD",
			})
		} else {
			c.JSON(http.StatusInternalServerError, types.Response{
				Success: false,
				Message: "Ошибка проверки пароля: " + err.Error(),
				Error:   "PASSWORD_CHECK_ERROR",
			})
		}
		return
	}

	var userID int64
	var balance float64
	var name, surname string
	var sessionNullable sql.NullString 

	err = database.DB.QueryRow(
		"SELECT id, name, surname, balance, session FROM users WHERE phone_number = ?",
		phoneNumber,
	).Scan(&userID, &name, &surname, &balance, &sessionNullable)

	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "Ошибка при получении данных пользователя: " + err.Error(),
			Error:   "USER_DATA_ERROR",
		})
		return
	}

	var session string
	if sessionNullable.Valid {
		session = sessionNullable.String
	} else {
		session = methods.GenerateSecureSession(types.DefaultSession)
		_, err = database.DB.Exec(
			"UPDATE users SET session = ? WHERE id = ?",
			session, userID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, types.Response{
				Success: false,
				Message: "Ошибка при создании сессии: " + err.Error(),
				Error:   "SESSION_CREATE_ERROR",
			})
			return
		}
	}

	c.JSON(http.StatusOK, types.Response{
		Success: true,
		Message: "Успешный вход в систему",
		Data: map[string]interface{}{
			"user_id":      userID,
			"name":         name,
			"surname":      surname,
			"phone_number": phoneNumber,
			"balance":      balance,
			"session":      session,
		},
	})
}

func Logout(c *gin.Context) {
	session := c.Query("session")
	if session == "" {
		session = c.GetHeader("Session")
	}

	if session == "" {
		c.JSON(http.StatusBadRequest, types.Response{
			Success: false,
			Message: "Сессия обязательна",
			Error:   "MISSING_FIELDS",
		})
		return
	}

	result, err := database.DB.Exec("UPDATE users SET session = NULL WHERE session = ?", session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "Не удалось удалить сессию",
			Error:   "SESSION_DELETE_ERROR",
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
			Message: "Пользователь с такой сессией не найден",
			Error:   "USER_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Success: true,
		Message: "Успешный выход из системы",
	})
}

func RefreshSession(c *gin.Context) {
	session := c.PostForm("session")
	if session == "" {
		var req struct {
			Session string `json:"session"`
		}
		if err := c.ShouldBindJSON(&req); err == nil {
			session = req.Session
		}
	}

	if session == "" {
		c.JSON(http.StatusBadRequest, types.Response{
			Success: false,
			Message: "Сессия обязательна",
			Error:   "MISSING_FIELDS",
		})
		return
	}

	newSession := methods.GenerateSecureSession(types.DefaultSession)

	result, err := database.DB.Exec("UPDATE users SET session = ? WHERE session = ?", newSession, session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "Не удалось изменить сессию",
			Error:   "SESSION_UPDATE_ERROR",
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
			Message: "Пользователь с такой сессией не найден",
			Error:   "USER_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Success: true,
		Message: "Успешное обновление сессии",
		Data: map[string]interface{}{
			"session": newSession,
		},
	})
}

func RefreshPassword(c *gin.Context) {
	userID := c.PostForm("id")
	password := c.PostForm("new_password")

	if userID == "" || password == "" {
		var req struct {
			ID       string `json:"id"`
			Password string `json:"new_password"`
		}
		if err := c.ShouldBindJSON(&req); err == nil {
			userID = req.ID
			password = req.Password
		}
	}

	if userID == "" || password == "" {
		c.JSON(http.StatusBadRequest, types.Response{
			Success: false,
			Message: "Пароль и айди обязательны",
			Error:   "MISSING_FIELDS",
		})
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), types.BcryptSalt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "Ошибка при хешировании пароля",
			Error:   "HASH_ERROR",
		})
		return
	}

	result, err := database.DB.Exec("UPDATE users SET password_hash = ? WHERE id = ?", string(passwordHash), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "Не удалось сменить пароль",
			Error:   "PASSWORD_UPDATE_ERROR",
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
			Message: "Пользователь с таким айди не найден",
			Error:   "USER_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Success: true,
		Message: "Пароль успешно изменен",
	})
}

func GetBySession(c *gin.Context) {
	session := c.Query("session")

	if session == "" {
		c.JSON(http.StatusBadRequest, types.Response{
			Success: false,
			Message: "Сессия обязательна",
			Error:   "MISSING_FIELDS",
		})
		return
	}

	rows, err := database.DB.Query("SELECT * FROM users WHERE session = ?", session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "Ошибка базы данных: " + err.Error(),
			Error:   "DATABASE_ERROR",
		})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		c.JSON(http.StatusNotFound, types.Response{
			Success: false,
			Message: "Пользователь с такой сессией не найден",
			Error:   "USER_NOT_FOUND",
		})
		return
	}

	columns, err := rows.Columns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "Ошибка получения колонок: " + err.Error(),
			Error:   "COLUMNS_ERROR",
		})
		return
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	err = rows.Scan(valuePtrs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "Ошибка чтения данных: " + err.Error(),
			Error:   "SCAN_ERROR",
		})
		return
	}

	userData := make(map[string]interface{})
	for i, col := range columns {
		var v interface{}
		val := values[i]
		b, ok := val.([]byte)
		if ok {
			v = string(b)
		} else {
			v = val
		}
		userData[col] = v
	}

	c.JSON(http.StatusOK, types.Response{
		Success: true,
		Message: "Пользователь найден по сессии",
		Data:    userData,
	})
}
