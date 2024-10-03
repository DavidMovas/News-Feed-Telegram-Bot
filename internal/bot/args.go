package bot

import "encoding/json"

func ParseJSON[T any](scr string) (T, error) {
	var args T

	if err := json.Unmarshal([]byte(scr), &args); err != nil {
		return *(new(T)), err
	}
	return args, nil
}
