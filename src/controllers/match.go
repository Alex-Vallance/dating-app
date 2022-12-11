package controllers

import (
	"database/sql"
	"dating-app/src/interactors"
	"dating-app/src/models"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"math"
	"net/http"
	"sort"
	"time"
)

type Match struct {
	authInteractor *interactors.Auth
	matchInteractor *interactors.Match
}

func NewMatch(db *sql.DB) *Match {
	return &Match{
		authInteractor: interactors.NewAuth(db),
		matchInteractor: interactors.NewMatch(db),
	}
}

type getProfilesRequest struct {
	AgeMin int `json:"age_min"`
	AgeMax int `json:"age_max"`
	Gender string `json:"gender"`
	Sort string `json:"sort"`
}
/*
Profiles - returns potential matches for the requesting user
distance is calculated relative to the requesting user.
In order to achieve this, we retrieve the users first, then calculate distance for each
 */
func (m *Match) Profiles (c echo.Context) error {
	userID := m.authInteractor.GetUserFromRequest(c)
	if userID < 1 {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorised access")
	}

	request := &getProfilesRequest{}
	if err := c.Bind(request); err != nil || request.AgeMax < request.AgeMin {
		return c.JSON(http.StatusBadRequest, nil)
	}

	ageMin := time.Time{}
	ageMax := time.Time{}
	currentTime := time.Now()

	// convert number of years to date for db comparison
	if request.AgeMin > 18 { // 18 is the minimum age
		ageMin = currentTime.AddDate(-1 * request.AgeMin, 0 , 0)
	}

	// convert number of years to date for db comparison
	if request.AgeMax > 18 && request.AgeMax < 65 { // 65 is the maximum age
		ageMax = currentTime.AddDate(-1 * request.AgeMax, 0 , 0)
	}

	filterOpts := interactors.FilterOpts{
		AgeMin: &ageMin,
		AgeMax: &ageMax,
		Gender: models.ToGenderTypeFromString(request.Gender),
		SortRecommended: false,
	}

	if request.Sort == "recommended" {
		filterOpts.SortRecommended = true
	}

	profiles, err := m.matchInteractor.GetProfilesForUser(userID, filterOpts)
	if err != nil {
		log.Error(err)
		return err
	}

	requestingUser := m.authInteractor.GetUserByID(userID)

	for _, profile := range profiles {
		profile.Distance = distance(requestingUser.Latitude, requestingUser.Longitude, profile.Latitude, profile.Longitude)
	}

	if request.Sort == "distance" {
		sort.SliceStable(profiles, func(i, j int) bool {
			return *profiles[i].Distance < *profiles[j].Distance
		})
	}

	return c.JSON(http.StatusOK, profiles)
}

/*
	distance is a helper function for calculating the distance of profiles from the requesting user using latitude and longitude
	this was a formula I found online to calculate distance between lat and long points
*/
func distance(lat1 float64, lng1 float64, lat2 float64, lng2 float64) *float64 {
	const PI float64 = 3.141592653589793

	radLat1 := PI * lat1 / 180
	radLat2 := PI * lat2 / 180

	theta := lng1 - lng2
	radTheta := PI * theta / 180

	dist := math.Sin(radLat1) * math.Sin(radLat2) + math.Cos(radLat1) * math.Cos(radLat2) * math.Cos(radTheta)

	if dist > 1 {
		dist = 1
	}

	dist = math.Acos(dist)
	dist = dist * 180 / PI
	dist = dist * 60 * 1.1515

	return &dist
}

type swipeRequest struct {
	ProfileID int `json:"profile_id"`
	Preference string `json:"preference"`
}

type swipeResponse struct {
	Matched bool `json:"matched"`
	MatchID *int `json:"match_id,omitempty"`
}

/*
Swipe - requesting user can swipe yes or no on the specified user
 	request is rejected if unauthorised, missing profile or incorrect value for preference
 	if the users action has already been completed they will get a conflict error status
		ie a user swiping right twice
	a user is able to change their swipe at any time (this would allow for rematch or unmatching)
	if the Swipe action is completed, the user receiving the swipe will get an updated likeability score (used in filtering profile results)
*/
func (m *Match) Swipe (c echo.Context) error {
	userID := m.authInteractor.GetUserFromRequest(c)
	if userID < 1 {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorised access")
	}
	request := &swipeRequest{}
	if err := c.Bind(request); err != nil || request.ProfileID < 1 || request.Preference == "" || (request.Preference != "YES" && request.Preference != "NO") {
		return c.JSON(http.StatusBadRequest, nil)
	}

	currentMatch, err := m.matchInteractor.GetRelationship(userID, request.ProfileID)
	if err != nil {
		log.Error(err)
		return err
	}

	if currentMatch == nil {
		// create new relationship
		preference := models.Pending
		if request.Preference == "NO" {
			preference = models.Unmatched
		}
		currentMatch = &models.Match{
			UserID:  userID,
			MatchID: request.ProfileID,
			State:   preference,
		}

		err = m.matchInteractor.Create(currentMatch)
		if err != nil {
			log.Error(err)
			return err
		}
	} else {
		// user has already swiped
		if userID == currentMatch.UserID &&
			(currentMatch.State != models.Unmatched && request.Preference == "YES") ||
			(currentMatch.State == models.Unmatched && request.Preference == "NO") ||
			(request.Preference == "YES" && currentMatch.State == models.Matched) {
			return c.JSON(http.StatusConflict, "already swiped this profile")
		}

		needsUpdate := false

		// update existing relationship
		if request.Preference == "NO" {
			currentMatch.State = models.Unmatched
			needsUpdate = true
		}

		if request.Preference == "YES" && currentMatch.State == models.Pending {
			currentMatch.State = models.Matched
			needsUpdate = true
		}

		if needsUpdate {
			err = m.matchInteractor.Update(currentMatch)
			if err != nil {
				log.Error(err)
				return err
			}
		}
	}

	if request.Preference == "YES" {
		err = m.authInteractor.UpdateUserLikabilityScore(request.ProfileID, 1)
		if err != nil {
			log.Error(err)
			return err
		}
	} else if request.Preference == "NO" {
		err = m.authInteractor.UpdateUserLikabilityScore(request.ProfileID, -1)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	response := swipeResponse{
		Matched: false,
		MatchID: nil,
	}
	if currentMatch.State == models.Matched {
		response.Matched = true
		response.MatchID = &request.ProfileID
	}

	return c.JSON(http.StatusOK, response)
}
