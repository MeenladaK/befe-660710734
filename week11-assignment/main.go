package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "week11-assignment/docs"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type Book struct {
	ID            int      `json:"id"`
	Title         string   `json:"title"`
	Author        string   `json:"author"`
	ISBN          string   `json:"isbn"`
	Year          int      `json:"year"`
	Price         float64  `json:"price"`
	Category      string   `json:"category"`
	OriginalPrice *float64 `json:"original_price,omitempty"`
	Discount      int      `json:"discount"`
	CoverImage    string   `json:"cover_image"`
	Rating        float64  `json:"rating"`
	ReviewsCount  int      `json:"reviews_count"`
	IsNew         bool     `json:"is_new"`
	Pages         *int     `json:"pages,omitempty"`
	Language      string   `json:"language"`
	Publisher     string   `json:"publisher"`
	Description   string   `json:"description"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var db *sql.DB

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func initDB() {
	var err error

	host := getEnv("DB_HOST", "")
	name := getEnv("DB_NAME", "")
	user := getEnv("DB_USER", "")
	password := getEnv("DB_PASSWORD", "")
	port := getEnv("DB_PORT", "")

	conStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, name)

	db, err = sql.Open("postgres", conStr)
	if err != nil {
		log.Fatal("failed to open DB:", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		log.Fatal("failed to connect DB:", err)
	}
	log.Println("DB connected")
}

func scanBooks(rows *sql.Rows) ([]Book, error) {
	var books []Book
	for rows.Next() {
		var b Book
		err := rows.Scan(
			&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Year, &b.Price,
			&b.Category, &b.OriginalPrice, &b.Discount, &b.CoverImage,
			&b.Rating, &b.ReviewsCount, &b.IsNew, &b.Pages,
			&b.Language, &b.Publisher, &b.Description,
			&b.CreatedAt, &b.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, nil
}

// ใช้ COALESCE สำหรับคอลัมน์ที่อาจ NULL เพื่อให้ scan เข้า type ปกติได้
const selectCols = `
SELECT
	id, title, author, isbn, year, price,
	COALESCE(category, '')              AS category,
	original_price,                     -- pointer -> ให้ NULL ได้
	COALESCE(discount, 0)               AS discount,
	COALESCE(cover_image, '')           AS cover_image,
	COALESCE(rating, 0.0)               AS rating,
	COALESCE(reviews_count, 0)          AS reviews_count,
	COALESCE(is_new, false)             AS is_new,
	pages,                              -- pointer -> ให้ NULL ได้
	COALESCE(language, '')              AS language,
	COALESCE(publisher, '')             AS publisher,
	COALESCE(description, '')           AS description,
	created_at, updated_at
FROM books
`

// @Summary Health check
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func getHealth(c *gin.Context) {
	if err := db.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "unhealthy", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "healthy"})
}

// @Summary Get all books (optional: category filter)
// @Description Get all books, optionally filter by ?category=xxx
// @Tags Books
// @Produce json
// @Param category query string false "Filter by category"
// @Success 200 {array} Book
// @Failure 500 {object} ErrorResponse
// @Router /books [get]
func getAllBooks(c *gin.Context) {
	category := c.Query("category")

	var (
		rows *sql.Rows
		err  error
	)

	if category != "" {
		rows, err = db.Query(selectCols+` WHERE category = $1 ORDER BY id ASC`, category)
	} else {
		rows, err = db.Query(selectCols + ` ORDER BY id ASC`)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	books, err := scanBooks(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "scan error: " + err.Error()})
		return
	}
	if books == nil {
		books = []Book{}
	}
	c.JSON(http.StatusOK, books)
}

// @Summary Get new books
// @Description Get books marked as new (is_new=true). Optional ?limit=10 (default 10)
// @Tags Books
// @Produce json
// @Param limit query int false "Number of books (default 10)"
// @Success 200 {array} Book
// @Failure 500 {object} ErrorResponse
// @Router /books/new [get]
func getNewBooks(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	rows, err := db.Query(selectCols+`
		WHERE COALESCE(is_new, false) = true
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	books, err := scanBooks(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "scan error: " + err.Error()})
		return
	}
	if books == nil {
		books = []Book{}
	}
	c.JSON(http.StatusOK, books)
}

// @Summary Get featured books
// @Description Get books with rating >= 4.5 (top 10)
// @Tags Books
// @Produce json
// @Success 200 {array} Book
// @Failure 500 {object} ErrorResponse
// @Router /books/featured [get]
func getFeaturedBooks(c *gin.Context) {
	rows, err := db.Query(selectCols + `
		WHERE COALESCE(rating, 0.0) >= 4.5
		ORDER BY rating DESC, reviews_count DESC
		LIMIT 10
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	books, err := scanBooks(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "scan error: " + err.Error()})
		return
	}
	if books == nil {
		books = []Book{}
	}
	c.JSON(http.StatusOK, books)
}

// @Summary Get discounted books
// @Description Get books with discount > 0
// @Tags Books
// @Produce json
// @Success 200 {array} Book
// @Failure 500 {object} ErrorResponse
// @Router /books/discounted [get]
func getDiscountedBooks(c *gin.Context) {
	rows, err := db.Query(selectCols + `
		WHERE COALESCE(discount, 0) > 0
		ORDER BY discount DESC
		LIMIT 10
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	books, err := scanBooks(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "scan error: " + err.Error()})
		return
	}
	if books == nil {
		books = []Book{}
	}
	c.JSON(http.StatusOK, books)
}

