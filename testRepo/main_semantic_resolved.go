package testrepo

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

func (u *User) Validate() error {
	if len(u.Name) < 2 {
		return fmt.Errorf("name must be at least 2 characters")
	}
	if u.Email == "" || !strings.Contains(u.Email, "@") {
		return fmt.Errorf("email must be valid")
	}
	return nil
}

func main() {
	user := &User{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
	}
	
	if err := user.Validate(); err != nil {
		log.Printf("Validation error: %v", err)
		return
	}
	
	fmt.Printf("User %s created successfully\n", user.Name)
}

func GetUserByID(id int) (*User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}
	// TODO: implement database lookup
	return nil, fmt.Errorf("not implemented")
}