package handlers

import (
	redis "authsvc/internal/Redis"
	"authsvc/internal/db"
	"authsvc/internal/keys"
	"authsvc/internal/models"
	"authsvc/internal/tokens"
	"authsvc/internal/users"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func SignIn(c *gin.Context) {

	type SignIn struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	var user SignIn

	//taking data from user response
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	//check if user already exist or not
	var count int64
	verify := db.DB.Model(&models.User{}).Where("email=?", user.Email).Count(&count)

	if verify.Error != nil {
		log.Println("error : ", verify.Error.Error())
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user already exist",
		})
		return
	}

	//hasing
	hash, err := users.HashPassword(user.Password)
	if err != nil {
		log.Println("error :", err.Error())
		return
	}

	DbUser := models.User{
		Name:     user.Name,
		Email:    user.Email,
		Password: string(hash),
	}

	if err := db.DB.Create(&DbUser).Error; err != nil {
		log.Println("error while creating user:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal error while creating user",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    DbUser.Id,
		"email": DbUser.Email,
		"name":  DbUser.Name,
	})
}

func Login(key *keys.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {

		type Login struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=6"`
		}

		var login Login
		err := c.ShouldBindJSON(&login)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		var user models.User
		verify := db.DB.Where("email=?", login.Email).First(&user)
		if verify.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": verify.Error.Error(),
			})
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"err": err.Error(),
			})
			return
		}

		token, err := tokens.GenerateJWT(login.Email, key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		var refreshToken string
		if user.RefreshToken != "" {
			refreshToken = user.RefreshToken
		} else {

			refreshToken, _ = tokens.GenerateRefreshToken()

			err := redis.StoreRefreshToken(user.Id, refreshToken)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
		}

		db.DB.Save(&user)

		// 4. Return both tokens
		c.JSON(http.StatusOK, gin.H{
			"access_token":  token,
			"refresh_token": refreshToken,
		})

	}
}

func Profile(c *gin.Context) {
	email, ok := c.Get("email")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Error": "unauthorized user",
		})
		return
	}

	var user models.User
	verify := db.DB.Where("email=?", email).First(&user)
	if verify.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": verify.Error.Error(),
		})
		return
	}

	if verify.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"err": "user not exist",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})

}

func Refresh(c *gin.Context, keys *keys.Manager) {

	type RefreshToken struct {
		Token string `json:"refreshToken"`
	}

	var token RefreshToken
	err := c.ShouldBindJSON(&token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if token.Token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "token isnt found in the response",
		})
		return
	}

	var user models.User
	verify := db.DB.Model(&models.User{}).Where("refresh_token=?", token.Token).First(&user)
	if verify.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": verify.Error.Error(),
		})
		return
	}

	if verify.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "not found",
		})
		return
	}

	// convert uint/uint64 to string

	key := strconv.FormatUint(uint64(user.Id), 10)
	storedToken, err := redis.RDB.Get(redis.Ctx, string(key)).Result()
	if err != nil || storedToken != token.Token {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	//generate access token
	accessToken, err := tokens.GenerateJWT(user.Email, keys)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return

	}
	// generate refresh token
	Rtoken, err := tokens.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	user.RefreshToken = Rtoken
	result := db.DB.Save(&user)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"refresh token": Rtoken,
		"access token":  accessToken,
	})

}
