package interactors

import (
	"database/sql"
	"dating-app/src/models"
	"github.com/golang-jwt/jwt"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"math"
	"time"
)

type Auth struct {
	db *sql.DB
}

func NewAuth(db *sql.DB) *Auth {
	return &Auth{
		db: db,
	}
}

/*
GetUserByID - returns a user by the provided id
*/
func (a *Auth) GetUserByID (userID int) *models.User {
	relationshipQuery := `SELECT * FROM users WHERE id = ?;`

	row := a.db.QueryRow(relationshipQuery, userID)
	user := new(models.User)
	row.Scan(&user.ID, &user.Email, &user.Password, &user.Name, &user.Gender, &user.DateOfBirth, &user.Latitude, &user.Longitude)
	return user
}

/*
Create - add a new user row in the db with the random data generated
*/
func (a *Auth) Create(user models.User) (models.User, error) {
	result, err := a.db.Exec("INSERT INTO users (email, password, name, gender, date_of_birth, latitude, longitude) VALUES (?,?,?,?,?,?,?)",
		user.Email, user.Password, user.Name, user.Gender, user.DateOfBirth, user.Latitude, user.Longitude)
	if err != nil {
		return user, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return user, err
	}

	user.ID = int(id)
	user.Age = int(math.Floor(time.Since(user.DateOfBirth).Hours() / 24 / 365))

	return user, nil
}

/*
Login - check the provided credentials. If they are correct generate a token
*/
func (a *Auth) Login(email, password string) (string, error) {
	userQuery := `SELECT id, email, password FROM users 
WHERE email = ? AND password = ?`
	row := a.db.QueryRow(userQuery, email, password)
	user := new(models.User)
	err := row.Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		return "", err
	}

	expirationTime := time.Now().Add(60 * time.Minute)

	claims := &models.Claims{
		UserID: user.ID,
		RegisteredClaims: jwtv4.RegisteredClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: jwtv4.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString(models.JWT_KEY)
	if err != nil {
		return "", err
	}
	return tokenString, err
}

/*
GetUserFromRequest - read the header token to get the user's id
*/
func (a *Auth) GetUserFromRequest (c echo.Context) int {
	token := c.Get("token")
	if token == nil {
		return -1
	}
	user := token.(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	if val, ok := claims["user_id"]; ok {
		floatUserId := val.(float64)
		return int(floatUserId)
	}
	return -1
}

/*
UpdateUserLikabilityScore - increasing the likability for the provided user by the modifier
*/
func (a *Auth) UpdateUserLikabilityScore(userID, modifier int) error {
	_, err := a.db.Exec("UPDATE users set likability = likability + (?) WHERE id = ? ", modifier, userID)
	if err != nil {
		return err
	}

	return nil
}