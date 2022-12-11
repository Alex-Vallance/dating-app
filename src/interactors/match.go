package interactors

import (
	"database/sql"
	"dating-app/src/models"
	"errors"
	"fmt"
	"math"
	"time"
)

type Match struct {
	db *sql.DB
}

func NewMatch(db *sql.DB) *Match {
	return &Match{
		db: db,
	}
}

type FilterOpts struct {
	AgeMin *time.Time
	AgeMax *time.Time
	Gender models.GenderType
	SortRecommended bool
}
/*
GetProfilesForUser - gets the profiles for a requesting user within filtering options
convert date of birth to age
*/
func (m *Match) GetProfilesForUser (userID int, opts FilterOpts) ([]*models.Profile, error) {
	profileQuery := `SELECT id, name, gender, date_of_birth, latitude, longitude, likability FROM users 
WHERE
id NOT IN (SELECT match_user_id FROM matches WHERE user_id = ?)
AND id NOT IN (SELECT user_id FROM matches WHERE match_user_id = ? AND state != 0)
AND id != ?`

	if !opts.AgeMin.IsZero() {
		profileQuery += fmt.Sprintf(" AND date_of_birth < '%s'", opts.AgeMin)
	}

	if !opts.AgeMax.IsZero() {
		profileQuery += fmt.Sprintf(" AND date_of_birth > '%s'", opts.AgeMax)
	}

	if opts.Gender != models.NotSpecified {
		profileQuery += fmt.Sprintf(" AND gender = %d", opts.Gender)
	}

	if opts.SortRecommended {
		profileQuery += " ORDER BY likability DESC"
	}

	rows, err := m.db.Query(profileQuery, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var profiles []*models.Profile

	for rows.Next() {
		profile := new(models.Profile)

		var dateOfBirth string

		err = rows.Scan(&profile.ID, &profile.Name, &profile.Gender, &dateOfBirth, &profile.Latitude, &profile.Longitude, &profile.LikabilityScore)
		if err != nil {
			return nil, err
		}
		profile.DateOfBirth, err = time.Parse("2006-01-02 15:04:05", dateOfBirth)
		if err != nil {
			return nil, err
		}
		profile.Age = int(math.Floor(time.Since(profile.DateOfBirth).Hours() / 24 / 365))

		profiles = append(profiles, profile)
	}
	return profiles, nil
}

/*
GetRelationship - gets the current match status between two users.
This allows for swiping back if there is a pending match
*/
func (m *Match) GetRelationship (userID, profileID int) (*models.Match, error) {
	relationshipQuery := `SELECT * FROM matches 
WHERE (user_id = ? AND match_user_id = ?)
OR (user_id = ? AND match_user_id = ?);`

	row := m.db.QueryRow(relationshipQuery, userID, profileID, profileID, userID)
	relationship := new(models.Match)
	err := row.Scan(&relationship.UserID, &relationship.MatchID, &relationship.State)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return relationship, nil
}

/*
Create - new match when a user is swiped and no current relationship exists
*/
func (m *Match) Create(newMatch *models.Match) error {
	_, err := m.db.Exec("INSERT INTO matches (user_id, match_user_id, state) VALUES (?,?,?)", newMatch.UserID, newMatch.MatchID, newMatch.State)
	if err != nil {
		return err
	}

	return nil
}

/*
Update - existing relationship state change
ie pending to matched or unmatched
*/
func (m *Match) Update(newMatch *models.Match) error {
	_, err := m.db.Exec("UPDATE matches set state = ? WHERE user_id = ? AND match_user_id = ?", newMatch.State, newMatch.UserID, newMatch.MatchID)
	if err != nil {
		return err
	}

	return nil
}
