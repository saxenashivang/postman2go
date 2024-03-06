package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"
)

type PostmanCollection struct {
	Info     Info       `json:"info"`
	Items    []Item     `json:"item"`
	Variable []Variable `json:"variable"`
}

type Info struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Item struct {
	Name        string        `json:"name"`
	Request     Request       `json:"request"`
	Response    []interface{} `json:"response"`
	Event       []Event       `json:"event"`
	Description string        `json:"description"`
}

type Request struct {
	Method string          `json:"method"`
	Header []interface{}   `json:"header"`
	Body   json.RawMessage `json:"body"`
	URL    URL             `json:"url"`
}

type URL struct {
	Raw   string       `json:"raw"`
	Host  []string     `json:"host"`
	Path  []string     `json:"path"`
	Query []QueryParam `json:"query"`
}

type QueryParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Event struct {
	Listen string `json:"listen"`
	Script Script `json:"script"`
}

type Script struct {
	Type string   `json:"type"`
	Exec []string `json:"exec"`
}

type Variable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func main() {
	file, err := os.Open("basic_collection.json")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	var collection PostmanCollection
	err = json.NewDecoder(file).Decode(&collection)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	funcMap := template.FuncMap{
		"joinStrings": joinStrings,
		"joinLines":   joinLines,
		"replace":     strings.Replace,
		"replaceAll":  strings.ReplaceAll,
	}

	tmplStr := `{{ $funcName := .Name | joinStrings | replaceAll " " "_" }}
package razorpay

import (
    "bytes"
    "encoding/json"
    "io"
    "net/http"

    "github.com/myproject/config"
    "github.com/myproject/constants"
    "github.com/myproject/models"
    "github.com/myproject/utils"
    "github.com/pkg/errors"
)

// {{ .Name }}
func {{ $funcName }}(c *gin.Context, reqBody models.{{ .Name }}Req) (models.ProvidersRes, error) {
    var err error
    var res models.ProvidersRes
    res.ProviderName = constants.Providers.RAZORPAY
    url := config.Razorpay.Url + "{{ .Request.URL.Path }}"

    reqJson, err := json.Marshal(reqBody)
    if err != nil {
        return res, errors.Wrap(err, "[{{ $funcName }}][Marshal]")
    }

    // sandbox:true request returns mock response
    authData, err := utils.GetAuthData(c)
    if err != nil {
        return res, errors.Wrap(err, "[{{ $funcName }}][GetAuthData]")
    }
    if authData.Sandbox {
        res.ProviderResponseCode = http.StatusOK
        res.ProviderResponse = mock{{ $funcName }}200(c, reqBody)
        return res, err
    }

    req, _ := http.NewRequest("{{ .Request.Method }}", url, bytes.NewBuffer(reqJson))

    // this is the auth string required to be converted to base 64 as auth
    auth := config.Razorpay.KeyId + ":" + config.Razorpay.KeySecret
    auth = utils.Str2Base64(auth)

    {{ if .Request.Header }}
    // Add headers
    {{ range $header := .Request.Header }}
    req.Header.Add("{{ $header.Key }}", "{{ $header.Value }}")
    {{ end }}
    {{ end }}

    response, err := http.DefaultClient.Do(req)
    if err != nil {
        return res, errors.Wrap(err, "[{{ $funcName }}]")
    }
    res.ProviderResponseCode = response.StatusCode

    defer response.Body.Close()
    body, err := io.ReadAll(response.Body)
    if err != nil {
        return res, errors.Wrap(err, "[{{ $funcName }}][ReadAll]")
    }

    res.ProviderResponse = string(body)
    return res, err
}
`

	tmpl, err := template.New("request_template.gohtml").Funcs(funcMap).Parse(tmplStr)
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}

	// Create a folder for the collection
	folderName := strings.ReplaceAll(strings.ToLower(collection.Info.Name), " ", "_")
	err = os.Mkdir(folderName, 0755)
	if err != nil {
		fmt.Println("Error creating folder:", err)
		return
	}

	// Create models.go file
	modelsFile, err := os.Create(folderName + "/models.go")
	if err != nil {
		fmt.Println("Error creating models.go file:", err)
		return
	}
	defer modelsFile.Close()

	// Write struct definitions for requests and responses in models.go
	// Write struct definitions for requests and responses in models.go
	for _, item := range collection.Items {
		structName := item.Name + "Req"
		structName = strings.ReplaceAll(structName, " ", "")

		_, err = modelsFile.WriteString(fmt.Sprintf("type %s struct {\n", structName))
		if err != nil {
			fmt.Println("Error writing to models.go file:", err)
			return
		}

		if item.Request.Body != nil {
			var reqBody map[string]interface{}
			err = json.Unmarshal(item.Request.Body, &reqBody)
			if err != nil {
				fmt.Println("Error unmarshaling request body:", err)
				return
			}

			for key, value := range reqBody {
				fieldType, err := inferType(value)
				if err != nil {
					fmt.Println("Error inferring type:", err)
					return
				}

				_, err = modelsFile.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"`\n", key, fieldType, key))
				if err != nil {
					fmt.Println("Error writing to models.go file:", err)
					return
				}
			}
		}

		_, err = modelsFile.WriteString("}\n\n")
		if err != nil {
			fmt.Println("Error writing to models.go file:", err)
			return
		}
	}

	for _, item := range collection.Items {
		fileName := folderName + "/" + strings.ReplaceAll(strings.ToLower(item.Name), " ", "_") + ".go"
		file, err := os.Create(fileName)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		defer file.Close()

		err = tmpl.Execute(file, item)
		if err != nil {
			fmt.Println("Error executing template:", err)
			return
		}
	}
}

func joinStrings(input interface{}) string {
	var strs []string

	switch value := input.(type) {
	case []string:
		strs = value
	case string:
		strs = []string{value}
	default:
		return ""
	}

	return strings.Join(strs, ", ")
}

func joinLines(strs []string) string {
	return strings.Join(strs, "\n")
}

func inferType(value interface{}) (string, error) {
	switch value.(type) {
	case string:
		return "string", nil
	case float64:
		return "float64", nil
	case bool:
		return "bool", nil
	case nil:
		return "interface{}", nil
	case map[string]interface{}:
		return "map[string]interface{}", nil
	case []interface{}:
		return "[]interface{}", nil
	default:
		return "", fmt.Errorf("unknown type %T", value)
	}
}
