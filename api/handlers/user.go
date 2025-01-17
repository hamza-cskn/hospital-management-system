package handlers

import (
	"context"
	"fmt"
	"github.com/hamza/proglabodev3/api/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hamza/proglabodev3/api/config"
	"github.com/hamza/proglabodev3/api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UpdateProfileRequest struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	FirstName  string             `json:"firstName,omitempty"`
	LastName   string             `json:"lastName,omitempty"`
	Email      string             `json:"email,omitempty"`
	Password   string             `json:"password,omitempty"`
	WorkPlan   utils.PeriodicPlan `json:"workPlan,omitempty"`
	Expertises []string           `json:"expertises,omitempty"`
}

func GetUserProfile(db *config.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Request.Context().Value("userId").(primitive.ObjectID)

		var user models.User
		err := db.DB.Collection("users").FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Remove sensitive information
		user.Password = ""

		c.JSON(http.StatusOK, user)
	}
}

func UpdateUserProfile(db *config.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("userId")

		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
			return
		}
		id, err := primitive.ObjectIDFromHex(userID)

		var req UpdateProfileRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Build update document
		update := bson.M{}

		if req.FirstName != "" {
			update["firstName"] = req.FirstName
		}

		if req.LastName != "" {
			update["lastName"] = req.LastName
		}

		if req.Email != "" {
			// Check if email is already taken
			var existingUser models.User
			err := db.DB.Collection("users").FindOne(context.Background(), bson.M{
				"email": req.Email,
				"_id":   bson.M{"$ne": id},
			}).Decode(&existingUser)
			if err == nil {
				c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
				return
			}
			update["email"] = req.Email
		}

		if req.Password != "" {
			update["password"], _ = models.HashStr(req.Password)
		}

		if req.WorkPlan.Periods != nil {
			update["workPlan"] = req.WorkPlan
		}

		if req.Expertises != nil {
			update["expertises"] = req.Expertises
		}

		result, err := db.DB.Collection("users").UpdateOne(
			context.Background(),
			bson.M{"_id": id},
			bson.M{"$set": update},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found: " + userID})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
	}
}

// CreateUser handles the creation of a new user in the database
func CreateUser(user models.User, db *config.Database) (interface{}, error) {
	// Check if email already exists
	var existingUser models.User
	err := db.DB.Collection("users").FindOne(context.Background(), bson.M{"email": user.Email}).Decode(&existingUser)
	if err == nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Hash password
	if err := user.HashPassword(); err != nil {
		return nil, fmt.Errorf("failed to hash password")
	}

	// Set timestamps
	user.CreatedAt = time.Now()

	// Insert user into database
	result, err := db.DB.Collection("users").InsertOne(context.Background(), user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user")
	}

	return result.InsertedID, nil
}
