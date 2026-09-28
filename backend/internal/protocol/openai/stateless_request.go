package openai

import "github.com/tidwall/sjson"

func StatelessResponsesRequest(body []byte) []byte {
	normalized, err := sjson.SetBytes(body, "store", false)
	if err != nil {
		return body
	}
	if stripped, err := sjson.DeleteBytes(normalized, "previous_response_id"); err == nil {
		normalized = stripped
	}
	return normalized
}
