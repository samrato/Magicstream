package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/samrato/magicstream/controllers"
	"github.com/samrato/magicstream/middleware"
	"go.mongodb.org/mongo-driver/mongo"
)

func MovieRoutes(router *gin.Engine, client *mongo.Client) {
	// ================= PUBLIC ROUTES =================
	router.GET("/movies", controllers.GetMovies(client))
	router.GET("/movies/:imdb_id", controllers.GetMovie(client))
	router.GET("/movies/recommended", controllers.GetRecommendedMovies(client))
	router.GET("/genres", controllers.GetGenres(client))

	// ================= AUTHENTICATED ROUTES =================
	auth := router.Group("/")
	auth.Use(middleware.AuthMiddleware()) // require login
	{
		auth.POST("/movies", controllers.AddMovie(client))
	}

	// ================= ADMIN ROUTES =================
	admin := router.Group("/admin")
	admin.Use(
		middleware.AuthMiddleware(),  
		middleware.AdminOnly(),       
	)
	{
		admin.PUT("/movies/:imdb_id/review", controllers.AdminReviewUpdate(client))
	}
}
