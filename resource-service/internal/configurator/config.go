package configurator

import (
	"github.com/spf13/viper"

	"github.com/nzb3/diploma/resource-service/internal/validator"
)

type Configurator struct {
	viper *viper.Viper
}

func LoadConfig(path string, name string, configType string) error {
	viper.AddConfigPath(path)
	viper.SetConfigName(name)
	viper.SetConfigType(configType)

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return nil
}

func ParseConfig[T any]() (*T, error) {
	config := new(T)

	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	if err := validator.Validate(config); err != nil {
		return nil, err
	}

	return config, nil
}
