package controllers

import (
	"context"
	"fmt"
	"golang-speakbackend/database"
	"golang-speakbackend/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var exerciseCollection *mongo.Collection = database.OpenCollection(database.Client, "exercise")
var validate = validator.New()

func GetExercises() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
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

		// Count total number of exercises
		countStage := bson.D{{Key: "$count", Value: "total"}}

		// Skip and Limit for pagination
		skipStage := bson.D{{Key: "$skip", Value: startIndex}}
		limitStage := bson.D{{Key: "$limit", Value: recordPerPage}}

		// Project stage to include necessary fields
		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 1},
				{Key: "name", Value: 1},
				{Key: "description", Value: 1},
				{Key: "video_url", Value: 1},
				{Key: "tags", Value: 1},
				{Key: "created_at", Value: 1},
				{Key: "updated_at", Value: 1},
				{Key: "exercise_id", Value: 1},
			}},
		}

		// Aggregate total exercises count separately
		countResult, err := exerciseCollection.Aggregate(ctx, mongo.Pipeline{matchStage, countStage})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while counting exercises"})
			return
		}
		var countData []bson.M
		if err = countResult.All(ctx, &countData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while counting exercises"})
			return
		}
		totalCount := 0
		if len(countData) > 0 {
			totalCount = int(countData[0]["total"].(int32)) // Convert int32 to int
		}

		// Aggregate exercises with pagination
		result, err := exerciseCollection.Aggregate(ctx, mongo.Pipeline{matchStage, skipStage, limitStage, projectStage})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching exercises"})
			return
		}

		var exercises []bson.M
		if err = result.All(ctx, &exercises); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching exercises"})
			return
		}

		response := gin.H{
			"total":     totalCount,
			"exercises": exercises,
		}

		c.JSON(http.StatusOK, response)
	}
}


func GetExercise() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		exerciseID := c.Param("exercise_id")
		var exercise models.Exercise

		err := exerciseCollection.FindOne(ctx, bson.M{"exercise_id": exerciseID}).Decode(&exercise)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching exercise"})
			return
		}
		c.JSON(http.StatusOK, exercise)
	}
}

func CreateExercise() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		var exercise models.Exercise

		defer cancel()
		if err := c.BindJSON(&exercise); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationError := validate.Struct(exercise)
		if validationError != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationError.Error()})
			return
		}

		exercise.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		exercise.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		exercise.ID = primitive.NewObjectID()
		exercise.ExerciseID = exercise.ID.Hex()

		result, insertErr := exerciseCollection.InsertOne(ctx, exercise)
		if insertErr != nil {
			msg := fmt.Sprintf("Error while inserting exercise: %s", insertErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, result)
	}
}

func UpdateExercise() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		var exercise models.Exercise

		defer cancel()
		if err := c.BindJSON(&exercise); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		exerciseID := c.Param("exercise_id")
		filter := bson.M{"exercise_id": exerciseID}

		var updateObj primitive.D

		if exercise.Name != nil {
			updateObj = append(updateObj, bson.E{Key: "name", Value: exercise.Name})
		}
		if exercise.Description != nil {
			updateObj = append(updateObj, bson.E{Key: "description", Value: exercise.Description})
		}
		if exercise.VideoURL != "" {
			updateObj = append(updateObj, bson.E{Key: "video_url", Value: exercise.VideoURL})
		}
		if exercise.Tags != nil {
			updateObj = append(updateObj, bson.E{Key: "tags", Value: exercise.Tags})
		}

		exercise.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: exercise.UpdatedAt})

		upsert := true

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := exerciseCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{Key: "$set", Value: updateObj},
			},
			&opt,
		)

		if err != nil {
			msg := fmt.Sprintf("Exercise update failed: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}

		defer cancel()
		c.JSON(http.StatusOK, result)
	}
}

func DeleteExercise() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
