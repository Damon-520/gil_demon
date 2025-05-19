package encodingx

import "encoding/json"

func ToJson(v any) string {
	j, _ := json.Marshal(v)
	return string(j)
}
