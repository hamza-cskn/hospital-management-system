package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hamza/proglabodev3/api/config"
	"github.com/hamza/proglabodev3/api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CreateAppointmentRequest struct {
	DoctorID   string    `json:"doctorId" binding:"required"`
	DateTime   time.Time `json:"dateTime" binding:"required"`
	Duration   string    `json:"duration" binding:"required"`
	Department string    `json:"department" binding:"required"`
	Notes      string    `json:"notes"`
}

func CreateAppointment(db *config.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateAppointmentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Parse the duration string to time.Duration
		duration, err := time.ParseDuration(req.Duration)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid duration format"})
			return
		}

		// Get user ID from context
		userID := c.Request.Context().Value("userId").(primitive.ObjectID)

		// Convert doctor ID string to ObjectID
		doctorID, err := primitive.ObjectIDFromHex(req.DoctorID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid doctor ID"})
			return
		}

		// Check if doctor exists
		var doctor models.User
		err = db.DB.Collection("users").FindOne(context.Background(), bson.M{
			"_id":  doctorID,
			"role": models.RoleDoctor,
		}).Decode(&doctor)
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Doctor not found"})
			return
		}

		// Check if the doctor is available at this time.
		if doctor.WorkPlan.IsIn(req.DateTime.Add(duration)) && doctor.WorkPlan.IsIn(req.DateTime) {
			c.JSON(http.StatusConflict, gin.H{"error": "Doctor is not available at this time"})
			return
		}

		// Check if an appointment already exists at this time for the doctor.
		var existingAppointment models.Appointment
		newStart := req.DateTime
		newEnd := req.DateTime.Add(duration)

		filter := bson.M{
			"doctorId":  doctorID,
			"status":    bson.M{"$ne": models.StatusCancelled},
			"startTime": bson.M{"$lt": newEnd},   // existing start < new end
			"endTime":   bson.M{"$gt": newStart}, // existing end > new start
		}
		err = db.DB.Collection("appointments").FindOne(context.Background(), filter).Decode(&existingAppointment)
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "An appointment already exists at this time for the doctor. Appointment ID:" + existingAppointment.ID.Hex()})
			return
		}

		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing appointments"})
			return
		}

		// Create appointment
		appointment := models.Appointment{
			PatientID:  userID,
			DoctorID:   doctorID,
			StartTime:  req.DateTime,
			Status:     models.StatusPending,
			Notes:      req.Notes,
			Department: req.Department,
			CreatedAt:  time.Now(),
			EndTime:    req.DateTime.Add(duration),
		}

		result, err := db.DB.Collection("appointments").InsertOne(context.Background(), appointment)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create appointment"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Appointment created successfully",
			"id":      result.InsertedID,
		})
	}
}

func GetAppointments(db *config.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		filter := bson.M{}

		userID := c.Query("userId")
		if userID != "" {
			fmt.Printf("userID: %v\n", userID)

			oid, err := primitive.ObjectIDFromHex(userID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
				return
			}

			filter = bson.M{"$or": []bson.M{{"patientId": oid}, {"doctorId": oid}}}
		}

		cursor, err := db.DB.Collection("appointments").Find(context.Background(), filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch appointments"})
			return
		}
		defer cursor.Close(context.Background())

		var appointments []models.Appointment
		if err := cursor.All(context.Background(), &appointments); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode appointments"})
			return
		}

		if len(appointments) == 0 {
			appointments = []models.Appointment{}
		}

		c.JSON(http.StatusOK, appointments)
	}
}

func GetAppointment(db *config.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment ID"})
			return
		}

		var appointment models.Appointment
		err = db.DB.Collection("appointments").FindOne(context.Background(), bson.M{"_id": id}).Decode(&appointment)
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
			return
		}

		userID := c.Request.Context().Value("userId").(primitive.ObjectID)
		userRole := c.Request.Context().Value("userRole").(string)

		if userRole == string(models.RolePatient) && appointment.PatientID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to view this appointment"})
			return
		}

		if userRole == string(models.RoleDoctor) && appointment.DoctorID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to view this appointment"})
			return
		}

		c.JSON(http.StatusOK, appointment)
	}
}

func UpdateAppointment(db *config.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment ID"})
			return
		}

		var updateReq map[string]interface{}
		if err := c.ShouldBindJSON(&updateReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := db.DB.Collection("appointments").UpdateOne(
			context.Background(),
			bson.M{"_id": id},
			bson.M{"$set": updateReq},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update appointment"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Appointment updated successfully"})
	}
}

func DeleteAppointment(db *config.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment ID"})
			return
		}

		userID := c.Request.Context().Value("userId").(primitive.ObjectID)
		userRole := c.Request.Context().Value("userRole").(string)

		var filter bson.M
		if userRole == string(models.RolePatient) {
			filter = bson.M{"_id": id, "patientId": userID}
		} else if userRole == string(models.RoleDoctor) {
			filter = bson.M{"_id": id, "doctorId": userID}
		} else {
			filter = bson.M{"_id": id} // Admin can delete any appointment
		}

		result, err := db.DB.Collection("appointments").DeleteOne(context.Background(), filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete appointment"})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found or not authorized to delete"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Appointment deleted successfully"})
	}
}
