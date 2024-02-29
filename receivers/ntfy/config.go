package ntfy

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Config struct {
	ServerURL          string `json:"url,omitempty" yaml:"title,omitempty"`
	Channel            string `json:"channel,omitempty" yaml:"message,omitempty"`
}

func NewConfig(jsonData json.RawMessage) (Config, error) {
	var settings Config
	err := json.Unmarshal(jsonData, &settings)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal settings: %w", err)
	}
	if settings.ServerURL == "" {
		return Config{}, errors.New("could not find url property in settings")
	}
	if settings.ServerURL == "" {
		return Config{}, errors.New("could not find channel property in settings")
	}

	return settings, nil
}
