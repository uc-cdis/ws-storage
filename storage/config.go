package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Config for constructing an AppContext
type Config struct {
	Bucket             string            `json:"bucket"`
	BucketPrefix       string            `json:"bucketprefix"`
	LogLevel           string            `json:"loglevel"`
}



// LoadConfig from a json file
func LoadConfig(configFilePath string) (config *Config, err error) {
	configBytes, err := ioutil.ReadFile(configFilePath)
	if nil != err {
		return nil, err
	}

	config = &Config{}
	
	_ = json.Unmarshal(configBytes, config)
	if "" == config.Bucket {
		return nil, fmt.Errorf("no bucket in config file: %v", configFilePath)
	}
	if "" == config.LogLevel {
		config.LogLevel = "info"
	}
	return config, nil
}
