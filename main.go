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
)

// Response table
type Response struct {
	Table    string `json:"table"`
	Currency string `json:"currency"`
	Code     string `json:"code"`
	Rate     []Rate `json:"rates"`
}

// EUR rate
type Rate struct {
	No            string  `json:"no"`
	EffectiveDate string  `json:"effectiveDate"`
	Mid           float32 `json:"mid"`
}

func main() {
	// Create log file
	file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	logger := log.New(file, "NBP", log.LstdFlags)

	// Create http request
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.nbp.pl/api/exchangerates/rates/a/eur/last/100/?format=json", nil)
	if err != nil {
		logger.Fatal(err)
	}
	req.Header.Add("Accept", `*`)
	req.Header.Add("User-Agent", `*`)

	X, Y := 10, 5
	for i := 0; i < X; i++ {
		// Send request
		start := time.Now()
		res, err := sendRequest(logger, req, client)
		if err != nil {
			logger.Println("An unexpected error has occured:", err)
		}
		end := time.Now()
		logger.Println("Successfully sent http GET request.")

		// Measure response time
		elapsed := end.Sub(start)
		logger.Println("Response time:", elapsed)

		// Process response
		if res != nil {
			if err := processResponse(logger, res); err != nil {
				logger.Println(err)
			}
		}

		// Wait 5s
		time.Sleep(time.Duration(Y) * time.Second)
	}
}

func sendRequest(log *log.Logger, req *http.Request, client *http.Client) (*http.Response, error) {
	// Send request
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Send request failed with %s error", err)
	}

	// Process response
	return res, nil
}

func processResponse(log *log.Logger, res *http.Response) error {
	// Check http response code
	log.Println("Received response with status code:", res.StatusCode)

	// Check response content-type
	if contentType := getContentType(res); contentType != "application/json" {
		log.Printf("Response Content-Type is not application/json, got %s instead", contentType)
	} else {
		log.Printf("Response Content-Type is %s", contentType)
	}

	// Read and unmarshal response body
	response, err := getResponseFromJSON(log, res)
	if err != nil {
		return fmt.Errorf("Invalid JSON format: %s", err)
	}
	log.Println("Successfully unmarshalled response JSON into Response struct - JSON valid")

	// Check for out-of-bounds values
	checkOutOfBoundsValues(log, response)

	return nil
}

func getContentType(res *http.Response) string {
	contentType := strings.Split(res.Header.Values("content-type")[0], ";")
	return contentType[0]
}

func getResponseFromJSON(log *log.Logger, res *http.Response) (*Response, error) {
	response := &Response{}
	body, err := io.ReadAll(res.Body)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func checkOutOfBoundsValues(log *log.Logger, res *Response) {
	var min, max float32 = 4.5, 4.7
	for _, rate := range res.Rate {
		if rate.Mid < min || rate.Mid > max {
			log.Printf("Rate of %s out of [%.1f, %.1f] bounds on: %s", res.Currency, min, max, rate.EffectiveDate)
		}
	}
}
