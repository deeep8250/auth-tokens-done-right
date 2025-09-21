package models

type User struct {
	Id           uint `gorm:"primaryKey"`
	Name         string
	Email        string `gorm:"uniqueIndex;not null"`
	Password     string `json:"-"`
	RefreshToken string
}
