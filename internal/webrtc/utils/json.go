package utils

import (
	"encoding/json"
	"log/slog"
)

func ToJSONString(content any) (jsonString string, err error) {
	jsonResult, err := json.Marshal(content)
	if err != nil {
		slog.Error("Error converting response to Json", "content", content, "err", err)
		return "", err
	}

	return string(jsonResult), nil
}
