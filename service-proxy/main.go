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

type Request struct {
	Endpoints [][]string `json:"endpoints"`
}

type Response struct {
	Responses [][]string `json:"responses"`
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
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Read request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Error reading request body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println("Received request:", string(body))

	// Parse request
	var req Request
	err = json.Unmarshal(body, &req)
	if err != nil {
		fmt.Println("Error parsing request:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(req.Endpoints) == 0 {
		res := Response{Responses: [][]string{}}
		// res.Responses = append(res.Responses, []string{"EOF"})
		resJson, err := json.Marshal(res)
		if err != nil {
			fmt.Println("Error marshalling response:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(resJson)
		return
	}

	// Send requests and collect responses
	var res Response

	// Send the first URL a POST request with the remaining URLs as the payload
	payload, err := json.Marshal(Request{Endpoints: req.Endpoints[1:]})
	if err != nil {
		fmt.Println("Error marshalling payload:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for i, endpoint := range req.Endpoints[0] {
		resp, err := http.Post(endpoint, "application/json", ioutil.NopCloser(bytes.NewBuffer(payload)))
		if err != nil {
			fmt.Println("Error sending request to", endpoint, ":", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer resp.Body.Close()

		// Read the response body
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			fmt.Println("Error reading response from", endpoint, ":", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Parse the response and add it to the result
		var response Response
		_ = json.Unmarshal(body, &response)

		response.Responses = append([][]string{{endpoint}}, response.Responses...)

		if i == 0 {
			res.Responses = append(res.Responses, response.Responses...)
		} else {
			for j, resp := range response.Responses {
				res.Responses[j] = append(res.Responses[j], resp...)
			}
		}
	}

	// Write response
	resJson, err := json.Marshal(res)
	if err != nil {
		fmt.Println("Error marshalling response:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println("Sending response:", string(resJson))
	w.Header().Set("Content-Type", "application/json")
	w.Write(resJson)
}
