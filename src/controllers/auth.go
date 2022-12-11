package controllers

import (
	"database/sql"
	"dating-app/src/interactors"
	"dating-app/src/models"
	"github.com/goombaio/namegenerator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/sethvargo/go-password/password"
	"math/rand"
	"net/http"
	"time"
)

type Auth struct {
	db *sql.DB
	authInteractor *interactors.Auth
}

func NewAuth(db *sql.DB) *Auth {
	return &Auth{
		db: db,
		authInteractor:interactors.NewAuth(db),
	}
}

/*
Create - generates a random user with a birthday, name, login creds, gender, and location
when the user is returned, age is calculated from the users date of birth
 */
func (a *Auth) Create (c echo.Context)error{
	newUser := models.User{}

	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)

	newUser.Name = nameGenerator.Generate()
	newUser.Email = newUser.Name + "@gmail.com"

	securePassword, err := password.Generate(16, 4, 4, false, false)
	if err != nil {
		log.Error(err)
		return err
	}

	newUser.Password = securePassword

	newUser.Gender = models.GenderType(rand.Intn(3))

	currentYear := time.Now().Year() - 18
	minYear := currentYear - 47
	randomYear := rand.Intn(currentYear-minYear) + minYear
	randomMonth := rand.Intn(12) + 1
	randomDay := rand.Intn(28) + 1 //skipping days that don't occur every month for simplicity
	dateOfBirth := time.Date(randomYear, time.Month(randomMonth), randomDay, 0, 0, 0, 0,time.UTC)
	newUser.DateOfBirth = dateOfBirth

	newUser.Latitude = float64(rand.Intn(180)) - 90 // to account for negative values
	newUser.Longitude = float64(rand.Intn(360)) - 180 // to account for negative values

	newUser, err = a.authInteractor.Create(newUser)
	if err != nil {
		log.Error(err)
		return err
	}

	return c.JSON(http.StatusOK, newUser)
}

type loginRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

/*
Login - takes the users credentials and if the email/password exists in the DB, generates a token
 */
func (a *Auth) Login (c echo.Context) error {
	request := &loginRequest{}
	if err := c.Bind(request); err != nil {
		return c.JSON(http.StatusBadRequest, nil)
	}

	token, err := a.authInteractor.Login(request.Email, request.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusUnauthorized, "invalid login credentials")
		} else {
			return err
		}
	}

	if token != "" {
		return c.JSON(http.StatusOK, echo.Map{
			"token": token,
		})
	}

	return c.JSON(http.StatusInternalServerError, nil)
}