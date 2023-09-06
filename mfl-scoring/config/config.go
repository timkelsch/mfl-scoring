package config

type APIConfig struct {
	ApiKey string
}

type ParameterStore interface {
	Find(key string, withDecryption bool) string
}

const (
	API_KEY = "/MFL_API_KEY"
)

func NewConfig(parameterStore ParameterStore) APIConfig {
	// retrieve the username from parameter store
	apiKey := parameterStore.Find(API_KEY, true)

	return APIConfig{
		ApiKey: apiKey,
	}
}
