package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"io"
	"vk_march_backend/internal/handlers"
	"vk_march_backend/pkg/config"
	"vk_march_backend/pkg/database"
	"vk_march_backend/pkg/logger"
)

func main() {
	cfg := config.GetConfig()
	db := database.Init(cfg)

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Log.Errorln("Error closing database: " + err.Error())
		}
	}(db)

	logger.Log.Infoln("Starting service...")
	router := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard

	logger.Log.Infoln("Serving handlers...")

	router.GET("/login", func(c *gin.Context) {
		handlers.Login(c, db, cfg)
	})
	router.POST("/register", func(c *gin.Context) {
		handlers.Register(c, db)
	})
	router.POST("/postAd", func(c *gin.Context) {
		handlers.PlaceAd(c, db, cfg)
	})
	router.GET("/ads", func(c *gin.Context) {
		handlers.GetAds(c, db, cfg)
	})

	logger.Log.Info("Starting router...")
	logger.Log.Info("On port :" + cfg.Listen.Port)
	logger.Log.Fatal(router.Run(":" + cfg.Listen.Port))

}
