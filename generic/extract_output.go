package generic

import (
	"encoding/json"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func ExtractStringFromOutputs(outputs types.Outputs, key string) string {
	if item, ok := outputs[key]; ok {
		if val, ok := item.Value.(string); ok {
			return val
		}
	}
	return ""
}

func ExtractStructFromOutputs(outputs types.Outputs, key string, obj interface{}) bool {
	if item, ok := outputs[key]; ok {
		raw, _ := json.Marshal(item.Value)
		json.Unmarshal(raw, obj)
		return true
	}
	return false
}
