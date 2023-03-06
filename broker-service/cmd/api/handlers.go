package main

import (
	"broker/event"
	"broker/logs"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type MailPayload struct {
	Fom     string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

// handler functions require response writer
func (app *Config) Broker(writer http.ResponseWriter, request *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = app.writeJSON(writer, http.StatusOK, payload)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	fmt.Println("inside handle submission")
	fmt.Println(requestPayload)

	switch requestPayload.Action {
	case "auth":
		fmt.Println("inside auth")
		fmt.Println(requestPayload.Auth)
		app.authenticate(w, requestPayload.Auth)
	case "log":
		//app.logItem(w, requestPayload.Log)
		//app.logRabbitEvent(w, requestPayload.Log)
		app.logRPCEvent(w, requestPayload.Log)
	case "mail":
		app.sendMail(w, requestPayload.Mail)
	default:
		app.errorJSON(w, errors.New("unknown action"))
	}
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	//create json we'll send to auth service
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	fmt.Println("inside authenticate")
	fmt.Println(jsonData)

	//call service
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	//handle error when request to auth service fails
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	defer response.Body.Close()

	// make sure we get correct status code
	if response.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("invalid credentials"), response.StatusCode)
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling auth service"), response.StatusCode)
		return
	}

	var jsonFromService jsonResponse
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	//handle error when json response cannot be decoded
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// handle scenario when response body has error=true (unlikely since error=true means status !=OK so why repeat?)
	if jsonFromService.Error {
		app.errorJSON(w, err, http.StatusUnauthorized)
	}

	app.writeJSON(w, response.StatusCode, jsonFromService)
}

func (app *Config) logItem(w http.ResponseWriter, entry LogPayload) {
	//create json we'll send to log service
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	//call service
	request, err := http.NewRequest("POST", "http://logger-service/log", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	//handle error when request to logger service fails
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	defer response.Body.Close()

	// make sure we get correct status code
	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling logger service"), response.StatusCode)
		return
	}

	var jsonFromService jsonResponse
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	//handle error when json response cannot be decoded
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, response.StatusCode, jsonFromService)
}

func (app *Config) sendMail(w http.ResponseWriter, mailData MailPayload) {
	log.Println("inside sendMail")
	//create json we'll send to log service
	jsonData, _ := json.MarshalIndent(mailData, "", "\t")

	//call service
	request, err := http.NewRequest("POST", "http://mailer-service/send", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	//handle error when request to logger service fails
	if err != nil {
		log.Println(request)
		app.errorJSON(w, err)
		return
	}

	defer response.Body.Close()

	// make sure we get correct status code
	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling mail service"), response.StatusCode)
		return
	}

	var jsonFromService jsonResponse
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	//handle error when json response cannot be decoded
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, response.StatusCode, jsonFromService)
}

func (app *Config) logRabbitEvent(w http.ResponseWriter, entry LogPayload) {
	err := app.pushToQueue(entry.Name, entry.Data)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var jsonFromService jsonResponse
	jsonFromService.Error = false
	jsonFromService.Message = "logged via RabbitMQ"
	app.writeJSON(w, http.StatusAccepted, jsonFromService)
}

func (app *Config) pushToQueue(name, msg string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: msg,
	}

	jsonPayload, _ := json.MarshalIndent(&payload, "", "\t")
	err = emitter.Push(string(jsonPayload), "log.INFO")
	if err != nil {
		return err
	}
	return nil
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) logRPCEvent(w http.ResponseWriter, entry LogPayload) {
	client, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		log.Println(err)
		app.errorJSON(w, err)
		return
	}

	rpcPayload := RPCPayload{
		Name: entry.Name,
		Data: entry.Data,
	}

	var response string
	err = client.Call("RPCServer.LogInfo", rpcPayload, &response)
	if err != nil {
		log.Println(err)
		app.errorJSON(w, err)
		return
	}

	var result jsonResponse
	result.Error = false
	result.Message = response
	app.writeJSON(w, http.StatusAccepted, result)
}

func (app *Config) LogViaGRPC(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	conn, err := grpc.Dial("logger-service:50001",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		log.Println(err)
		app.errorJSON(w, err)
		return
	}

	defer conn.Close()

	c := logs.NewLogServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.WriteLog(ctx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: requestPayload.Log.Name,
			Data: requestPayload.Log.Data,
		},
	})
	if err != nil {
		log.Println(err)
		app.errorJSON(w, err)
		return
	}

	var resp jsonResponse
	resp.Error = false
	resp.Data = "logged via grpc"

	app.writeJSON(
		w,
		http.StatusAccepted,
		resp)
}
