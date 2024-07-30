package main

import (
	controller "golang-speakbackend/controllers"
	middleware "golang-speakbackend/middleware"
	routes "golang-speakbackend/routes"
	"log"
	"os"

	"github.com/Backblaze/blazer/b2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// var exerciseCollection *mongo.Collection = database.OpenCollection(database.Client, "exercise")

var b2Client *b2.Client
var b2Bucket *b2.Bucket
var s3Session *s3.S3

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	keyID := os.Getenv("B2_KEY_ID")
	applicationKey := os.Getenv("B2_APPLICATION_KEY")
	// bucketName := os.Getenv("B2_BUCKET_NAME")

	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(keyID, applicationKey, ""),
		Endpoint:    aws.String("https://s3.us-east-005.backblazeb2.com"),
		Region:      aws.String("us-east-005"),
	})
	if err != nil {
		log.Fatalf("Error creating new session: %v", err)
	}
	s3Session = s3.New(sess)
	log.Print("S3 session created")
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	router := gin.New()
	router.Use(gin.Logger())

	// Public routes
	publicRoutes := router.Group("/")
	{
		publicRoutes.POST("/signup", controller.SignUp())
		publicRoutes.POST("/login", controller.Login())
		publicRoutes.POST("/refresh", controller.RefreshToken()) // Refresh token doesn't need auth middleware
	}

	// Private routes
	privateRoutes := router.Group("/")
	privateRoutes.Use(middleware.Authentication())
	{
		routes.UserRoutes(privateRoutes)
		routes.ExerciseRoutes(privateRoutes)
		routes.PatientExerciseRoutes(privateRoutes)
	}

	router.Run(":" + port)
}
