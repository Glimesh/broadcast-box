package utils

import (
	"encoding/json"
	"log"
)

func ToJsonString(content any) string {
	jsonResult, err := json.Marshal(content)
	if err != nil {
		log.Println("Error converting response", content, "to Json")
	}

	return string(jsonResult)
}
