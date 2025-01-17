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
	Expertises []string           `json:"expertises" bson:"expertises"`
	WorkPlan   utils.PeriodicPlan `json:"workPlan" bson:"workPlan"`
}

type Expertise struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
}

func (u *User) HashPassword() error {
	var err error
	u.Password, err = HashStr(u.Password)
	return err
}

func HashStr(str string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(str), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func (u *User) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}
