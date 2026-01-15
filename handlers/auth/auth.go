// Package auth предоставляет функции для регистрации пользователей
// в банковской системе.
package auth

import (
	"database/sql"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	DataBase "backend_golang/dataBase"
	"backend_golang/methods"
	"backend_golang/types"
)

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methods.SendJSONError(w, "Только POST запросы", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()

	name := r.FormValue("name")
	surname := r.FormValue("surname")
	phoneNumber := r.FormValue("phone_number")
	password := r.FormValue("password")
	balanceStr := r.FormValue("balance")

	if name == "" || surname == "" || phoneNumber == "" || password == "" {
		methods.SendJSONError(w, "Имя, фамилия, телефон и пароль обязательны", http.StatusBadRequest)
		return
	}

	var existingID int64
	err := DataBase.DB.QueryRow(
		"SELECT id FROM users WHERE phone_number = ?",
		phoneNumber,
	).Scan(&existingID)

	if err == nil {
		methods.SendJSONError(w, "Пользователь с таким телефоном уже существует", http.StatusConflict)
		return
	} else if err != sql.ErrNoRows {
		methods.SendJSONError(w, "Ошибка при проверке пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var balance float64
	if balanceStr != "" {
		if bal, err := strconv.ParseFloat(balanceStr, 64); err == nil && bal >= 0 {
			balance = bal
		} else {
			methods.SendJSONError(w, "Неверный формат баланса", http.StatusBadRequest)
			return
		}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), types.BcryptSalt)
	if err != nil {
		methods.SendJSONError(w, "Ошибка при хешировании пароля: "+err.Error(), http.StatusInternalServerError)
		return
	}

	session := methods.GenerateSecureSession(types.DefaultSession)

	query := `
        INSERT INTO users 
        (name, surname, phone_number, balance, password_hash, session) 
        VALUES (?, ?, ?, ?, ?, ?)
    `

	result, err := DataBase.DB.Exec(
		query,
		name,
		surname,
		phoneNumber,
		balance,
		string(passwordHash),
		session,
	)

	if err != nil {
		methods.SendJSONError(w, "Ошибка при регистрации пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	userID, err := result.LastInsertId()
	if err != nil {
		methods.SendJSONError(w, "Ошибка при получении ID пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	methods.SendJSONResponse(w, types.ResponseForAuth{
		Success: true,
		Message: "Пользователь успешно зарегистрирован",
		UserID:  userID,
		Session: session,
		Balance: balance,
	}, http.StatusCreated)
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methods.SendJSONErrorResponse(w, "Только POST запросы", "METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()

	phoneNumber := r.FormValue("phone_number")
	password := r.FormValue("password")

	if phoneNumber == "" || password == "" {
		methods.SendJSONErrorResponse(w, "Телефон и пароль обязательны", "MISSING_FIELDS", http.StatusBadRequest)
		return
	}

	var passwordHash string
	err := DataBase.DB.QueryRow(
		"SELECT password_hash FROM users WHERE phone_number = ?",
		phoneNumber,
	).Scan(&passwordHash)

	if err != nil {
		if err == sql.ErrNoRows {
			methods.SendJSONErrorResponse(w, "Пользователь не найден", "USER_NOT_FOUND", http.StatusNotFound)
		} else {
			methods.SendJSONErrorResponse(w, "Ошибка базы данных: "+err.Error(), "DATABASE_ERROR", http.StatusInternalServerError)
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			methods.SendJSONErrorResponse(w, "Неправильный пароль", "INVALID_PASSWORD", http.StatusBadRequest)
		} else {
			methods.SendJSONErrorResponse(w, "Ошибка проверки пароля: "+err.Error(), "PASSWORD_CHECK_ERROR", http.StatusInternalServerError)
		}
		return
	}

	var userID int64
	var balance float64
	var name, surname, session string

	err = DataBase.DB.QueryRow(
		"SELECT id, name, surname, balance, session FROM users WHERE phone_number = ?",
		phoneNumber,
	).Scan(&userID, &name, &surname, &balance, &session)

	if err != nil {
		methods.SendJSONErrorResponse(w, "Ошибка при получении данных пользователя: "+err.Error(), "USER_DATA_ERROR", http.StatusInternalServerError)
		return
	}

	if session == "" {
		session = methods.GenerateSecureSession(32)
		_, err = DataBase.DB.Exec(
			"UPDATE users SET session = ? WHERE id = ?",
			session, userID,
		)
		if err != nil {
			methods.SendJSONErrorResponse(w, "Ошибка при создании сессии: "+err.Error(), "SESSION_CREATE_ERROR", http.StatusInternalServerError)
			return
		}
	}

	userData := map[string]interface{}{
		"user_id":      userID,
		"name":         name,
		"surname":      surname,
		"phone_number": phoneNumber,
		"balance":      balance,
		"session":      session,
	}

	methods.SendJSONResponseGeneric(w, types.Response{
		Success: true,
		Message: "Успешный вход в систему",
		Data:    userData,
	}, http.StatusOK)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methods.SendJSONErrorResponse(w, "Только POST запросы", "METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()

	session := r.FormValue("session")

	if session == "" {
		methods.SendJSONErrorResponse(w, "Сессия обязательна", "MISSING_FIELDS", http.StatusBadRequest)
		return
	}

	result, err := DataBase.DB.Exec("UPDATE users SET session = NULL WHERE session = ?", session)
	if err != nil {
		methods.SendJSONErrorResponse(w, "Не удалось удалить сессию", "SESSION_DELETE_ERROR", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		methods.SendJSONErrorResponse(w, "Ошибка при проверке результата", "RESULT_CHECK_ERROR", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		methods.SendJSONErrorResponse(w, "Пользователь с такой сессией не найден", "USER_NOT_FOUND", http.StatusNotFound)
		return
	}

	methods.SendJSONSuccessResponse(w, "Успешный выход из системы", http.StatusOK)
}

func RefreshSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methods.SendJSONErrorResponse(w, "Только POST запросы", "METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()

	session := r.FormValue("session")

	if session == "" {
		methods.SendJSONErrorResponse(w, "Сессия обязательна", "MISSING_FIELDS", http.StatusBadRequest)
		return
	}

	newSession := methods.GenerateSecureSession(types.DefaultSession)

	result, err := DataBase.DB.Exec("UPDATE users SET session = ? WHERE session = ?", newSession, session)
	if err != nil {
		methods.SendJSONErrorResponse(w, "Не удалось изменить сессию", "SESSION_UPDATE_ERROR", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		methods.SendJSONErrorResponse(w, "Ошибка при проверке результата", "RESULT_CHECK_ERROR", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		methods.SendJSONErrorResponse(w, "Пользователь с такой сессией не найден", "USER_NOT_FOUND", http.StatusNotFound)
		return
	}

	methods.SendJSONSuccessResponseWithData(w, "Успешное обновление сессии", map[string]interface{}{
		"session": newSession,
	}, http.StatusOK)
}

func RefreshPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methods.SendJSONErrorResponse(w, "Только POST запросы", "METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()

	userID := r.FormValue("id")
	password := r.FormValue("new_password")

	if userID == "" || password == "" {
		methods.SendJSONErrorResponse(w, "Пароль и айди обязательны", "MISSING_FIELDS", http.StatusBadRequest)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), types.BcryptSalt)
	if err != nil {
		methods.SendJSONErrorResponse(w, "Ошибка при хешировании пароля", "HASH_ERROR", http.StatusInternalServerError)
		return
	}

	result, err := DataBase.DB.Exec("UPDATE users SET password_hash = ? WHERE id = ?", string(passwordHash), userID)
	if err != nil {
		methods.SendJSONErrorResponse(w, "Не удалось сменить пароль", "PASSWORD_UPDATE_ERROR", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		methods.SendJSONErrorResponse(w, "Ошибка при проверке результата", "RESULT_CHECK_ERROR", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		methods.SendJSONErrorResponse(w, "Пользователь с таким айди не найден", "USER_NOT_FOUND", http.StatusNotFound)
		return
	}

	methods.SendJSONSuccessResponse(w, "Пароль успешно изменен", http.StatusOK)
}

func GetBySession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methods.SendJSONErrorResponse(w, "Только POST запросы", "METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()

	session := r.FormValue("session")

	if session == "" {
		methods.SendJSONErrorResponse(w, "Сессия обязательна", "MISSING_FIELDS", http.StatusBadRequest)
		return
	}

	rows, err := DataBase.DB.Query("SELECT * FROM users WHERE session = ?", session)
	if err != nil {
		methods.SendJSONErrorResponse(w, "Ошибка базы данных: "+err.Error(), "DATABASE_ERROR", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if !rows.Next() {
		methods.SendJSONErrorResponse(w, "Пользователь с такой сессией не найден", "USER_NOT_FOUND", http.StatusNotFound)
		return
	}

	columns, err := rows.Columns()
	if err != nil {
		methods.SendJSONErrorResponse(w, "Ошибка получения колонок: "+err.Error(), "COLUMNS_ERROR", http.StatusInternalServerError)
		return
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	err = rows.Scan(valuePtrs...)
	if err != nil {
		methods.SendJSONErrorResponse(w, "Ошибка чтения данных: "+err.Error(), "SCAN_ERROR", http.StatusInternalServerError)
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

	methods.SendJSONSuccessResponseWithData(w, "Пользователь найден по сессии", userData, http.StatusOK)
}