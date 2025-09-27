#!/bin/zsh

sed -i '/<<<<<<< HEAD/,/>>>>>>> feature\/user-validation/c\
func (u *User) Validate() error {\n\tif len(u.Name) < 2 {\n\t\treturn fmt.Errorf("name must be at least 2 characters")\n\t}\n\tif u.Email == "" || !strings.Contains(u.Email, "@") {\n\t\treturn fmt.Errorf("email must be valid")\n\t}\n\treturn nil\n}' testRepo/main.go
sed -i '/<<<<<<< HEAD/,/>>>>>>> feature\/improved-logging/c\
func main() {\n\tuser := &User{\n\t\tID:    1,\n\t\tName:  "John Doe",\n\t\tEmail: "john@example.com",\n\t}\n\t\n\tif err := user.Validate(); err != nil {\n\t\tlog.Printf("Validation error: %v", err)\n\t\treturn\n\t}\n\t\n\tfmt.Printf("User %s created successfully\\n", user.Name)\n}' testRepo/main.go
sed -i '/<<<<<<< HEAD/,/>>>>>>> feature\/input-validation/c\
func GetUserByID(id int) (*User, error) {\n\tif id <= 0 {\n\t\treturn nil, fmt.Errorf("invalid user ID")\n\t}\n\t// TODO: implement database lookup\n\treturn nil, fmt.Errorf("not implemented")\n}' testRepo/main.go
sed: 1: "testRepo/main.go": undefined label 'estRepo/main.go'
sed: 1: "testRepo/main.go": undefined label 'estRepo/main.go'
sed: 1: "testRepo/main.go": undefined label 'estRepo/main.go'