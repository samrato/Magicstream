package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/samrato/magicstream/database"
	"github.com/samrato/magicstream/models"
	"github.com/samrato/magicstream/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userValidate = validator.New()

// ========================== REGISTER USER ==========================
func RegisterUser(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.UserRegister
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if err := userValidate.Struct(input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		collection := database.GetCollection(client, "users")

		// Check if email exists
		count, err := collection.CountDocuments(ctx, bson.M{"email": input.Email})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}

		// Hash password
		hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		newUser := models.User{
			// UserID:          utils.GenerateID(), // utility to generate unique user ID
			FirstName:       input.FirstName,
			LastName:        input.LastName,
			Email:           input.Email,
			Password:        string(hashed),
			Role:            "USER",
			FavouriteGenres: input.FavouriteGenres,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		result, err := collection.InsertOne(ctx, newUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		accessToken, refreshToken, err := utils.GenerateTokens(newUser.UserID, newUser.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"user_id":          newUser.UserID,
			"first_name":       newUser.FirstName,
			"last_name":        newUser.LastName,
			"email":            newUser.Email,
			"role":             newUser.Role,
			"favourite_genres": newUser.FavouriteGenres,
			"token":            accessToken,
			"refresh_token":    refreshToken,
			"inserted_id":      result.InsertedID,
		})
	}
}

// ========================== LOGIN USER ==========================
func LoginUser(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.UserLogin
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if err := userValidate.Struct(input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		collection := database.GetCollection(client, "users")
		var user models.User
		err := collection.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		accessToken, refreshToken, err := utils.GenerateTokens(user.UserID, user.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id":          user.UserID,
			"first_name":       user.FirstName,
			"last_name":        user.LastName,
			"email":            user.Email,
			"role":             user.Role,
			"favourite_genres": user.FavouriteGenres,
			"token":            accessToken,
			"refresh_token":    refreshToken,
		})
	}
}

// ========================== LOGOUT USER ==========================
func LogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// In JWT, logout is handled client-side by deleting the token
		c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
	}
}

// ========================== REFRESH TOKEN ==========================
func RefreshTokenHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		refresh := c.GetHeader("Authorization")
		if refresh == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token missing"})
			return
		}

		claims, err := utils.ValidateToken(refresh, []byte(utils.GetEnv("JWT_REFRESH_SECRET")))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
			return
		}

		accessToken, refreshToken, err := utils.GenerateTokens(claims.UserID, claims.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": accessToken, "refresh_token": refreshToken})
	}
}

// ========================== GET USER PROFILE ==========================
func GetUserProfile(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := utils.GetUserIdFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		collection := database.GetCollection(client, "users")
		var user models.User
		err = collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id":          user.UserID,
			"first_name":       user.FirstName,
			"last_name":        user.LastName,
			"email":            user.Email,
			"role":             user.Role,
			"favourite_genres": user.FavouriteGenres,
		})
	}
}

// ========================== UPDATE FAVOURITE GENRES ==========================
func UpdateFavouriteGenres(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := utils.GetUserIdFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var input struct {
			FavouriteGenres []models.Genre `json:"favourite_genres" validate:"dive"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if err := userValidate.Struct(input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		collection := database.GetCollection(client, "users")
		update := bson.M{"$set": bson.M{"favourite_genres": input.FavouriteGenres, "updated_at": time.Now()}}
		_, err = collection.UpdateOne(ctx, bson.M{"user_id": userID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update favourite genres"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Favourite genres updated successfully"})
	}
}
