package models

import (
	"github.com/hamza/proglabodev3/api/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleDoctor  Role = "doctor"
	RolePatient Role = "patient"
)

type User struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email      string             `json:"email" bson:"email"`
	Password   string             `json:"-" bson:"password"`
	FirstName  string             `json:"firstName" bson:"firstName"`
	LastName   string             `json:"lastName" bson:"lastName"`
	Role       Role               `json:"role" bson:"role"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
	Expertises []Expertise        `json:"expertises" bson:"expertises"`
	WorkPlan   utils.PeriodicPlan `json:"workPlan" bson:"workPlan"`
}

type Expertise struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
}

func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}
