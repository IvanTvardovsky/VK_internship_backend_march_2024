package handlers

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
	"vk_march_backend/internal/middleware"
	"vk_march_backend/internal/structures"
	"vk_march_backend/internal/utils"
	"vk_march_backend/pkg/logger"
	"vk_march_backend/pkg/tkn"
)

const maxSize = int64(1024 * 1024) // 1 мегабайт

func Login(c *gin.Context, db *sql.DB, cfg *structures.Config) {
	authHeader := c.GetHeader("Authorization")
	isAuthenticated := false

	if authHeader != "" {
		if tokenInfo, err := middleware.VerifyAndGetInfoFromToken(c, cfg); err == nil {
			if err == nil {
				expirationTime := time.Unix(int64(tokenInfo.Expires), 0)
				currentTime := time.Now()

				logger.Log.Traceln("Logged by token")
				if expirationTime.After(currentTime) {
					isAuthenticated = true
					c.Header("Authorization", authHeader)
					c.JSON(http.StatusOK, gin.H{"message": "Login by token successful"})
				}
			} else {
				logger.Log.Traceln(err)
			}
		}
	}

	if !isAuthenticated {
		var user structures.User

		err := c.ShouldBindJSON(&user)
		if err != nil {
			logger.Log.Errorln("Error unmarshalling JSON: " + err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		logger.Log.Infoln(user.Login, user.Password)

		var databaseUser structures.User
		err = db.QueryRow("SELECT user_id, password_hash FROM users WHERE login = $1", user.Login).Scan(&databaseUser.ID, &databaseUser.PasswordHash)
		if err != nil {
			logger.Log.Errorln("Error querying database: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		if tkn.CheckPasswordHash(user.Password, databaseUser.PasswordHash) == false {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		user.ID = databaseUser.ID
		token, err := tkn.GenerateToken(&user, cfg)
		if err != nil {
			logger.Log.Errorln("Error while generating token: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.Header("Authorization", "Bearer "+token)
		c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
	}
}

func Register(c *gin.Context, db *sql.DB) {
	var user structures.User

	err := c.ShouldBindJSON(&user)
	if err != nil {
		logger.Log.Errorln("Error unmarshalling JSON: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(user.Login) > 16 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Login must be less than 16 symbols"})
		return
	}

	if len(user.Password) > 32 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be less than 32 symbols"})
		return
	}

	res, err := utils.IsLoginTaken(user.Login, db)
	if err != nil {
		logger.Log.Errorln("Error querying database: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	if res == true {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is already taken"})
		return
	}

	user.PasswordHash, err = tkn.HashPassword(user.Password)

	err = db.QueryRow("INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING user_id", user.Login, user.PasswordHash).Scan(&user.ID)
	if err != nil {
		logger.Log.Errorln("Error inserting data:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"login": user.Login,
	})
}

func PlaceAd(c *gin.Context, db *sql.DB, cfg *structures.Config) {
	tokenInfo, err := middleware.VerifyAndGetInfoFromToken(c, cfg)
	if err != nil {
		switch err.Error() {
		case "token signature is invalid":
			c.JSON(http.StatusBadRequest, gin.H{"error": "token signature is invalid"})
			return
		case "authorization header is missing":
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is missing"})
			return
		case "invalid authorization header format":
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid authorization header format"})
			return
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "error verifying token"})
			return
		}
	}

	var ad structures.Ad
	err = c.ShouldBindJSON(&ad)
	if err != nil {
		logger.Log.Errorln("Error unmarshalling JSON: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(ad.ImageAddress) > 255 || len(ad.Description) > 511 || len(ad.Title) > 31 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "too big length of title/image address/description"})
		return
	}

	if ad.Price < 0 || ad.Price > 10_000_000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "incorrect  price"})
		return
	}

	isImage, err := utils.IsImageURLWithSizeLimit(ad.ImageAddress, maxSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error while checking url"})
		return
	}

	if !isImage {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image address is not image or its size is bigger than 1MB"})
		return
	}

	err = db.QueryRow("INSERT INTO advertisements (user_id, title, description, image_address, price) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		tokenInfo.ID, ad.Title, ad.Description, ad.ImageAddress, ad.Price).Scan(&ad.ID)
	if err != nil {
		logger.Log.Errorln("Error inserting data:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            ad.ID,
		"user_id":       tokenInfo.ID,
		"title":         ad.Title,
		"description":   ad.Description,
		"image_address": ad.ImageAddress,
		"price":         ad.Price,
	})
}

func GetAds(c *gin.Context, db *sql.DB, cfg *structures.Config) {
	tokenInfo, err := middleware.VerifyAndGetInfoFromToken(c, cfg)

	pageStr := c.Query("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "DESC")

	minPriceStr := c.Query("min_price")
	minPrice, err := strconv.Atoi(minPriceStr)
	if err != nil {
		minPrice = 0
	}

	maxPriceStr := c.Query("max_price")
	maxPrice, err := strconv.Atoi(maxPriceStr)
	if err != nil {
		maxPrice = 10_000_001
	}

	limit := 5
	offset := (page - 1) * limit

	query := `
		SELECT id, user_id, title, description, image_address, price, created_at
		FROM advertisements
		WHERE price >= $1 AND price <= $2
		ORDER BY %s %s
		LIMIT $3 OFFSET $4;
	`

	rows, err := db.Query(fmt.Sprintf(query, sortBy, sortOrder), minPrice, maxPrice, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var ads []structures.Ad
	for rows.Next() {
		var ad structures.Ad
		err := rows.Scan(&ad.ID, &ad.UserID, &ad.Title, &ad.Description, &ad.ImageAddress, &ad.Price, &ad.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		userLogin, err := utils.GetUserLoginByID(db, ad.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ad.AuthorLogin = userLogin

		if ad.UserID == tokenInfo.ID {
			ad.IsOwner = true
		}

		ads = append(ads, ad)
	}

	c.JSON(http.StatusOK, ads)
}
