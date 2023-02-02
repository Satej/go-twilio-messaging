package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

type Response struct {
	Status     string
	StatusCode int
	Method     string
	Body       []byte
}

type ZenQuotes struct {
	Quote  string `json:"q"`
	Author string `json:"a"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	response, err := sendRequest()
	if err != nil {
		log.Fatal(err)
	}

	var message ZenQuotes
	msg := []ZenQuotes{{
		message.Author,
		message.Quote,
	}}

	err = json.Unmarshal(response.Body, &msg)
	if err != nil {
		log.Fatal(err)
	}

	err = sendSms(msg[0].Quote)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}

func sendSms(message string) error {
	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	messagingServiceSid := os.Getenv("TWILIO_MESSAGING_SERVICE_SID")

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSid,
		Password: authToken,
	})

	recipients := strings.Split(os.Getenv("RECIPIENTS"), ",")

	var failedRequests int
	var successRequests int

	for _, recipient := range recipients {
		params := &twilioApi.CreateMessageParams{}
		params.SetTo(recipient)
		params.SetMessagingServiceSid(messagingServiceSid)
		params.SetBody(message)

		_, err := client.Api.CreateMessage(params)
		if err != nil {
			fmt.Println(err.Error())
			failedRequests++
			continue
		}

		successRequests++
	}

	var output string
	if failedRequests > 0 {
		output = fmt.Sprintf("%v messages could not be sent. Please check your Twilio logs for more information", failedRequests)
	} else {
		output = fmt.Sprintf("%v messages successfully sent", successRequests)
	}

	fmt.Println(output)

	return nil

}

func sendRequest() (*Response, error) {
	r := &Response{}

	httpClient := &http.Client{Timeout: 20 * time.Second}
	zenQuotesUrl := "https://zenquotes.io/api/random"

	req, err := http.NewRequest(http.MethodGet, zenQuotesUrl, nil)
	if err != nil {
		return nil, err
	}

	response, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	r.Status = response.Status
	r.StatusCode = response.StatusCode
	r.Body = body

	return r, nil
}
