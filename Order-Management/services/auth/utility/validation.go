package utility

import "strings"

func ValidCredentials(username, password string) bool {
	return strings.TrimSpace(username) != "" &&
		strings.TrimSpace(password) != ""
}

func ValidRole(role string) bool {
	return role == "admin" || role == "user"
}
