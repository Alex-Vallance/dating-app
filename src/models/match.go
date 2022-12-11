package models

/*
MatchState - type to define states of a relationship
Pending occurs when user1 swipes yes on user2
If user2 swipes yes on user1 their state will be Matched
If either swipes no it will be updated to Unmatched
*/
type MatchState int

const (
	Pending MatchState = iota
	Matched
	Unmatched
)

type Match struct {
	UserID int `json:"userId"`
	MatchID int `json:"matchId"`
	State MatchState `json:"state"`
}
