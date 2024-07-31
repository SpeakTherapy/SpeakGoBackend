package helpers

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/joho/godotenv"
)

var s3Session *session.Session

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	key := os.Getenv("SPACES_KEY")
	secret := os.Getenv("SPACES_SECRET")

	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(key, secret, ""),
		Endpoint:         aws.String("https://nyc3.digitaloceanspaces.com"),
		Region:           aws.String("us-east-1"),
		S3ForcePathStyle: aws.Bool(false),
	}

	newSession, err := session.NewSession(s3Config)
	if err != nil {
		log.Fatalf("Error creating session: %s", err)
	}
	s3Session = newSession
	log.Print("S3 session created")
}

func GetS3Session() *session.Session {
	return s3Session
}