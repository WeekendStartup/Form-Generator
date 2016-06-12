package main

import "encoding/json"

func getFormModel() (string, error) {
	var questions []Question
	usernameQuestion := Question{
		Name:        "username",
		Type:        "text",
		Required:    true,
		Placeholder: "Type your username",
	}
	questions = append(questions, usernameQuestion)

	passwordQuestion := Question{
		Name:        "password",
		Type:        "text",
		Required:    true,
		Placeholder: "******",
	}
	questions = append(questions, passwordQuestion)

	roleOption := Option{
		Name:  "user",
		Value: "user",
	}
	roleOption2 := Option{
		Name:  "admin",
		Value: "admin",
	}
	roleQuestion := Question{
		Name:     "role",
		Type:     "dropdown",
		Required: false,
		Options:  []Option{roleOption, roleOption2},
	}
	questions = append(questions, roleQuestion)

	b, err := json.Marshal(questions)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
