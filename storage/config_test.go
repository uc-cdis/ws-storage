package storage

import (
	"fmt"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig("../testData/testConfig.json")
	if nil != err {
		t.Error(fmt.Sprintf("failed to load config, got: %v", err))
		return
	}
	if config.Bucket != "dashboard-707767160287-devplanetv1-gen3" { //"bogus-test-bucket" {
		t.Error(fmt.Sprintf("config did not load the expected bucket: %v", config.Bucket))
		return
	}
	if config.BucketPrefix != "frickjack/test" {
		t.Error(fmt.Sprintf("config did not load the expected bucket prefix: %v", config.BucketPrefix))
		return
	}
	if config.LogLevel != "debug" {
		t.Error(fmt.Sprintf("config did not load the expected log level: %v", config.LogLevel))
		return
	}
}
