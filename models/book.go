package models

import "time"

type Book struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	ISBN  string `json:"isbn"`
	Year  int    `json:"year"`
}

type BookWithAuthors struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	ISBN    string `json:"isbn"`
	Year    int    `json:"year"`
	Authors string `json:"authors"`
	Genres  string `json:"genres"`
}

type BookAvailability struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	IsAvailable bool   `json:"is_available"`
}

type LoanInfo struct {
	LoanID     int       `json:"loan_id"`
	Title      string    `json:"title"`
	LoanDate   time.Time `json:"loan_date"`
	DaysOnLoan int       `json:"days_on_loan"`
}

type BookHistory struct {
	LoanID     int        `json:"loan_id"`
	ReaderName string     `json:"reader_name"`
	LoanDate   time.Time  `json:"loan_date"`
	ReturnDate *time.Time `json:"return_date"`
}

type HistoryInfo struct {
	LoanID     int       `json:"loan_id"`
	Title      string    `json:"title"`
	LoanDate   time.Time `json:"loan_date"`
	ReturnDate time.Time `json:"return_date"`
}

type DebtorInfo struct {
	ReaderID    int       `json:"reader_id"`
	FullName    string    `json:"full_name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	BookTitle   string    `json:"book_title"`
	LoanDate    time.Time `json:"loan_date"`
	OverdueDays int       `json:"overdue_days"`
}

type PopularBook struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	BorrowCount int    `json:"borrow_count"`
}

type Reader struct {
	ID       int    `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}
