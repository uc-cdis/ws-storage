package storage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// Config package-global shared storage config
var mgrSingleton Manager = nil

// SetupHttpListeners setup endpoints with the http engine
func SetupHttpListeners(mgr Manager) error {
	if nil != mgrSingleton {
		return fmt.Errorf("http listeners already configured")
	}
	mgrSingleton = mgr

	http.HandleFunc("/ws-storage/", apiHandler)
	http.HandleFunc("/ws-storage/healthy", healthyHandler)
	http.HandleFunc("/ws-storage/info", infoHandler)
	return nil
}

type ApiRequest struct {
	Verb      string
	Workspace string
	Key       string
	Cx        *SessionContext
}

type ApiResultData interface{}

type ApiResult struct {
	Version int
	Method  string
	Result  string
	Data    ApiResultData
}

// NewApiRequest extracts the api request parameters from
// the URL path and the remote user header.
// urlPath should be $verb/$workspace/$key,
// remoteUser should be set by the API gateway after verfying
// authentication and authorization
func NewApiRequest(url *url.URL, method string, remoteUser string) (*ApiRequest, error) {
	tokens := strings.Split(url.Path, "/")
	if strings.HasPrefix(url.Path, "/") {
		tokens = tokens[1:]
	}
	// remove the /ws-storage prefix
	tokens = tokens[1:]
	if "" == remoteUser {
		return nil, fmt.Errorf("remote user not specified")
	}
	if len(tokens) < 2 {
		return nil, fmt.Errorf("unable to determine verb and workspace from input path")
	}
	result := &ApiRequest{
		Verb:      tokens[0],
		Workspace: tokens[1],
		Key:       strings.Join(tokens[2:], "/"),
		Cx:        NewSessionContext(remoteUser),
	}
	if result.Verb != "list" && result.Verb != "upload" && result.Verb != "download" {
		return nil, fmt.Errorf("invalid request verb: %v", result.Verb)
	}
	if result.Verb == "list" && method == http.MethodDelete {
		result.Verb = "delete"
	}
	if result.Workspace != "@user" {
		return nil, fmt.Errorf("currently only support @user workspace, got %v", result.Workspace)
	}
	return result, nil
}

func (self *ApiRequest) HandleApiRequest(mgr Manager) *ApiResult {
	result := &ApiResult{
		Version: 1,
		Method:  self.Verb,
		Result:  "ok",
		Data:    nil,
	}
	var data ApiResultData = nil
	var err error = nil

	switch self.Verb {
	case "list":
		data, err = mgr.List(self.Cx, self.Workspace, self.Key, "")
	case "upload":
		data, err = mgr.UploadUrl(self.Cx, self.Workspace, self.Key)
	case "download":
		data, err = mgr.DownloadUrl(self.Cx, self.Workspace, self.Key)
	case "delete":
		err = mgr.DeleteObject(self.Cx, self.Workspace, self.Key)
	default:
		err = fmt.Errorf("invalid verb %v", self.Verb)
	}

	if nil != err {
		result.Result = fmt.Sprintf("error - %v", err.Error())
	} else {
		result.Data = data
	}
	return result
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	index := `
	{
		"endpoints": [
			"/ws-storage/list/$workspace/$key",
			"/ws-storage/download/$workspace/$key",
			"/ws-storage/upload/$workspace/$key",
			"/ws-storage/healthy",
			"/ws-storage/info"		
		]
	}`
	w.Header().Add("ContentType", "application/json")
	fmt.Fprint(w, index)
}

func healthyHandler(w http.ResponseWriter, r *http.Request) {
	index := `
	{
		"status": "awesome"
	}`
	w.Header().Add("ContentType", "application/json")
	fmt.Fprint(w, index)
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	apiReq, err := NewApiRequest(r.URL, r.Method, r.Header.Get("REMOTE_USER"))
	start := time.Now()
	sublog := log.Info().
		Str("request", fmt.Sprintf("%v", r.URL))

	w.Header().Add("ContentType", "application/json")
	if nil != err {
		http.Error(w, "{ \"Result\": \"invalid input\" }", 400)
		sublog.Int("statuscode", 400).Dur("durationms", time.Since(start)).Send()
		return
	}

	result := apiReq.HandleApiRequest(mgrSingleton)

	bytes, err := json.Marshal(result)
	if nil != err {
		sublog.Int("statuscode", 500).Dur("durationms", time.Since(start)).Msg("failed json marshall")
		http.Error(w, "error marshaling result", 500)
		return
	}
	_, err = w.Write(bytes)
	if nil != err {
		sublog.Msgf("Error in writing - got %v", err)
	}
	sublog.Int("statuscode", 200).Dur("durationms", time.Since(start)).Send()
}
