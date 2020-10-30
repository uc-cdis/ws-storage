package storage

import (
	"fmt"
	"testing"
)


func TestManagerList(t *testing.T) {
	config, err := LoadConfig("../testData/testConfig.json")
	if nil != err {
		t.Error(fmt.Sprintf("failed to load config, got: %v", err))
		return
	}
	mgr, err := NewManager(config)
	if nil != err {
		t.Error(fmt.Sprintf("failed to initialize storage manager, got: %v", err))
		return
	}
	err = mgr.List()
	if nil != err {
		t.Error(fmt.Sprintf("failed to list bucket, got: %v", err))
		return
	}
}
