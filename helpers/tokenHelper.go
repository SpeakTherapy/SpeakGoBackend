package helpers

import (
	"context"
	"golang-speakbackend/database"
	"log"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type SignedDetails struct {
	Email     string
	FirstName string
	LastName  string
	UserID    string
	jwt.StandardClaims
}

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

var SECRET_KEY string = os.Getenv("SECRET_KEY")

func GenerateAllTokens(email string, firstName string, lastName string, userID string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		UserID:    userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(), // 1 day expiry
		},
	}

	refreshClaims := &SignedDetails{
		UserID: userID,
		Email:  email, // Include minimal necessary information
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24*7)).Unix(), // 1 week expiry
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
		return
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
		return
	}

	return token, refreshToken, err
}

func UpdateAllTokens(signedToken string, signedRefreshToken string, userID string) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}

	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "token", Value: signedToken},
			{Key: "refresh_token", Value: signedRefreshToken},
			{Key: "updated_at", Value: time.Now()},
		}},
	}

	result, err := userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("Could not update tokens for user %s: %v", userID, err)
		return err
	}

	if result.MatchedCount == 0 {
		log.Printf("No document matched for user ID %s. No update occurred.", userID)
	} else {
		log.Printf("Updated tokens for user ID %s", userID)
	}

	return nil
}

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {

	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		// msg = fmt.Sprint("the token is invalid")
		msg = err.Error()
		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		// msg = fmt.Sprint("the token has expired")
		msg = err.Error()
		return
	}

	return claims, msg

}
