package gcloudz

import (
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"
	"io/ioutil"
)

type ZipRequest struct {
	context context.Context
	bucket  *storage.BucketHandle
}

func NewWithBucketNamed(c context.Context, bucketName string) (*ZipRequest, error) {
	client, err := storage.NewClient(c)
	if err != nil {
		return nil, err
	}
	return NewWithBucket(c, client.Bucket(bucketName)), nil
}

func NewWithCredentials(c context.Context, bucketName string, keyFile string) (*ZipRequest, error) {
	// Load the credentials
	jsonKey, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	conf, err := google.JWTConfigFromJSON(
		jsonKey,
		storage.ScopeFullControl,
	)
	if err != nil {
		return nil, err
	}

	// Create the cloud storage client to use
	client, err := storage.NewClient(c, cloud.WithTokenSource(conf.TokenSource(c)))
	if err != nil {
		return nil, err
	}

	return NewWithBucket(c, client.Bucket(bucketName)), nil
}

func NewWithBucket(c context.Context, bucket *storage.BucketHandle) *ZipRequest {
	return &ZipRequest{
		context: c,
		bucket:  bucket,
	}
}
