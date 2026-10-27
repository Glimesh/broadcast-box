package utils

import (
	"encoding/json"
	"log"
)

func ToJSONString(content any) (jsonString string, err error) {
	jsonResult, err := json.Marshal(content)
	if err != nil {
		log.Println("Error converting response", content, "to Json")
		return "", err
	}

	return string(jsonResult), nil
}
