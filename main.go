package main

import (
	controller "golang-speakbackend/controllers"
	routes "golang-speakbackend/routes"
	// "log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	// "github.com/joho/godotenv"
)


// func init() {
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatalf("Error loading .env file")
// 	}
// }

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	router := gin.New()
	router.Use(gin.Logger())

	// Enable CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Public routes
	publicRoutes := router.Group("/")
	{
		publicRoutes.POST("/signup", controller.SignUp())
		publicRoutes.POST("/login", controller.Login())
		publicRoutes.POST("/refresh", controller.RefreshToken()) // Refresh token doesn't need auth middleware
	}

	// Private routes
	privateRoutes := router.Group("/")
	{
		routes.UserRoutes(privateRoutes)
		routes.ExerciseRoutes(privateRoutes)
		routes.PatientExerciseRoutes(privateRoutes)
	}

	router.Run(":" + port)
}
