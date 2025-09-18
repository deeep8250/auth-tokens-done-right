package handlers

import (
	"authsvc/internal/db"
	"authsvc/internal/models"
	"authsvc/internal/tokens"
	"authsvc/internal/users"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SignIn(c *gin.Context) {

	type SignIn struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	var user SignIn

	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	var count int64
	verify := db.DB.Model(models.User{}).Where("email=?", user.Email).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user already exist",
		})
		return
	}

	if verify.Error != nil {
		log.Println("error : ", verify.Error.Error())
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

func Login(c *gin.Context) {

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

	isSame := users.VerifyPassword(user.Password, login.Password)
	if !isSame {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "password isnt matched",
		})
		return
	}

	token, err := tokens.GenerateJWT(login.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})

}
