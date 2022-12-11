package models

import (
	"encoding/json"
	"github.com/golang-jwt/jwt/v4"
	"strings"
	"time"
)

/*
GenderType - represents the possible values for gender
Will be stored in the db as an int and read to be returned as a string
Users have the right not to specify their gender but this may affect how they are returned in results
*/
type GenderType int

const (
	Male GenderType = iota
	Female
	NotSpecified
)

func ToGenderTypeFromString(s string) GenderType {
	switch s {
	case "Male":
		return Male
	case "Female":
		return Female
	default:
		return NotSpecified
	}
}

func (t GenderType) String() string {
	switch t {
	case Male:
		return "Male"
	case Female:
		return "Female"
	default:
		return "Not Specified"
	}
}

func (t GenderType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func (t *GenderType) UnmarshalJSON(b []byte) error {
	var gender string
	err := json.Unmarshal(b, &gender)
	if err != nil {
		return err
	}

	switch strings.ToLower(gender) {
	case "male":
		*t = Male
	case "female":
		*t = Female
	default:
		*t = NotSpecified
	}

	return nil
}

/*
	User - struct is used for creating users.
	This is because we don't want to return login credentials when retrieving profiles
*/
type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Profile
}

/*
	Profile - holds all non-sensitive user information. Used when getting profiles
*/
type Profile struct {
	ID       int `json:"id"`
	Name     string `json:"name"`
	Gender   GenderType `json:"gender"`
	DateOfBirth      time.Time `json:"-"`
	Age      int `json:"age"`
	Latitude float64 `json:"-"`
	Longitude float64 `json:"-"`
	Distance *float64 `json:"distance,omitempty"`
	LikabilityScore *int `json:"likability,omitempty"`
}

/*
 Used for generating tokens
*/
var JWT_KEY = []byte("my_secret_key")

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}