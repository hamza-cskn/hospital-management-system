package handlers

import (
	"context"
	"errors"
	"github.com/hamza/proglabodev3/api/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hamza/proglabodev3/api/config"
	"github.com/hamza/proglabodev3/api/middleware"
	"github.com/hamza/proglabodev3/api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type RegisterRequest struct {
	Email      string      `json:"email" binding:"required,email"`
	Password   string      `json:"password" binding:"required,min=6"`
	FirstName  string      `json:"firstName" binding:"required"`
	LastName   string      `json:"lastName" binding:"required"`
	Role       models.Role `json:"role" binding:"required,oneof=patient doctor"`
	Expertises []string    `json:"expertises"`
	WorkPlan   utils.PeriodicPlan
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Register(db *config.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user := models.User{
			Email:     req.Email,
			Password:  req.Password,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Role:      req.Role,
			CreatedAt: time.Time{},
		}

		userId, err := CreateUser(user, db)
		if err != nil {
			switch err.Error() {
			case "email already exists":
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			case "failed to hash password":
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			}
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "User registered successfully",
			"userId":  userId,
		})
	}
}

func Login(db *config.Database, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Find user by email
		var user models.User
		err := db.DB.Collection("users").FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&user)
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// Verify password
		if err := user.ComparePassword(req.Password); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// Generate JWT token
		expiration := time.Now().Add(cfg.TokenExpiration).Unix()
		token, err := middleware.GenerateToken(user.ID, string(user.Role), cfg.JWTSecret, expiration)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user": gin.H{
				"id":        user.ID,
				"email":     user.Email,
				"firstName": user.FirstName,
				"lastName":  user.LastName,
				"role":      user.Role,
			},
		})
	}
}
