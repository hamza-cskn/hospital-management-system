package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hamza/proglabodev3/api/config"
	"github.com/hamza/proglabodev3/api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CreateExpertiseRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type UpdateExpertiseRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

func DeleteExpertise(db *config.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		qid := c.Query("id")
		if qid == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Expertise ID is required"})
			return
		}

		expertiseID, err := primitive.ObjectIDFromHex(qid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expertise ID"})
			return
		}

		result, err := db.DB.Collection("expertises").DeleteOne(context.Background(), bson.M{"_id": expertiseID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete expertise"})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Expertise not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Expertise deleted successfully"})
	}
}

func CreateExpertise(db *config.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateExpertiseRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if expertise already exists
		var existingExpertise models.Expertise
		err := db.DB.Collection("expertises").FindOne(context.Background(), bson.M{
			"name": req.Name,
		}).Decode(&existingExpertise)
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Expertise already exists"})
			return
		}

		expertise := models.Expertise{
			Name:        req.Name,
			Description: req.Description,
			CreatedAt:   time.Now(),
		}

		result, err := db.DB.Collection("expertises").InsertOne(context.Background(), expertise)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create expertise"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Expertise created successfully",
			"id":      result.InsertedID,
		})
	}
}

func GetAllExpertises(db *config.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		cursor, err := db.DB.Collection("expertises").Find(context.Background(), bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch expertises"})
			return
		}
		defer cursor.Close(context.Background())

		var expertises []models.Expertise
		if err := cursor.All(context.Background(), &expertises); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode expertises"})
			return
		}
		if len(expertises) == 0 {
			expertises = []models.Expertise{}
		}

		c.JSON(http.StatusOK, expertises)
	}
}

func GetAllUsers(db *config.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.Query("role")
		filter := bson.M{}
		if role != "" {
			filter["role"] = role
		}

		cursor, err := db.DB.Collection("users").Find(context.Background(), filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
			return
		}
		defer cursor.Close(context.Background())

		var users []models.User
		if err := cursor.All(context.Background(), &users); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode users"})
			return
		}

		// Remove password field from response
		for i := range users {
			users[i].Password = ""
		}

		c.JSON(http.StatusOK, users)
	}
}

func UpdateUser(db *config.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var updateReq map[string]interface{}
		if err := c.ShouldBindJSON(&updateReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Don't allow updating role to admin through this endpoint
		if role, ok := updateReq["role"]; ok && role == string(models.RoleAdmin) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot update user to admin role"})
			return
		}

		result, err := db.DB.Collection("users").UpdateOne(
			context.Background(),
			bson.M{"_id": userID},
			bson.M{"$set": updateReq},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
	}
}

func DeleteUser(db *config.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		paramId := c.Param("id")

		if paramId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
			return
		}

		userID, err := primitive.ObjectIDFromHex(paramId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Check if user exists and is not an admin
		var user models.User
		err = db.DB.Collection("users").FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		if user.Role == models.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete admin user"})
			return
		}

		// Delete user's appointments
		_, err = db.DB.Collection("appointments").DeleteMany(context.Background(), bson.M{
			"$or": []bson.M{
				{"patientId": userID},
				{"doctorId": userID},
			},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user's appointments"})
			return
		}

		// Delete user
		result, err := db.DB.Collection("users").DeleteOne(context.Background(), bson.M{"_id": userID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User and related data deleted successfully"})
	}
}

func UpdateExpertise(db *config.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		expertiseID, err := primitive.ObjectIDFromHex(c.Query("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expertise ID"})
			return
		}

		var req UpdateExpertiseRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		newData := bson.M{
			"name":        req.Name,
			"description": req.Description,
		}
		update := bson.M{
			"$set": newData,
		}

		if req.Name != "" {
			newData["name"] = req.Name
		}

		if req.Description != "" {
			newData["description"] = req.Description
		}

		result, err := db.DB.Collection("expertises").UpdateOne(context.Background(), bson.M{"_id": expertiseID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update expertise"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Expertise not found: " + expertiseID.Hex()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Expertise updated successfully"})
	}
}
