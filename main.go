package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type InputRequest map[string]interface{}

type Attribute struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

type Trait struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

type ModifyRequest struct {
	Event           string                 `json:"event"`
	EventType       string                 `json:"event_type"`
	AppID           string                 `json:"app_id"`
	UserID          string                 `json:"user_id"`
	MessageID       string                 `json:"message_id"`
	PageTitle       string                 `json:"page_title"`
	PageURL         string                 `json:"page_url"`
	BrowserLanguage string                 `json:"browser_language"`
	ScreenSize      string                 `json:"screen_size"`
	Attributes      map[string]interface{} `json:"attributes"`
	Traits          map[string]interface{} `json:"traits"`
}

func main() {

	requestChannel := make(chan InputRequest)

	var wg sync.WaitGroup

	go worker(requestChannel, &wg)

	http.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		var inputRequest InputRequest
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&inputRequest)

		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		wg.Add(1)

		go processInputRequest(inputRequest, requestChannel, &wg)

		fmt.Fprintf(w, "Received the request successfully")

	})

	port := 7778

	fmt.Printf("Server is running on port %d...\n", port)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	if err != nil {
		fmt.Println("Error starting server:", err)
	}

	wg.Wait()
}

func worker(requestChannel <-chan InputRequest, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		inputRequest := <-requestChannel

		go func(req InputRequest) {
			ModifyRequest := modifyInputRequest(req)
			Producer(ModifyRequest)
		}(inputRequest)
	}
}

func processInputRequest(inputRequest InputRequest, ch chan<- InputRequest, wg *sync.WaitGroup) {

	defer wg.Done()
	ch <- inputRequest
}

func getString(data InputRequest, key string) string {
	val, ok := data[key].(string)
	if !ok {
		return ""
	}
	return val
}

func modifyInputRequest(inputRequest InputRequest) ModifyRequest {

	fmt.Println("INPut:", inputRequest)
	modifyRequest := ModifyRequest{
		Event:           getString(inputRequest, "ev"),
		EventType:       getString(inputRequest, "et"),
		AppID:           getString(inputRequest, "id"),
		UserID:          getString(inputRequest, "uid"),
		MessageID:       getString(inputRequest, "mid"),
		PageTitle:       getString(inputRequest, "t"),
		PageURL:         getString(inputRequest, "p"),
		BrowserLanguage: getString(inputRequest, "l"),
		ScreenSize:      getString(inputRequest, "sc"),
		Attributes:      make(map[string]interface{}),
		Traits:          make(map[string]interface{}),
	}
	mapValues(inputRequest, "atrk", "atrv", "atrt", modifyRequest.Attributes)
	mapValues(inputRequest, "uatrk", "uatrv", "uatrt", modifyRequest.Traits)

	return modifyRequest
}

func Producer(request ModifyRequest) {
	payload, err := json.Marshal(request)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	resp, err := http.Post("https://webhook.site/fbfb25e0-9c2d-433a-90b7-c8e55d341e91", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error sending request to external endpoint:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Request sent to external endpoint. Response:", resp.Status)
}

func mapValues(data InputRequest, key, value, typeAU string, output map[string]interface{}) {
	for i := 1; ; i++ {
		attrKey := fmt.Sprintf("%s%d", key, i)
		valueKey := fmt.Sprintf("%s%d", value, i)
		typeKey := fmt.Sprintf("%s%d", typeAU, i)

		if _, ok := data[attrKey]; !ok {
			break
		}

		attr := getString(data, attrKey)
		attrValue := getString(data, valueKey)
		attrType := getString(data, typeKey)

		if _, exists := output[attrKey]; !exists {
			output[attr] = make(map[string]interface{})
		}
		(output[attr].(map[string]interface{}))["value"] = attrValue
		(output[attr].(map[string]interface{}))["type"] = attrType
	}
}
