package handlers

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"library-app/models"

	"github.com/gin-gonic/gin"
)

type BookHandler struct {
	DB *sql.DB
}

func (h *BookHandler) GetBooks(c *gin.Context) {
	rows, err := h.DB.Query(`
		SELECT 
			b.id, 
			b.title, 
			b.isbn, 
			b.publication_year,
			COALESCE((
				SELECT STRING_AGG(a.first_name || ' ' || a.last_name, ', ')
				FROM book_authors ba
				JOIN authors a ON ba.author_id = a.id
				WHERE ba.book_id = b.id
			), '') AS authors,
			COALESCE((
				SELECT STRING_AGG(g.name, ', ')
				FROM book_genres bg
				JOIN genres g ON bg.genre_id = g.id
				WHERE bg.book_id = b.id
			), '') AS genres
		FROM books b
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	books := []models.BookWithAuthors{}
	for rows.Next() {
		var b models.BookWithAuthors
		var year sql.NullInt64 // Используем NullInt64, так как год может быть NULL в БД
		if err := rows.Scan(&b.ID, &b.Title, &b.ISBN, &year, &b.Authors, &b.Genres); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if year.Valid {
			b.Year = int(year.Int64)
		}
		books = append(books, b)
	}
	c.JSON(http.StatusOK, books)
}

// Запрос 1: Поиск книги по названию, автору или жанру
func (h *BookHandler) SearchBooks(c *gin.Context) {
	query := c.Query("q")
	searchPattern := "%" + query + "%"

	rows, err := h.DB.Query(`
		SELECT 
			b.id, 
			b.title, 
			b.isbn, 
			b.publication_year,
			COALESCE((
				SELECT STRING_AGG(a.first_name || ' ' || a.last_name, ', ')
				FROM book_authors ba
				JOIN authors a ON ba.author_id = a.id
				WHERE ba.book_id = b.id
			), '') AS authors,
			COALESCE((
				SELECT STRING_AGG(g.name, ', ')
				FROM book_genres bg
				JOIN genres g ON bg.genre_id = g.id
				WHERE bg.book_id = b.id
			), '') AS genres
		FROM books b
		WHERE b.title ILIKE $1 
		   OR EXISTS (
			   SELECT 1 
			   FROM book_authors ba
			   JOIN authors a ON ba.author_id = a.id
			   WHERE ba.book_id = b.id 
				 AND (a.last_name ILIKE $1 OR a.first_name ILIKE $1)
		   )
		   OR EXISTS (
			   SELECT 1 
			   FROM book_genres bg
			   JOIN genres g ON bg.genre_id = g.id
			   WHERE bg.book_id = b.id 
				 AND g.name ILIKE $1
		   )
	`, searchPattern)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var books []models.BookWithAuthors
	for rows.Next() {
		var b models.BookWithAuthors
		var year sql.NullInt64
		if err := rows.Scan(&b.ID, &b.Title, &b.ISBN, &year, &b.Authors, &b.Genres); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if year.Valid {
			b.Year = int(year.Int64)
		}
		books = append(books, b)
	}
	if books == nil {
		books = []models.BookWithAuthors{}
	}
	c.JSON(http.StatusOK, books)
}

// Запрос 2: Детальный статус книги (Страница книги)
func (h *BookHandler) GetBookStatus(c *gin.Context) {
	bookID := c.Param("id")
	var status models.BookAvailability

	err := h.DB.QueryRow(`
		SELECT 
			b.id, 
			b.title,
			NOT EXISTS (
				SELECT 1 
				FROM loans l 
				WHERE l.book_id = b.id 
				  AND l.return_date IS NULL
			) AS is_available
		FROM books b
		WHERE b.id = $1
	`, bookID).Scan(&status.ID, &status.Title, &status.IsAvailable)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Книга не найдена"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

// Запрос 3: Текущие книги на руках (Личный кабинет)
func (h *BookHandler) GetCurrentLoans(c *gin.Context) {
	readerID := c.Param("id")
	rows, err := h.DB.Query(`
		SELECT 
			l.id AS loan_id, 
			b.title, 
			l.loan_date,
			CURRENT_DATE - l.loan_date AS days_on_loan
		FROM loans l
		JOIN books b ON l.book_id = b.id
		WHERE l.reader_id = $1
		  AND l.return_date IS NULL
		ORDER BY days_on_loan DESC
	`, readerID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var loans []models.LoanInfo
	for rows.Next() {
		var l models.LoanInfo
		if err := rows.Scan(&l.LoanID, &l.Title, &l.LoanDate, &l.DaysOnLoan); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		loans = append(loans, l)
	}
	if loans == nil {
		loans = []models.LoanInfo{}
	}
	c.JSON(http.StatusOK, loans)
}

// Запрос 4: История прочитанного (Личный кабинет)
func (h *BookHandler) GetReadingHistory(c *gin.Context) {
	readerID := c.Param("id")
	rows, err := h.DB.Query(`
		SELECT 
			l.id AS loan_id, 
			b.title, 
			l.loan_date, 
			l.return_date
		FROM loans l
		JOIN books b ON l.book_id = b.id
		WHERE l.reader_id = $1
		  AND l.return_date IS NOT NULL
		ORDER BY l.return_date DESC
	`, readerID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var history []models.HistoryInfo
	for rows.Next() {
		var hi models.HistoryInfo
		if err := rows.Scan(&hi.LoanID, &hi.Title, &hi.LoanDate, &hi.ReturnDate); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		history = append(history, hi)
	}
	if history == nil {
		history = []models.HistoryInfo{}
	}
	c.JSON(http.StatusOK, history)
}

// Запрос 5: Список должников (Панель библиотекаря)
func (h *BookHandler) GetDebtors(c *gin.Context) {
	rows, err := h.DB.Query(`
		SELECT 
			r.id AS reader_id, 
			r.full_name, 
			r.email, 
			r.phone, 
			b.title AS book_title, 
			l.loan_date,
			CURRENT_DATE - l.loan_date AS overdue_days
		FROM loans l
		JOIN readers r ON l.reader_id = r.id
		JOIN books b ON l.book_id = b.id
		WHERE l.return_date IS NULL 
		  AND (CURRENT_DATE - l.loan_date) > 14
		ORDER BY overdue_days DESC
	`)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var debtors []gin.H
	for rows.Next() {
		var d models.DebtorInfo
		var phone sql.NullString // Телефон может быть не указан (NULL)
		if err := rows.Scan(&d.ReaderID, &d.FullName, &d.Email, &phone, &d.BookTitle, &d.LoanDate, &d.OverdueDays); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if phone.Valid {
			d.Phone = phone.String
		}
		debtors = append(debtors, gin.H{
			"reader_id":    d.ReaderID,
			"full_name":    d.FullName,
			"email":        d.Email,
			"phone":        d.Phone,
			"book_title":   d.BookTitle,
			"loan_date":    d.LoanDate,
			"overdue_days": d.OverdueDays,
		})
	}
	if debtors == nil {
		debtors = []gin.H{}
	}
	c.JSON(http.StatusOK, debtors)
}

// Запрос 6: Топ самых популярных книг (Статистика)
func (h *BookHandler) GetPopularBooks(c *gin.Context) {
	rows, err := h.DB.Query(`
		SELECT 
			b.id, 
			b.title, 
			COUNT(l.id) AS borrow_count
		FROM books b
		JOIN loans l ON b.id = l.book_id
		GROUP BY b.id, b.title
		ORDER BY borrow_count DESC
		LIMIT 3
	`)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var popular []gin.H
	for rows.Next() {
		var p models.PopularBook
		if err := rows.Scan(&p.ID, &p.Title, &p.BorrowCount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		popular = append(popular, gin.H{
			"id":           p.ID,
			"title":        p.Title,
			"borrow_count": p.BorrowCount,
		})
	}
	if popular == nil {
		popular = []gin.H{}
	}
	c.JSON(http.StatusOK, popular)
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	// Жестко заданные тестовые пользователи
	if req.Username == "admin" && req.Password == "admin123" {
		c.JSON(http.StatusOK, gin.H{"token": "admin-token", "role": "admin"})
		return
	}
	if req.Username == "reader" && req.Password == "reader123" {
		c.JSON(http.StatusOK, gin.H{"token": "reader-token", "role": "reader"})
		return
	}

	c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин или пароль"})
}

// AuthMiddleware проверяет токен и права доступа пользователя
func AuthMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		role := ""

		if token == "Bearer admin-token" {
			role = "admin"
		} else if token == "Bearer reader-token" {
			role = "reader"
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
			c.Abort()
			return
		}

		// Администратор имеет доступ ко всему, читатель - только к своему
		if requiredRole != "" && role != "admin" && role != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
			c.Abort()
			return
		}

		c.Set("role", role)
		c.Next()
	}
}

// Запрос: История выдач конкретной книги (Панель библиотекаря)
func (h *BookHandler) GetBookHistory(c *gin.Context) {
	bookID := c.Param("id")
	rows, err := h.DB.Query(`
		SELECT l.id, r.full_name, l.loan_date, l.return_date
		FROM loans l
		JOIN readers r ON l.reader_id = r.id
		WHERE l.book_id = $1
		ORDER BY l.loan_date DESC
	`, bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var history []models.BookHistory
	for rows.Next() {
		var hi models.BookHistory
		var retDate sql.NullTime
		if err := rows.Scan(&hi.LoanID, &hi.ReaderName, &hi.LoanDate, &retDate); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if retDate.Valid {
			hi.ReturnDate = &retDate.Time
		}
		history = append(history, hi)
	}
	if history == nil {
		history = []models.BookHistory{}
	}
	c.JSON(http.StatusOK, history)
}

// Запрос: Экспорт истории книги в CSV
func (h *BookHandler) ExportBookHistory(c *gin.Context) {
	bookID := c.Param("id")
	rows, err := h.DB.Query(`
		SELECT l.id, r.full_name, l.loan_date, l.return_date
		FROM loans l
		JOIN readers r ON l.reader_id = r.id
		WHERE l.book_id = $1
		ORDER BY l.loan_date DESC
	`, bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	c.Writer.Header().Set("Content-Type", "text/csv; charset=utf-8")
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=book_%s_history.csv", bookID))

	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF}) // Добавляем BOM, чтобы Excel корректно читал кириллицу
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	writer.Write([]string{"ID выдачи", "Читатель", "Дата выдачи", "Дата возврата"})

	for rows.Next() {
		var loanID int
		var readerName string
		var loanDate time.Time
		var returnDate *time.Time
		if err := rows.Scan(&loanID, &readerName, &loanDate, &returnDate); err != nil {
			return
		}

		retDateStr := "На руках"
		if returnDate != nil {
			retDateStr = returnDate.Format("2006-01-02")
		}

		writer.Write([]string{fmt.Sprintf("%d", loanID), readerName, loanDate.Format("2006-01-02"), retDateStr})
	}
}

// Запрос: Выдача книги (Админ-панель)
type IssueRequest struct {
	ReaderID int `json:"reader_id"`
	BookID   int `json:"book_id"`
}

func (h *BookHandler) IssueBook(c *gin.Context) {
	var req IssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	// Проверяем, не выдана ли уже эта книга
	var isAvailable bool
	err := h.DB.QueryRow(`
		SELECT NOT EXISTS (
			SELECT 1 FROM loans WHERE book_id = $1 AND return_date IS NULL
		)
	`, req.BookID).Scan(&isAvailable)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка проверки статуса книги"})
		return
	}

	if !isAvailable {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Книга уже выдана"})
		return
	}

	// Оформляем выдачу
	_, err = h.DB.Exec(`
		INSERT INTO loans (book_id, reader_id, loan_date) 
		VALUES ($1, $2, CURRENT_DATE)
	`, req.BookID, req.ReaderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось оформить выдачу"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Книга успешно выдана"})
}

// Запрос: Возврат книги (Админ-панель)
type ReturnRequest struct {
	BookID int `json:"book_id"`
}

func (h *BookHandler) ReturnBook(c *gin.Context) {
	var req ReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	// Оформляем возврат путем проставления текущей даты
	res, err := h.DB.Exec(`
		UPDATE loans 
		SET return_date = CURRENT_DATE 
		WHERE book_id = $1 AND return_date IS NULL
	`, req.BookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных"})
		return
	}

	// Проверяем, обновилась ли запись (была ли книга вообще на руках)
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Книга не находится на руках"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Возврат успешно оформлен"})
}

// Запрос: Список читателей (Админ-панель)
func (h *BookHandler) GetReaders(c *gin.Context) {
	rows, err := h.DB.Query(`SELECT id, full_name, email, phone FROM readers ORDER BY id`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var readers []models.Reader
	for rows.Next() {
		var r models.Reader
		var phone sql.NullString
		if err := rows.Scan(&r.ID, &r.FullName, &r.Email, &phone); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if phone.Valid {
			r.Phone = phone.String
		}
		readers = append(readers, r)
	}
	if readers == nil {
		readers = []models.Reader{}
	}
	c.JSON(http.StatusOK, readers)
}

// Запрос: Добавление книги (Админ)
type AddBookRequest struct {
	Title           string `json:"title" binding:"required"`
	ISBN            string `json:"isbn" binding:"required"`
	PublicationYear int    `json:"publication_year"`
}

func (h *BookHandler) AddBook(c *gin.Context) {
	var req AddBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	var newID int
	err := h.DB.QueryRow(`
		INSERT INTO books (title, isbn, publication_year)
		VALUES ($1, $2, $3) RETURNING id
	`, req.Title, req.ISBN, req.PublicationYear).Scan(&newID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Книга успешно добавлена", "id": newID})
}

// Запрос: Удаление книги (Админ)
func (h *BookHandler) DeleteBook(c *gin.Context) {
	id := c.Param("id")
	res, err := h.DB.Exec(`DELETE FROM books WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении: " + err.Error()})
		return
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Книга не найдена"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Книга успешно удалена"})
}

// Запрос: Добавление читателя (Админ)
type AddReaderRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Phone    string `json:"phone"`
}

func (h *BookHandler) AddReader(c *gin.Context) {
	var req AddReaderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	var newID int
	err := h.DB.QueryRow(`
		INSERT INTO readers (full_name, email, phone)
		VALUES ($1, $2, $3) RETURNING id
	`, req.FullName, req.Email, req.Phone).Scan(&newID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Email должен быть уникальным"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Читатель успешно добавлен", "id": newID})
}

// Запрос: Удаление читателя (Админ)
func (h *BookHandler) DeleteReader(c *gin.Context) {
	id := c.Param("id")
	res, err := h.DB.Exec(`DELETE FROM readers WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении"})
		return
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Читатель не найден"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Читатель успешно удален"})
}