// @Summary Get categories
// @Description Distinct categories
// @Tags Categories
// @Produce json
// @Success 200 {array} string
// @Failure 500 {object} ErrorResponse
// @Router /categories [get]
func getCategories(c *gin.Context) {
	rows, err := db.Query(`
		SELECT DISTINCT COALESCE(category, '') AS category
		FROM books
		WHERE category IS NOT NULL AND category <> ''
		ORDER BY category
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var cats []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		cats = append(cats, s)
	}
	if cats == nil {
		cats = []string{}
	}
	c.JSON(http.StatusOK, cats)
}

// @Summary Search books
// @Description Search by title or author (?q=keyword)
// @Tags Books
// @Produce json
// @Param q query string true "keyword"
// @Success 200 {array} Book
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /books/search [get]
func searchBooks(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing query parameter q"})
		return
	}

	rows, err := db.Query(selectCols+`
		WHERE title ILIKE '%' || $1 || '%'
		   OR author ILIKE '%' || $1 || '%'
		ORDER BY id ASC
	`, q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	books, err := scanBooks(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "scan error: " + err.Error()})
		return
	}
	if books == nil {
		books = []Book{}
	}
	c.JSON(http.StatusOK, books)
}

// @Summary Get book by ID
// @Description Retrieve one book
// @Tags Books
// @Produce json
// @Param id path int true "Book ID"
// @Success 200 {object} Book
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /books/{id} [get]
func getBook(c *gin.Context) {
	id := c.Param("id")
	var b Book

	err := db.QueryRow(selectCols+` WHERE id = $1`, id).Scan(
		&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Year, &b.Price,
		&b.Category, &b.OriginalPrice, &b.Discount, &b.CoverImage,
		&b.Rating, &b.ReviewsCount, &b.IsNew, &b.Pages,
		&b.Language, &b.Publisher, &b.Description,
		&b.CreatedAt, &b.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, b)
}

// @Summary Create a new book
// @Tags Books
// @Accept json
// @Produce json
// @Param book body Book true "Book data (รองรับฟิลด์เสริมถ้า DB มีคอลัมน์)"
// @Success 201 {object} Book
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /books [post]
func createBook(c *gin.Context) {
	var nb Book
	if err := c.ShouldBindJSON(&nb); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var id int
	var createdAt, updatedAt time.Time
	err := db.QueryRow(`
		INSERT INTO books (title, author, isbn, year, price)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`, nb.Title, nb.Author, nb.ISBN, nb.Year, nb.Price).
		Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	nb.ID = id
	nb.CreatedAt = createdAt
	nb.UpdatedAt = updatedAt
	c.JSON(http.StatusCreated, nb)
}

// @Summary Update a book
// @Tags Books
// @Accept json
// @Produce json
// @Param id path int true "Book ID"
// @Param book body Book true "Updated book"
// @Success 200 {object} Book
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /books/{id} [put]
func updateBook(c *gin.Context) {
	id := c.Param("id")
	var ub Book
	if err := c.ShouldBindJSON(&ub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var bookID int
	var updatedAt time.Time
	err := db.QueryRow(`
		UPDATE books
		SET title=$1, author=$2, isbn=$3, year=$4, price=$5
		WHERE id=$6
		RETURNING id, updated_at
	`, ub.Title, ub.Author, ub.ISBN, ub.Year, ub.Price, id).
		Scan(&bookID, &updatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ub.ID = bookID
	ub.UpdatedAt = updatedAt
	c.JSON(http.StatusOK, ub)
}

// @Summary Delete a book
// @Tags Books
// @Produce json
// @Param id path int true "Book ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /books/{id} [delete]
func deleteBook(c *gin.Context) {
	id := c.Param("id")
	res, err := db.Exec(`DELETE FROM books WHERE id=$1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "book deleted successfully"})
}

// @title       Bookstore API (Week 11)
// @version     1.0
// @description Gin + PostgreSQL + Swagger
// @host        localhost:8080
// @BasePath    /api/v1
func main() {
	initDB()
	defer db.Close()

	r := gin.Default()

	// CORS
	cfg := cors.DefaultConfig()
	cfg.AllowAllOrigins = true
	cfg.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(cfg))

	// Swagger UI
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health
	r.GET("/health", getHealth)

	// API
	api := r.Group("/api/v1")
	{
		api.GET("/books/search", searchBooks)
		api.GET("/books/featured", getFeaturedBooks)
		api.GET("/books/new", getNewBooks)
		api.GET("/books/discounted", getDiscountedBooks)

		api.GET("/books", getAllBooks)
		api.POST("/books", createBook)

		api.GET("/books/:id", getBook)
		api.PUT("/books/:id", updateBook)
		api.DELETE("/books/:id", deleteBook)

		api.GET("/categories", getCategories)
	}

	r.Run(":8080")
}
