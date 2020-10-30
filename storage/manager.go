package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"fmt"
)


type Manager struct {
	config   *Config
	s3client *s3.S3
}

func NewManager(config *Config)(mgr *Manager, err error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})

	// Create S3 service client
	s3client := s3.New(sess)
	mgr = &Manager{
		config: config,
		s3client: s3client,
	};

	return mgr, nil
}

func (mgr *Manager) List() error {
	resp, err := mgr.s3client.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(mgr.config.Bucket)})
	if err != nil {
		return err
	}

	for _, item := range resp.Contents {
		fmt.Println("Name:         ", *item.Key)
		fmt.Println("Last modified:", *item.LastModified)
		fmt.Println("Size:         ", *item.Size)
		fmt.Println("Storage class:", *item.StorageClass)
		fmt.Println("")
	}
	return nil
}
