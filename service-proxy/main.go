package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Route struct {
	Designation string  `json:"designation"`
	Routes      []Route `json:"routes"`
}

type Response struct {
	Service  string     `json:"service"`
	Response []Response `json:"response"`
}

func main() {
	hostName := fmt.Sprintf(":%s", os.Getenv("PORT"))
	fmt.Println("Listening on", hostName)

	os.Setenv("GODEBUG", "http2server=0")
	os.Setenv("GODEBUG", "http2client=0")

	m := http.NewServeMux()
	m.HandleFunc("/", handleRequest)

	srv := &http.Server{
		Handler:      m,
		Addr:         hostName,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	log.Fatal(srv.ListenAndServe())
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Read the input JSON from the HTTP request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Unmarshal the input JSON into a Route struct
	var rootRoute Route
	err = json.Unmarshal(body, &rootRoute)
	if err != nil {
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}
	fmt.Println("Received request ", rootRoute)

	// Send requests to each service in the input JSON and build the response
	response := sendRequests(rootRoute)

	// Marshal the response into JSON and send it in the HTTP response body
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error creating response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
	fmt.Println("Sent response ", response)
}

func sendRequests(route Route) Response {
	response := Response{
		Service:  route.Designation,
		Response: []Response{},
	}

	for _, nestedRoute := range route.Routes {
		// Send a POST request to the service with the nested route information
		reqBody, err := json.Marshal(nestedRoute)
		if err != nil {
			fmt.Println("Error creating request body:", err)
			response.Response = append(response.Response, Response{})
			continue
		}

		resp, err := http.Post(nestedRoute.Designation, "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			fmt.Println("Error sending request:", err)
			response.Response = append(response.Response, Response{Service: nestedRoute.Designation, Response: []Response{}})
			continue
		}
		defer resp.Body.Close()

		// Read the response from the nested service
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			response.Response = append(response.Response, Response{Service: nestedRoute.Designation, Response: []Response{}})
			continue
		}

		// Unmarshal the response JSON into a Response struct
		var nestedResponse Response
		err = json.Unmarshal(body, &nestedResponse)
		if err != nil {
			fmt.Println("Error parsing response body:", err)
			response.Response = append(response.Response, Response{Service: nestedRoute.Designation, Response: []Response{}})
			continue
		}

		// Append the nested response to the current response
		response.Response = append(response.Response, nestedResponse)
	}

	return response
}
