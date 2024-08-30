package helper

import (
	"encoding/json"
)

func Convert[Source any, Target any](s *Source, t *Target) error {
	// JSON marshal
	byteString, err := json.Marshal(s)
	if err != nil {
		return err
	}

	// JSON unmarshal
	err = json.Unmarshal(byteString, t)
	if err != nil {
		return err
	}

	return nil
}
