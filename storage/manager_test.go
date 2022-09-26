package storage

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/rs/zerolog/log"
)

var testUser = "goTestUser"
var singletonMgr Manager = nil

func getTestMgr(t *testing.T) (Manager, error) {
	if nil != singletonMgr {
		return singletonMgr, nil
	}

	config, err := LoadConfig("../testData/testConfig.json")
	if nil != err {
		t.Errorf("failed to load config, got: %v", err)
		return nil, err
	}
	mgr, err := NewManager(config)
	if nil != err {
		t.Errorf("failed to initialize storage manager, got: %v", err)
		return nil, err
	}
	singletonMgr = mgr
	return mgr, nil
}

var singletonHttpClient *http.Client = nil

func getTestHttpClient() *http.Client {
	if nil != singletonHttpClient {
		return singletonHttpClient
	}
	singletonHttpClient = &http.Client{}
	return singletonHttpClient
}

func TestMgrMakePath(t *testing.T) {
	testPrefix := "goTestPrefix"
	testCases := [][]string{
		{"/whatever/whatever", "whatever/whatever"},
		{"/whatever/whatever/", "whatever/whatever/"},
		{"whatever/whatever", "whatever/whatever"},
		{"", ""},
	}
	for _, it := range testCases {
		path, err := MakeS3Path(testPrefix, testUser, it[0])
		if nil != err {
			t.Errorf("unexpected path failed validation, got: %v", err)
			return
		}
		expected := fmt.Sprintf("%v/%v/%v", testPrefix, testUser, it[1])
		if path != expected {
			t.Errorf("unexpected result: %v != %v", path, expected)
			return
		}
	}
	invalidTests := []string{
		"frick//jack", "frick/../jack", "..",
		"frick/./jack",
		"`echo hello` > /etc/passwd",
	}
	for _, it := range invalidTests {
		_, err := MakeS3Path(testPrefix, testUser, it)
		if nil == err {
			t.Errorf("path should have failed validation: %v", it)
			return
		}
	}
}

func TestMgrList(t *testing.T) {
	mgr, err := getTestMgr(t)
	if nil != err {
		return
	}
	cx := NewSessionContext(testUser)
	info, err := mgr.List(cx, "@user", "", "")
	if nil != err {
		t.Errorf("failed to list bucket, got: %v", err)
		return
	}
	log.Info().Msg(fmt.Sprintf("ws-storage List got Workspace: %v, Prefix: %v, Objects: %v, Prefixes %v",
		info.Workspace, info.Prefix, info.Objects, info.Prefixes,
	))
	for _, it := range info.Objects {
		if strings.Contains(it.WorkspaceKey, testUser) {
			t.Errorf("workspace key should be below user level: %v", it.WorkspaceKey)
			return
		}
	}
	for _, it := range info.Prefixes {
		if strings.Contains(it, testUser) {
			t.Errorf("workspace prefix should be below user level: %v", it)
			return
		}
	}
}

func TestMgrUpDown(t *testing.T) {
	key := "testObject.txt"
	mgr, err := getTestMgr(t)
	if nil != err {
		return
	}
	cx := NewSessionContext(testUser)
	uploadUrl, err := mgr.UploadUrl(cx, "@user", key)
	if nil != err {
		t.Errorf("failed to generate upload url, got: %v", err)
		return
	}
	httpClient := getTestHttpClient()
	testMessage := "this is a test"
	// upload twice - make sure PUT works either way
	for i := 0; i < 2; i += 1 {
		testBytes := bytes.NewBufferString(testMessage)
		req, err := http.NewRequest(http.MethodPut, uploadUrl, testBytes)
		req.ContentLength = int64(testBytes.Len())
		if nil != err {
			t.Errorf("%v failed to setup upload request, got: %v", i, err)
			return
		}
		resp, err := httpClient.Do(req)
		if nil != err {
			t.Errorf("%v failed to upload test content, got: %v", i, err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Errorf("%v non-200 upload response, got: %v", i, resp)
			return
		}
	}

	downloadUrl, err := mgr.DownloadUrl(cx, "@user", key)
	if nil != err {
		t.Errorf("failed to generate download url, got: %v", err)
		return
	}

	resp, err := httpClient.Get(downloadUrl)
	if nil != err {
		t.Errorf("failed to download test content, got: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("non-200 download response, got: %v", resp)
		return
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		t.Errorf("failed to ReadAll, got: %v", err)
	}
	body := string(bodyBytes[:])
	log.Debug().Msg(fmt.Sprintf("Download got: %v", body))
	if body != testMessage {
		t.Errorf("download does not match upload: %v ?= %v", body, testMessage)
		return
	}

	// List the test object
	info, err := mgr.List(cx, "@user", key, "")
	if nil != err {
		t.Errorf("failed to list object, got: %v", err)
		return
	}
	if len(info.Objects) != 1 || info.Objects[0].WorkspaceKey != key {
		t.Errorf("failed to list object, got: %v", info.Objects)
		return
	}

	// Finally - delete the object
	err = mgr.DeleteObject(cx, "@user", key)
	if nil != err {
		t.Errorf("failed to delete test object, got: %v", err)
		return
	}

	// List again
	info, err = mgr.List(cx, "@user", key, "")
	if nil != err {
		t.Errorf("failed to list object, got: %v", err)
		return
	}
	if len(info.Objects) != 0 {
		t.Errorf("list still has deleted object, got: %v", info.Objects)
		return
	}
}
