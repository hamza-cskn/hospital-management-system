package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AppointmentStatus string

const (
	StatusPending   AppointmentStatus = "pending"
	StatusConfirmed AppointmentStatus = "confirmed"
	StatusCancelled AppointmentStatus = "cancelled"
	StatusCompleted AppointmentStatus = "completed"
)

type Appointment struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	PatientID   primitive.ObjectID `json:"patientId" bson:"patientId"`
	DoctorID    primitive.ObjectID `json:"doctorId" bson:"doctorId"`
	StartTime   time.Time          `json:"startTime" bson:"startTime"`
	EndTime     time.Time          `json:"endTime" bson:"endTime"`
	Status      AppointmentStatus  `json:"status" bson:"status"`
	Notes       string             `json:"notes" bson:"notes"`
	Department  string             `json:"department" bson:"department"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	CancelledAt *time.Time         `json:"cancelledAt,omitempty" bson:"cancelledAt,omitempty"`
}
