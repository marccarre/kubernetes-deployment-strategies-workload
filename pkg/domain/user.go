package domain

import (
	"encoding/json"
	"fmt"
)

// User encapsulates data about an user, expose related behaviour, and specifies how to serialise/deserialise the corresponding object.
type User struct {
	ID         int    `json:"id,omitempty"`
	FirstName  string `json:"firstName"`
	FamilyName string `json:"familyName"`
	Age        int    `json:"age"`
}

// FullName returns this user's full name.
func (u User) FullName() string {
	return fmt.Sprintf("%v %v", u.FirstName, u.FamilyName)
}

// Marshal serialises this user as JSON.
func (u User) Marshal() ([]byte, error) {
	return json.Marshal(u)
}

var blankUser = User{}

// UnmarshalUser deserialises the provided JSON into an user object.
func UnmarshalUser(jsonBytes []byte) (*User, error) {
	user := &User{}
	if err := json.Unmarshal(jsonBytes, user); err != nil {
		return nil, err
	}
	if *user == blankUser {
		return nil, fmt.Errorf("invalid JSON: doesn't yield a valid user: %v", string(jsonBytes))
	}
	return user, nil
}
