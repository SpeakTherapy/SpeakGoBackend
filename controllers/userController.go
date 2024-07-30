package controllers

import (
	"context"
	"golang-speakbackend/database"
	"golang-speakbackend/helpers"
	"golang-speakbackend/models"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

// func UploadProfilePicture(s3Session ) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		file, header, err := c.Request.FormFile("profile_picture")
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile picture upload"})
// 			return
// 		}
// 		defer file.Close()

// 		fileName := fmt.Sprintf("profile-picture/%s-%s", c.Param("user_id"), header.Filename)

// 		var buf bytes.Buffer
// 		io.Copy(&buf, file)
// 		_, err = s3Session.PutObject(&s3.PutObjectInput{
// 			Bucket:        aws.String(os.Getenv("BUCKET_NAME")),
// 			Key:           aws.String(fileName),
// 			Body:          bytes.NewReader(buf.Bytes()),
// 			ContentLength: aws.Int64(int64(buf.Len())),
// 			ContentType:   aws.String(http.DetectContentType(buf.Bytes())),
// 			ACL:           aws.String("public-read"),
// 		})
// 	}
// }

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		userID := c.Param("user_id")

		var user models.User

		err := userCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)

		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while listing the users"})
		}

		c.JSON(http.StatusOK, user)
	}
}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage

		// Match stage
		matchStage := bson.D{{Key: "$match", Value: bson.D{}}}

		// Count total number of users
		countStage := bson.D{{Key: "$count", Value: "total"}}

		// Skip and Limit for pagination
		skipStage := bson.D{{Key: "$skip", Value: startIndex}}
		limitStage := bson.D{{Key: "$limit", Value: recordPerPage}}

		// Project stage to include necessary fields
		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 1},
				{Key: "first_name", Value: 1},
				{Key: "last_name", Value: 1},
				{Key: "email", Value: 1},
				{Key: "role", Value: 1},
				{Key: "profile_image", Value: 1},
				{Key: "created_at", Value: 1},
				{Key: "updated_at", Value: 1},
				{Key: "user_id", Value: 1},
			}},
		}

		// Aggregate total users count separately
		countResult, err := userCollection.Aggregate(ctx, mongo.Pipeline{matchStage, countStage})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while counting users"})
			return
		}
		var countData []bson.M
		if err = countResult.All(ctx, &countData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while counting users"})
			return
		}
		totalCount := 0
		if len(countData) > 0 {
			totalCount = int(countData[0]["total"].(int32)) // Convert int32 to int
		}

		// Aggregate users with pagination
		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{matchStage, skipStage, limitStage, projectStage})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching users"})
			return
		}

		var users []bson.M
		if err = result.All(ctx, &users); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching users"})
			return
		}

		response := gin.H{
			"total": totalCount,
			"users": users,
		}

		c.JSON(http.StatusOK, response)
	}
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists"})
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.UserID = user.ID.Hex()

		token, refreshToken, _ := helpers.GenerateAllTokens(*user.Email, *user.FirstName, *user.LastName, user.UserID)
		user.Token = &token
		user.RefreshToken = &refreshToken

		if user.Role == "patient" {
			user.ReferenceCode = ""
		} else if user.Role == "therapist" {
			user.ReferenceCode = generateUniqueReferenceCode(ctx)
		}

		result, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User could not be created"})
			return
		}

		// return status OK and send result back
		c.JSON(http.StatusOK, result)
	}
}

func generateUniqueReferenceCode(ctx context.Context) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var code string

	for {
		code = generateRandomString(8, charset)
		count, _ := userCollection.CountDocuments(ctx, bson.M{"reference_code": code})
		if count == 0 {
			break
		}
	}
	return code
}

func generateRandomString(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Email not found"})
			return
		}

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		if !passwordIsValid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			return
		}

		token, refreshToken, _ := helpers.GenerateAllTokens(*foundUser.Email, *foundUser.FirstName, *foundUser.LastName, foundUser.UserID)
		helpers.UpdateAllTokens(token, refreshToken, foundUser.UserID)

		// return status OK and send result back
		c.JSON(http.StatusOK, foundUser)
	}
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""
	if err != nil {
		msg = "Password is incorrect"
		check = false
	}
	return check, msg
}

func UpdateUser() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func DeleteUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		userID := c.Param("user_id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
			return
		}

		objID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID"})
			return
		}

		filter := bson.M{"_id": objID}

		result, err := userCollection.DeleteOne(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred while deleting the user"})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
	}
}

func RefreshToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		type RefreshTokenRequest struct {
			RefreshToken string `json:"refresh_token" bson:"refresh_token" binding:"required"`
		}

		var req RefreshTokenRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		claims, msg := helpers.ValidateToken(req.RefreshToken)
		if msg != "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			return
		}

		userID := claims.UserID
		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}

		token, refreshToken, err := helpers.GenerateAllTokens(*user.Email, *user.FirstName, *user.LastName, user.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while generating token"})
			return
		}

		helpers.UpdateAllTokens(token, refreshToken, user.UserID)

		c.JSON(http.StatusOK, gin.H{"token": token, "refresh_token": refreshToken})
	}
}

func LinkToTherapist() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		type RequestBody struct {
			ReferenceCode string `json:"reference_code" bson:"reference_code" binding:"required"`
		}
		var requestBody RequestBody

		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		userID := c.Param("user_id")
		var patient models.User
		var therapist models.User

		err := userCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&patient)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}

		if patient.Role != "patient" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Only patients can link to therapists"})
			return
		}

		err = userCollection.FindOne(ctx, bson.M{"reference_code": requestBody.ReferenceCode, "role": "therapist"}).Decode(&therapist)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Therapist not found"})
			return
		}

		// add reference code to patient
		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "reference_code", Value: therapist.ReferenceCode},
			}},
		}

		_, err = userCollection.UpdateOne(ctx, bson.M{"user_id": userID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while linking to therapist"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Linked to therapist successfully"})
	}
}

func GetPatients() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		therapistID := c.Param("therapist_id")
		if therapistID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Therapist ID is required"})
			return
		}

		var therapist models.User
		err := userCollection.FindOne(ctx, bson.M{"user_id": therapistID}).Decode(&therapist)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Therapist not found"})
			return
		}

		if therapist.Role != "therapist" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Only therapists can view patients"})
			return
		}

		var patients []models.User
		cursor, err := userCollection.Find(ctx, bson.M{"reference_code": therapist.ReferenceCode, "role": "patient"})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching patients"})
			return
		}

		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var patient models.User
			cursor.Decode(&patient)
			patients = append(patients, patient)
		}

		c.JSON(http.StatusOK, patients)
	}
}
