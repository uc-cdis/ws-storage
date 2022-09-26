package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type SimpleManager struct {
	config   *Config
	s3client *s3.S3
}

type ObjectInfo struct {
	Workspace    string
	WorkspaceKey string
	SizeBytes    int64
	LastModified time.Time
}

type ListResult struct {
	Workspace string
	Prefix    string
	Objects   []ObjectInfo
	Prefixes  []string
}

type Manager interface {
	List(cx *SessionContext, workspaceIn string, prefix string, page string) (*ListResult, error)
	UploadUrl(cx *SessionContext, workspaceIn string, key string) (string, error)
	DownloadUrl(cx *SessionContext, workspaceIn string, key string) (string, error)
	DeleteObject(cx *SessionContext, workspaceIn string, key string) error
}

//---------------------------------------

// NewManager makes a new manager with the given configuration
func NewManager(config *Config) (mgr Manager, err error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		log.Error().Msgf("Failed to create session - got %v", err)
	}

	// Create S3 service client
	s3client := s3.New(sess)
	mgr = &SimpleManager{
		config:   config,
		s3client: s3client,
	}

	return mgr, nil
}

// MakeS3Path internal method validates inputs,
// and constructs a bucket path from the
// given bucket prefix, user id, and userPath
func MakeS3Path(bucketPrefix, user, userPath string) (string, error) {
	errorPath := "error-in-ws-storage-make-s3-path"
	path := ""
	if bucketPrefix != "" && !strings.HasSuffix(bucketPrefix, "/") {
		path += bucketPrefix + "/"
	} else {
		path = bucketPrefix
	}
	if user == "" || strings.Contains(user, "/") {
		return errorPath, fmt.Errorf("invalid user %v", user)
	}
	path += user
	if !strings.HasPrefix(userPath, "/") {
		path += "/"
	}
	path += userPath
	if strings.Contains(path, "//") || strings.Contains(path, "..") || strings.Contains(path, "/./") || strings.ContainsAny(path, "<>|`'\"$!#&*()\\[]{};~") {
		return errorPath, fmt.Errorf("invalid path - // or .. or /./: %v", path)
	}
	return path, nil
}

// List the prefixes and objects under a given workspace and prefix.
// Currently no paging support - limit to 1000.
// Currently only support user workspace.
func (self *SimpleManager) List(cx *SessionContext, workspaceIn string, prefix string, page string) (*ListResult, error) {
	if workspaceIn != "@user" {
		return nil, fmt.Errorf("invalid workspace - currently only support personal workspaces")
	}
	if page != "" {
		return nil, fmt.Errorf("invalid page - paging not yet implemented")
	}
	workspace := cx.User
	s3path, err := MakeS3Path(self.config.BucketPrefix, workspace, prefix)
	if err != nil {
		return nil, err
	}
	s3prefix, err := MakeS3Path(self.config.BucketPrefix, cx.User, "")
	if err != nil {
		return nil, err
	}
	resp, err := self.s3client.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(self.config.Bucket), Delimiter: aws.String("/"), Prefix: &s3path})
	if err != nil {
		return nil, err
	}

	result := &ListResult{
		Workspace: workspace,
		Prefix:    prefix,
		Objects:   make([]ObjectInfo, len(resp.Contents)),
		Prefixes:  make([]string, len(resp.CommonPrefixes)),
	}
	for ix, item := range resp.Contents {
		result.Objects[ix] = ObjectInfo{
			Workspace:    workspace,
			WorkspaceKey: strings.Replace(*item.Key, s3prefix, "", 1),
			SizeBytes:    *item.Size,
			LastModified: *item.LastModified,
		}
	}
	for ix, item := range resp.CommonPrefixes {
		result.Prefixes[ix] = strings.Replace(*item.Prefix, s3prefix, "", 1)
	}
	return result, nil
}

// UploadUrl generates a presigned upload url
// TODO - support multipart upload
func (self *SimpleManager) UploadUrl(cx *SessionContext, workspaceIn string, key string) (string, error) {
	if workspaceIn != "@user" {
		return "", fmt.Errorf("invalid workspace - currently only support personal workspaces")
	}
	workspace := cx.User
	s3path, err := MakeS3Path(self.config.BucketPrefix, workspace, key)
	if err != nil {
		return "", err
	}
	req, _ := self.s3client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: &self.config.Bucket,
		Key:    &s3path,
	})
	log.Info().Str("Func", "UploadUrl").
		Str("Workspace", workspace).
		Str("Key", key).
		Send()
	return req.Presign(60 * time.Minute)
}

// DownloadUrl generates a presigned download url
// Use the range HTTP header to download range of bytes -
//   https://docs.aws.amazon.com/AmazonS3/latest/dev/GettingObjectsUsingAPIs.html
func (self *SimpleManager) DownloadUrl(cx *SessionContext, workspaceIn string, key string) (string, error) {
	if workspaceIn != "@user" {
		return "", fmt.Errorf("invalid workspace - currently only support personal workspaces")
	}
	workspace := cx.User
	s3path, err := MakeS3Path(self.config.BucketPrefix, workspace, key)
	if err != nil {
		return "", err
	}
	req, _ := self.s3client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: &self.config.Bucket,
		Key:    &s3path,
	})
	log.Info().Str("Func", "DownloadUrl").
		Str("Workspace", workspace).
		Str("Key", key).
		Send()
	return req.Presign(60 * time.Minute)
}

// DownloadUrl generates a presigned download url
// Use the range HTTP header to download range of bytes -
//   https://docs.aws.amazon.com/AmazonS3/latest/dev/GettingObjectsUsingAPIs.html
func (self *SimpleManager) DeleteObject(cx *SessionContext, workspaceIn string, key string) error {
	if workspaceIn != "@user" {
		return fmt.Errorf("invalid workspace - currently only support personal workspaces")
	}
	workspace := cx.User
	s3path, err := MakeS3Path(self.config.BucketPrefix, workspace, key)
	if err != nil {
		return err
	}
	_, err = self.s3client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &self.config.Bucket,
		Key:    &s3path,
	})
	log.Info().Str("Func", "DeleteObject").
		Str("Workspace", workspace).
		Str("Key", key).
		Send()
	return err
}
