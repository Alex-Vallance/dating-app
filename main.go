package main

import (
	"database/sql"
	"dating-app/src/controllers"
	"dating-app/src/models"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

/*
main
	connect to db
	create database tables
	initialise controllers with db access (to init interactors)
	specify routes
	start server
*/
func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Connect to the database with the name of the database container and it's login details.
	fmt.Println("Connecting to db")
	conn, err := sql.Open("mysql", "root:mypassword@tcp(db:3306)/testdb")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// MySQL server isn't fully active yet.
	// Block until connection is accepted. This is a docker problem with v3 & container doesn't start
	// up in time.
	for conn.Ping() != nil {
		fmt.Println("Attempting connection to db")
		time.Sleep(5 * time.Second)
	}
	fmt.Println("Connected")

	err = migrateTable(conn)
	if err != nil {
		log.Fatal(err)
	}

	readTokenForRequest := readToken(models.JWT_KEY, "token")

	auth := controllers.NewAuth(conn)
	e.POST("/user/create", auth.Create)
	e.POST("/login", auth.Login)

	match := controllers.NewMatch(conn)
	e.GET("/profiles", match.Profiles, readTokenForRequest)
	e.POST("/swipe", match.Swipe, readTokenForRequest)

	e.GET("/health", healthCheck)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}

func healthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "")
}

/*
readToken - read JWT token from request so we can pull the user id in the required functions
*/
func readToken(signingKey []byte, contextKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			token, err := parse(c, signingKey)
			if token != nil && err == nil {
				c.Set(contextKey, token)
			}

			return next(c)
		}
	}
}

var jwtRegex = regexp.MustCompile("^Bearer\\s+(.*)$")

func parse(c echo.Context, key []byte) (*jwt.Token, error) {
	tokenString := getTokenStringFromHeader(c)
	if tokenString == "" {
		return nil, fmt.Errorf("Token not present")
	}

	return parseTokenString(tokenString, key)
}

func getTokenStringFromHeader(c echo.Context) string {
	header := c.Request().Header.Get(echo.HeaderAuthorization)
	if jwtRegex.MatchString(header) {
		return strings.TrimSpace(jwtRegex.FindStringSubmatch(header)[1])
	}

	return ""
}

func parseTokenString(tokenString string, key []byte) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != "HS256" {
			return nil, fmt.Errorf("Invalid jwt signing method")
		}
		return key, nil
	})
}

func migrateTable(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS users
(
	id int auto_increment,
	email varchar(255) NOT NULL,
	password varchar(255) NOT NULL,
	name varchar(255) NOT NULL,
	gender int NOT NULL DEFAULT 2,
	date_of_birth datetime,
	latitude int,
	longitude int,
	likability int NOT NULL DEFAULT 0,
	PRIMARY KEY (id)
);`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS matches
(
	user_id int,
	match_user_id int,
	state int default 0
);`)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}