package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/rs/zerolog/log"
)

func TestNewApiRequest(t *testing.T) {
	testCases := [][]string{
		{"list/@user", "list", ""},
		{"list/@user/key/key/key", "delete", "key/key/key"},
		{"upload/@user/abc/def", "upload", "abc/def"},
		{"download/@user/123", "download", "123"},
	}
	for _, it := range testCases {
		method := http.MethodGet
		if "delete" == it[1] {
			method = http.MethodDelete
		}
		testUrl, err := url.Parse("https://whatever/ws-storage/" + it[0])
		if nil != err {
			t.Errorf("failed url construction for %v, got: %v", it[0], err)
			return
		}
		req, err := NewApiRequest(testUrl, method, testUser)
		if nil != err {
			t.Errorf("unexpected path %v failed validation, got: %v", testUrl.Path, err)
			return
		}
		if it[1] != req.Verb {
			t.Errorf("unexpected path %v verb, got: %v", it[0], req.Verb)
			return
		}
		if "@user" != req.Workspace {
			t.Errorf("unexpected path %v workspace, got: %v", it[0], req.Workspace)
			return
		}
		if it[2] != req.Key {
			t.Errorf("unexpected path %v key, got: %v", it[0], req.Key)
			return
		}
		if testUser != req.Cx.User {
			t.Errorf("unexpected path %v user, got: %v", it[0], req.Cx.User)
			return
		}
	}

	invalidTests := []string{
		"/ws-storage/frick/whatever/jack",
		"/ws-storage/list/frick/jack",
	}
	for _, it := range invalidTests {
		testUrl, err := url.Parse("https://whatever/ws-storage/" + it)
		if nil != err {
			t.Errorf("failed url construction for %v, got: %v", it[0], err)
			return
		}
		_, err = NewApiRequest(testUrl, http.MethodGet, testUser)
		if nil == err {
			t.Errorf("path should have failed validation: %v", it)
			return
		}
	}
}

func doApiRequest(t *testing.T, verb string) (*ApiResult, error) {
	key := "testObject.txt"
	mgr, err := getTestMgr(t)
	if nil != err {
		return nil, err
	}

	urlStr := "https://whatever/ws-storage/" + verb + "/@user/" + key
	urlMethod := http.MethodGet
	if "delete" == verb {
		urlStr = "https://whatever/ws-storage/list/@user/" + key
		urlMethod = http.MethodDelete
	}
	testUrl, err := url.Parse(urlStr)
	if nil != err {
		t.Errorf("failed url construction for upload/@user, got: %v", err)
		return nil, err
	}
	req, err := NewApiRequest(testUrl, urlMethod, testUser)
	if nil != err {
		t.Errorf("unexpected path %v failed validation, got: %v", testUrl.Path, err)
		return nil, err
	}
	result := req.HandleApiRequest(mgr)
	if "ok" != result.Result {
		t.Errorf("unexpected path %v failed handling, got: %v", testUrl.Path, err)
		return nil, err
	}
	json, err := json.MarshalIndent(result, "", "  ")
	if nil != err {
		t.Errorf("failed to marshall json for %v, got: %v", result, err)
		return nil, err
	}
	fmt.Printf("json result: " + string(json))
	//log.Debug().Msg(fmt.Sprintf("json result: " + string(json)))
	return result, nil
}

func doUploadApiRequest(t *testing.T) bool {
	result, err := doApiRequest(t, "upload")
	if nil != err {
		return false
	}

	uploadUrl := result.Data.(string)
	httpClient := getTestHttpClient()
	testMessage := "this is a test"
	testBytes := bytes.NewBufferString(testMessage)
	req, err := http.NewRequest(http.MethodPut, uploadUrl, testBytes)
	req.ContentLength = int64(testBytes.Len())
	if nil != err {
		t.Errorf("failed to setup upload request, got: %v", err)
		return false
	}
	resp, err := httpClient.Do(req)
	if nil != err {
		t.Errorf("failed to upload test content, got: %v", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("non-200 upload response, got: %v", resp)
		return false
	}
	return true
}

func doDownloadApiRequest(t *testing.T) bool {
	result, err := doApiRequest(t, "download")
	if nil != err {
		return false
	}

	testMessage := "this is a test"
	downloadUrl := result.Data.(string)
	httpClient := getTestHttpClient()
	resp, err := httpClient.Get(downloadUrl)
	if nil != err {
		t.Errorf("failed to download test content, got: %v", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("non-200 download response, got: %v", resp)
		return false
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		t.Errorf("failed to ReadAll, got: %v", err)
	}
	body := string(bodyBytes[:])
	log.Debug().Msg(fmt.Sprintf("Download got: %v", body))
	if body != testMessage {
		t.Errorf("download does not match upload: %v ?= %v", body, testMessage)
		return false
	}
	return true
}

func doListApiRequest(t *testing.T) bool {
	result, err := doApiRequest(t, "list")
	if nil != err {
		return false
	}
	listData := result.Data.(*ListResult)
	key := "testObject.txt"
	if len(listData.Objects) != 1 || listData.Objects[0].WorkspaceKey != key {
		t.Errorf("failed to list object, got: %v", listData.Objects)
		return false
	}

	return true
}

func doDeleteApiRequest(t *testing.T) bool {
	_, err := doApiRequest(t, "delete")
	return nil == err
}

func TestHandleApiRequest(t *testing.T) {
	_ = doUploadApiRequest(t) &&
		doDownloadApiRequest(t) &&
		doListApiRequest(t) &&
		doDeleteApiRequest(t)
}
