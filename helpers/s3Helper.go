package helpers

import (
    "log"
    "os"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    // "github.com/joho/godotenv"
)

var s3Session *session.Session
var kmsSession *session.Session

func init() {
    // err := godotenv.Load()
    // if err != nil {
    //     log.Fatalf("Error loading .env file")
    // }

    key := os.Getenv("SPACES_KEY")
    secret := os.Getenv("SPACES_SECRET")
	awsKey := os.Getenv("AWS_ACCESS")
	awsSecret := os.Getenv("AWS_SECRET")

    // Setup DigitalOcean Spaces session
    s3Config := &aws.Config{
        Credentials:      credentials.NewStaticCredentials(key, secret, ""),
        Endpoint:         aws.String("https://nyc3.digitaloceanspaces.com"),
        Region:           aws.String("us-east-1"),
        S3ForcePathStyle: aws.Bool(false),
    }

    newS3Session, err := session.NewSession(s3Config)
    if err != nil {
        log.Fatalf("Error creating S3 session: %s", err)
    }
    s3Session = newS3Session
    log.Print("S3 session created")

    // Setup AWS KMS session
    kmsConfig := &aws.Config{
		Credentials: credentials.NewStaticCredentials(awsKey, awsSecret, ""),
        Region: aws.String("us-east-2"), // KMS key region
    }

    newKMSSession, err := session.NewSession(kmsConfig)
    if err != nil {
        log.Fatalf("Error creating KMS session: %s", err)
    }
    kmsSession = newKMSSession
    log.Print("KMS session created")
}

func GetS3Session() *session.Session {
    return s3Session
}

func GetKMSSession() *session.Session {
    return kmsSession
}
