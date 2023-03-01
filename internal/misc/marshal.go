package misc

import "encoding/json"

func Unmarshal(m map[string]any, v any) error {
	byts, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(byts, v)
}

func Marshal(v any) (m map[string]any, err error) {
	byts, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	m = make(map[string]any)
	err = json.Unmarshal(byts, &m)
	return
}
