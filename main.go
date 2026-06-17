package main

import (
	"log"
	"net/http"

	"library-app/database"
	"library-app/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Ошибка подключения к БД: ", err)
	}
	defer db.Close()

	bookHandler := &handlers.BookHandler{DB: db}

	r := gin.Default()

	api := r.Group("/api")
	api.POST("/login", handlers.Login)

	api.GET("/books", bookHandler.GetBooks)
	api.GET("/books/search", bookHandler.SearchBooks)
	api.GET("/books/:id/status", bookHandler.GetBookStatus)

	readerGroup := api.Group("/readers")
	readerGroup.Use(handlers.AuthMiddleware("reader"))
	readerGroup.GET("/:id/loans", bookHandler.GetCurrentLoans)
	readerGroup.GET("/:id/history", bookHandler.GetReadingHistory)

	adminGroup := api.Group("/admin")
	adminGroup.Use(handlers.AuthMiddleware("admin"))
	adminGroup.GET("/debtors", bookHandler.GetDebtors)
	adminGroup.GET("/popular-books", bookHandler.GetPopularBooks)
	adminGroup.GET("/books/:id/history", bookHandler.GetBookHistory)
	adminGroup.GET("/books/:id/export-history", bookHandler.ExportBookHistory)
	adminGroup.GET("/readers", bookHandler.GetReaders)
	adminGroup.POST("/issue", bookHandler.IssueBook)
	adminGroup.POST("/return", bookHandler.ReturnBook)
	adminGroup.POST("/books", bookHandler.AddBook)
	adminGroup.DELETE("/books/:id", bookHandler.DeleteBook)
	adminGroup.POST("/readers", bookHandler.AddReader)
	adminGroup.DELETE("/readers/:id", bookHandler.DeleteReader)

	r.NoRoute(gin.WrapH(http.FileServer(http.Dir("./static"))))

	log.Println("Сервер успешно запущен на http://localhost:8080")
	log.Fatal(r.Run(":8080"))
}
