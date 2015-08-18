package config

import (
	"fmt"
	"strings"
)

// AuthVar implements the Flag.Value interface and allows the user to specify
// authentication in the username[:password] form.
type AuthVar Auth

func (a *AuthVar) Set(value string) error {
	a.Enabled = true

	if strings.Contains(value, ":") {
		split := strings.SplitN(value, ":", 2)
		a.Username = split[0]
		a.Password = split[1]
	} else {
		a.Username = value
	}

	return nil
}

func (a *AuthVar) String() string {
	if a.Password == "" {
		return a.Username
	}

	return fmt.Sprintf("%s:%s", a.Username, a.Password)
}
