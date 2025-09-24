package handlers

import (
	redis "authsvc/internal/Redis"
	"authsvc/internal/db"
	"authsvc/internal/keys"
	"authsvc/internal/models"
	"authsvc/internal/tokens"
	"authsvc/internal/users"
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

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

//This handler is called from frontend whenever the access token is expired that time frontent detect the error  402
// and in that time frontend call this handler and generate a new access token after that
// it generate a new refresh token and replaced that in place of old refresh token in redis

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

func RequestReset(c *gin.Context) {

	email, ok := c.Get("email")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized User",
		})
		return
	}

	var user models.User
	verify := db.DB.Model(models.User{}).Where("email=?", email).First(&user)
	if verify.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "db error",
		})
		return
	}
	if verify.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": " user not found",
		})
		return
	}

	AccessToken := c.GetHeader("Authorization")
	if AccessToken == "" {
		c.JSON(401, gin.H{
			"error": "authorization header is missing",
		})
		return
	}
	ctx := context.Background()
	Key := fmt.Sprintf("reset:%s", user.Email)
	err := redis.RDB.Set(ctx, Key, AccessToken, 15*time.Minute).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "redis error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"response": "reset request is accepted validity is 5 min",
	})

}

func VerifyReset(c *gin.Context) {
	var body struct {
		Email       string `json:"email"`
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	ctx := context.Background()
	key := fmt.Sprintf("reset:%s", body.Email)
	storedToken, err := redis.RDB.Get(ctx, key).Result()
	if err == redis.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error1": err.Error()})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error2": err.Error()})
		return
	}

	Token := "Bearer " + body.Token

	if storedToken != Token {
		c.JSON(http.StatusUnauthorized, gin.H{"error": storedToken})
		return
	}

	// ✅ Update password in DB (hash it before saving!)
	hashed, _ := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
	db.DB.Model(&models.User{}).Where("email = ?", body.Email).Update("password", string(hashed))

	// Delete reset token after use
	redis.RDB.Del(ctx, key)

	c.JSON(http.StatusOK, gin.H{"response": "password successfully updated"})
}
