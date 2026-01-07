package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/samrato/magicstream/controllers"
	"github.com/samrato/magicstream/middleware"
	"go.mongodb.org/mongo-driver/mongo"
)

func UserRoutes(router *gin.Engine, client *mongo.Client) {
	// ================= PUBLIC ROUTES =================
	public := router.Group("/users")
	{
		public.POST("/register", controllers.RegisterUser(client))
		public.POST("/login", controllers.LoginUser(client))
		public.POST("/refresh-token", controllers.RefreshTokenHandler())
	}

	// ================= AUTHENTICATED ROUTES =================
	auth := router.Group("/users")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.GET("/profile", controllers.GetUserProfile(client))
		auth.PUT("/favourite-genres", controllers.UpdateFavouriteGenres(client))
		auth.POST("/logout", controllers.LogoutHandler())
	}

	// ================= ADMIN ROUTES =================
	// Example: if you want admin-only user actions in future
	admin := router.Group("/admin/users")
	admin.Use(
		middleware.AuthMiddleware(),
		middleware.AdminOnly(),
	)
	{
		// e.g., get all users, delete users, etc.
		// admin.GET("/", controllers.GetAllUsers(client))
	}
}
