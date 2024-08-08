package helpers

import (
	"context"
	"log"
	"mime/multipart"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	// "github.com/joho/godotenv"
)

// Global session variables
var s3Client *s3.Client
var kmsClient *kms.Client

func init() {
	// err := godotenv.Load()
	// if err != nil {
	//     log.Fatalf("Error loading .env file")
	// }

	// Load environment variables
	key := os.Getenv("SPACES_KEY")
	secret := os.Getenv("SPACES_SECRET")
	awsKey := os.Getenv("AWS_ACCESS")
	awsSecret := os.Getenv("AWS_SECRET")

	// Setup DigitalOcean Spaces session
	s3Cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(key, secret, "")),
	)
	if err != nil {
		log.Fatalf("Error creating S3 session: %s", err)
	}

	// Customize S3 endpoint for DigitalOcean Spaces
	s3Client = s3.NewFromConfig(s3Cfg, func(o *s3.Options) {
		o.EndpointResolver = s3.EndpointResolverFromURL("https://nyc3.digitaloceanspaces.com")
	})

	// Setup AWS KMS session
	kmsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-east-2"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsKey, awsSecret, "")),
	)
	if err != nil {
		log.Fatalf("Error creating KMS session: %s", err)
	}

	kmsClient = kms.NewFromConfig(kmsCfg)
	log.Print("S3 and KMS sessions created")
}

func GetS3Client() *s3.Client {
	return s3Client
}

func GetKMSClient() *kms.Client {
	return kmsClient
}

func UploadFileToS3(ctx context.Context, s3Client *s3.Client, bucket string, key string, file multipart.File) error {
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
		ACL:    "public-read", // Set the ACL to public-read
	})
	return err
}
