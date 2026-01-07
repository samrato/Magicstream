package controllers

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/samrato/magicstream/database"
	"github.com/samrato/magicstream/models"
	"github.com/samrato/magicstream/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms/openai"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var validate = validator.New()

// ========================== MOVIES ==========================

func GetMovies(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		collection := database.GetCollection(client, "movies")
		cursor, err := collection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch movies"})
			return
		}
		defer cursor.Close(ctx)

		var movies []models.Movie
		if err := cursor.All(ctx, &movies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode movies"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"count": len(movies), "data": movies})
	}
}

func GetMovie(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		imdbID := c.Param("imdb_id")
		if imdbID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "imdb_id is required"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		collection := database.GetCollection(client, "movies")
		var movie models.Movie
		err := collection.FindOne(ctx, bson.M{"imdb_id": imdbID}).Decode(&movie)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		c.JSON(http.StatusOK, movie)
	}
}

func AddMovie(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var movie models.Movie
		if err := c.ShouldBindJSON(&movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if err := validate.Struct(movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		collection := database.GetCollection(client, "movies")
		result, err := collection.InsertOne(ctx, movie)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add movie"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"inserted_id": result.InsertedID})
	}
}

// ========================== ADMIN REVIEW ==========================

func AdminReviewUpdate(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		imdbID := c.Param("imdb_id")
		if imdbID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "imdb_id required"})
			return
		}

		var req struct {
			AdminReview string `json:"admin_review"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		sentiment, rankVal, err := GetReviewRanking(req.AdminReview, client, c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		update := bson.M{
			"$set": bson.M{
				"admin_review": req.AdminReview,
				"ranking": bson.M{
					"ranking_name":  sentiment,
					"ranking_value": rankVal,
				},
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		collection := database.GetCollection(client, "movies")
		res, err := collection.UpdateOne(ctx, bson.M{"imdb_id": imdbID}, update)
		if err != nil || res.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"admin_review": req.AdminReview, "ranking": sentiment})
	}
}

// ========================== AI RANKING ==========================

func GetReviewRanking(review string, client *mongo.Client, c *gin.Context) (string, int, error) {
	rankings, err := GetRankings(client)
	if err != nil {
		return "", 0, err
	}

	var names []string
	for _, r := range rankings {
		if r.RankingValue != 999 {
			names = append(names, r.RankingName)
		}
	}

	_ = godotenv.Load(".env")
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", 0, errors.New("OPENAI_API_KEY not set")
	}

	llm, err := openai.New(openai.WithToken(apiKey))
	if err != nil {
		return "", 0, err
	}

	prompt := "Classify this review into one of these sentiments: " +
		strings.Join(names, ", ") + ". Review: " + review

	response, err := llm.Call(c, prompt)
	if err != nil {
		return "", 0, err
	}

	for _, r := range rankings {
		if r.RankingName == response {
			return response, r.RankingValue, nil
		}
	}
	return response, 0, nil
}

func GetRankings(client *mongo.Client) ([]models.Ranking, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := database.GetCollection(client, "rankings")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rankings []models.Ranking
	if err := cursor.All(ctx, &rankings); err != nil {
		return nil, err
	}

	return rankings, nil
}

// ========================== RECOMMENDATIONS ==========================

func GetRecommendedMovies(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := utils.GetUserIdFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		genres, err := GetUsersFavouriteGenres(userID, client)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		limit := int64(5)
		if v := os.Getenv("RECOMMENDED_MOVIE_LIMIT"); v != "" {
			limit, _ = strconv.ParseInt(v, 10, 64)
		}

		opts := options.Find().SetSort(bson.M{"ranking.ranking_value": 1}).SetLimit(limit)
		filter := bson.M{"genres.genre_name": bson.M{"$in": genres}}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		collection := database.GetCollection(client, "movies")
		cursor, err := collection.Find(ctx, filter, opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recommendations"})
			return
		}
		defer cursor.Close(ctx)

		var movies []models.Movie
		if err := cursor.All(ctx, &movies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode recommendations"})
			return
		}

		c.JSON(http.StatusOK, movies)
	}
}

func GetUsersFavouriteGenres(userID string, client *mongo.Client) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := database.GetCollection(client, "users")
	var result struct {
		FavouriteGenres []struct {
			GenreName string `bson:"genre_name"`
		} `bson:"favourite_genres"`
	}

	err := collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}

	var genres []string
	for _, g := range result.FavouriteGenres {
		genres = append(genres, g.GenreName)
	}

	return genres, nil
}

// ========================== GENRES ==========================

func GetGenres(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		collection := database.GetCollection(client, "genres")
		cursor, err := collection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch genres"})
			return
		}
		defer cursor.Close(ctx)

		var genres []models.Genre
		if err := cursor.All(ctx, &genres); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode genres"})
			return
		}

		c.JSON(http.StatusOK, genres)
	}
}
