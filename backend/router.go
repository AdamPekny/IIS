package api

import (
	"net/http"

	"github.com/AdamPekny/IIS/backend/middleware"
	"github.com/AdamPekny/IIS/backend/models"
	"github.com/AdamPekny/IIS/backend/utils"
	"github.com/AdamPekny/IIS/backend/views"
	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {
	router := gin.Default()

	router.GET("/api/users/list", middleware.RequireAuth("admin"), func(ctx *gin.Context) {
		var users []models.User
		utils.DB.Find(&users)
		ctx.IndentedJSON(http.StatusOK, users)
	})

	router.POST("/api/users/signup", views.Signup)
	router.POST("/api/users/login", views.Login)

	router.GET("/api/vehicles/list", views.List_vehicles)

	router.POST("/api/vehicles/create", views.Create_vehicle)

	router.GET("/api/connections", views.ListConnections)
	router.GET("/api/connections/:id", views.GetConnection)
	return router
}
