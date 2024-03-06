package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PostmanCollection represents the structure of a Postman Collection.
type PostmanCollection struct {
	Info struct {
		Name string `json:"name"`
	} `json:"info"`
	Item []struct {
		Name string `json:"name"`
		Item []struct {
			Name     string     `json:"name"`
			Request  Request    `json:"request"`
			Response []Response `json:"response"`
		} `json:"item"`
	} `json:"item"`
}
type Response struct {
	Name string `json:"name"`
}

// Request represents an HTTP request.
type Request struct {
	Method string `json:"method"`
	URL    struct {
		Raw string `json:"raw"`
	} `json:"url"`
	Header  []Header `json:"header"`
	Body    *Body    `json:"body"`
	Auth    *Auth    `json:"auth"`
	Example string   `json:"example"`
}

// Header represents an HTTP header.
type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Body represents the body of an HTTP request.
type Body struct {
	Mode string `json:"mode"`
	Raw  string `json:"raw"`
}

// Auth represents the authentication details for an HTTP request.
type Auth struct {
	Type   string `json:"type"`
	Bearer []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"bearer"`
}

func main() {
	// Read the Postman Collection JSON file
	data, err := os.ReadFile("postman_collection.json")
	if err != nil {
		fmt.Println("Error reading Postman Collection JSON:", err)
		return
	}

	// Unmarshal JSON
	var collection PostmanCollection
	err = json.Unmarshal(data, &collection)
	if err != nil {
		fmt.Println("Error parsing Postman Collection JSON:", err)
		return
	}

	// Create a directory for the package
	dirName := strings.ToLower(strings.ReplaceAll(collection.Info.Name, " ", "_"))
	err = os.Mkdir(dirName, 0755)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}

	// Generate Go code for request and response models
	for _, group := range collection.Item {
		for _, item := range group.Item {
			// Write request model
			requestFileName := strings.ToLower(strings.ReplaceAll(item.Name, " ", "_")) + "_request.go"
			requestFilePath := filepath.Join(dirName, requestFileName)
			err = os.WriteFile(requestFilePath, []byte(generateRequestModel(item.Request)), 0644)
			if err != nil {
				fmt.Println("Error writing request model:", err)
				return
			}

			// Write response model
			responseFileName := strings.ToLower(strings.ReplaceAll(item.Name, " ", "_")) + "_response.go"
			responseFilePath := filepath.Join(dirName, responseFileName)
			err = os.WriteFile(responseFilePath, []byte(generateResponseModel(item.Response)), 0644)
			if err != nil {
				fmt.Println("Error writing response model:", err)
				return
			}
		}
	}
	fmt.Println("Go code generated successfully!")
}

// generateRequestModel generates Go code for the request model.
func generateRequestModel(req Request) string {
	// Here you can write code to generate Go structs representing the request model
	// Example code generation:
	return fmt.Sprintf(`
package %s

// %sRequest represents the request for %s.
type %sRequest struct {
	// Define struct fields based on request parameters
}`, strings.ToLower(strings.ReplaceAll(req.URL.Raw, " ", "_")), strings.ReplaceAll(req.URL.Raw, " ", ""), req.URL.Raw, strings.ReplaceAll(req.URL.Raw, " ", ""))
}

// generateResponseModel generates Go code for the response model.
func generateResponseModel(responses []Response) string {
	// Here you can write code to generate Go structs representing the response model
	// Example code generation:
	return `
package main

// Response represents the response for the API call.
type Response struct {
	// Define struct fields based on response properties
}`
}
