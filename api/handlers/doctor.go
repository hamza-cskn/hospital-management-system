package handlers

import (
	"context"
	"github.com/hamza/proglabodev3/api/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hamza/proglabodev3/api/config"
	"github.com/hamza/proglabodev3/api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateDoctorRequest struct {
	Email      string             `json:"email" binding:"required,email"`
	Password   string             `json:"password" binding:"required,min=6"`
	FirstName  string             `json:"firstName" binding:"required"`
	LastName   string             `json:"lastName" binding:"required"`
	Expertises []string           `json:"expertises" binding:"required,min=1"`
	WorkPlan   utils.PeriodicPlan `json:"workPlan" binding:"required"`
}

type UpdateExpertiseRequest struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ExpertiseIDs []string           `json:"expertiseIds" binding:"required"`
}

func CreateDoctor(db *config.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateDoctorRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if email already exists
		var existingUser models.User
		err := db.DB.Collection("users").FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&existingUser)
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}

		// Create doctor's expertises
		var expertises []models.Expertise
		for _, exp := range req.Expertises {
			expertise := models.Expertise{
				Name: exp,
			}
			expertises = append(expertises, expertise)
		}

		if len(req.WorkPlan.Periods) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Work plan must have at least one period"})
			return
		}

		// Create doctor
		doctor := models.User{
			Email:      req.Email,
			Password:   req.Password,
			FirstName:  req.FirstName,
			LastName:   req.LastName,
			Role:       models.RoleDoctor,
			CreatedAt:  time.Now(),
			Expertises: expertises,
			WorkPlan:   req.WorkPlan,
		}

		// Hash password
		if err := doctor.HashPassword(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		// Insert doctor into database
		result, err := db.DB.Collection("users").InsertOne(context.Background(), doctor)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create doctor"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Doctor created successfully",
			"id":      result.InsertedID,
		})
	}
}

func UpdateDoctorExpertise(db *config.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateExpertiseRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		doctorID := req.ID

		var cursor *mongo.Cursor
		var err error
		var expertiseIDs []primitive.ObjectID
		var expertises []models.Expertise

		if len(req.ExpertiseIDs) > 0 {
			for _, id := range req.ExpertiseIDs {
				oid, err := primitive.ObjectIDFromHex(id)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expertise ID format"})
					return
				}
				expertiseIDs = append(expertiseIDs, oid)
			}

			cursor, err = db.DB.Collection("expertises").Find(context.Background(), bson.M{
				"_id": bson.M{"$in": expertiseIDs},
			})

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify expertises"})
				return
			}
			defer cursor.Close(context.Background())

			if err := cursor.All(context.Background(), &expertises); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode expertises"})
				return
			}

			if len(expertises) != len(expertiseIDs) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "One or more expertise IDs are invalid"})
				return
			}
		}

		// Update doctor's expertises
		result, err := db.DB.Collection("users").UpdateOne(
			context.Background(),
			bson.M{"_id": doctorID, "role": models.RoleDoctor},
			bson.M{
				"$set": bson.M{
					"expertises": expertises,
				},
			},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update expertises"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Doctor not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Expertises updated successfully"})
	}
}
