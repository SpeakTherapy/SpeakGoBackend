package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PatientExercise struct {
	ID                primitive.ObjectID `json:"id" bson:"_id"`
	PatientID         *string            `json:"patient_id" bson:"patient_id" validate:"required"`
	TherapistID       *string            `json:"therapist_id" bson:"therapist_id" validate:"required"`
	ExerciseID        *string            `json:"exercise_id" bson:"exercise_id" validate:"required"`
	Status            string             `json:"status" validate:"required,eq=PENDING|eq=COMPLETED"`
	Recording         string             `json:"recording"`
	CreatedAt         time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at" bson:"updated_at"`
	PatientExerciseID string             `json:"patient_exercise_id" bson:"patient_exercise_id"`
}