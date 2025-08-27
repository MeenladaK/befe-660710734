package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// โครงสร้างข้อมูลภาพยนตร์
type Movie struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Director string `json:"director"`
	Year     int    `json:"year"`
	Genre    string `json:"genre"`
}

// จำลองฐานข้อมูลในหน่วยความจำ
var movies = []Movie{
	{ID: "1", Title: "The Shawshank Redemption", Director: "Frank Darabont", Year: 1994, Genre: "Drama"},
	{ID: "2", Title: "The Dark Knight", Director: "Christopher Nolan", Year: 2008, Genre: "Action"},
	{ID: "3", Title: "Inception", Director: "Christopher Nolan", Year: 2010, Genre: "Sci-Fi"},
	{ID: "4", Title: "Forrest Gump", Director: "Robert Zemeckis", Year: 1994, Genre: "Drama"},
	{ID: "5", Title: "The Matrix", Director: "Wachowski Brothers", Year: 1999, Genre: "Sci-Fi"},
}

// getMovies ดึงข้อมูลภาพยนตร์
func getMovies(c *gin.Context) {

	directorQuery := c.Query("director")

	if directorQuery != "" {
		filteredMovies := []Movie{}
		// วนลูปหาหนังที่มีผู้แต่งตรงกับ query
		for _, movie := range movies {
			if movie.Director == directorQuery {
				filteredMovies = append(filteredMovies, movie)
			}
		}
		// ส่งข้อมูลที่กรองแล้วกลับไป
		c.JSON(http.StatusOK, filteredMovies)
		return
	}

	// ถ้าไม่มี query parameter, ส่งหนังทั้งหมดกลับไป
	c.JSON(http.StatusOK, movies)
}

func main() {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "up and running"})
	})

	api := r.Group("/api/v1")
	{

		api.GET("/movies", getMovies)
	}

	fmt.Println("Movie API server is running at http://localhost:8080")
	r.Run(":8080")
}
