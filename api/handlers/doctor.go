package handlers

import (
	"context"
	"github.com/hamza/proglabodev3/api/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hamza/proglabodev3/api/config"
	"github.com/hamza/proglabodev3/api/models"
	"go.mongodb.org/mongo-driver/bson"
)

type CreateDoctorRequest struct {
	Email      string             `json:"email" binding:"required,email"`
	Password   string             `json:"password" binding:"required,min=6"`
	FirstName  string             `json:"firstName" binding:"required"`
	LastName   string             `json:"lastName" binding:"required"`
	Expertises []string           `json:"expertises" binding:"required,min=1"`
	WorkPlan   utils.PeriodicPlan `json:"workPlan" binding:"required"`
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
			Expertises: req.Expertises,
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
