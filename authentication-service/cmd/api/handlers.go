package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)

	fmt.Println("inside authenticate")
	fmt.Println(requestPayload)

	if err != nil {
		fmt.Println("inside error check for readJson")
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	//validate the user against the database
	user, err := app.Repo.GetByEmail(requestPayload.Email)
	if err != nil {
		fmt.Println("inside error check for getByEmail")
		app.errorJSON(w, errors.New("user not registered"), http.StatusUnauthorized)
		return
	}

	valid, err := app.Repo.PasswordMatches(requestPayload.Password, *user)

	if err != nil || !valid {
		fmt.Println("inside error check for Password Match")
		app.errorJSON(w, errors.New("invalid password"), http.StatusUnauthorized)
		return
	}

	err = app.logRequest("authentication", fmt.Sprintf("%s logged in ", user.Email))
	if err != nil {
		fmt.Println("inside error check for logRequest")
		app.errorJSON(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logRequest(name, data string) error {
	fmt.Println("inside logRequest")
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("inside error check for newRequest")
		return err
	}

	_, err = app.Client.Do(request)
	if err != nil {
		fmt.Println("inside error check for Client.Do")
		return err
	}

	return nil
}
