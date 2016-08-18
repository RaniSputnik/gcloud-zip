package gcloudzip

import (
	"archive/zip"
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/appengine/log"
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
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

// Pack a folder into zip file
func (r *ZipRequest) Zip(srcFolder string, fileName string, contentType string, metaData *map[string]string) bool {
	c := r.context
	bucket := r.bucket

	log.Infof(c, "Packing bucket %v folder %v to file %v", bucket, srcFolder, fileName)

	srcFolder = fmt.Sprintf("%v/", srcFolder)
	query := &storage.Query{Prefix: srcFolder, Delimiter: "/"}

	// TODO read all pages of the response, see example
	// https://cloud.google.com/appengine/docs/go/googlecloudstorageclient/read-write-to-cloud-storage

	objs, err := bucket.List(c, query)
	if err != nil {
		log.Errorf(c, "Packing failed to list bucket %q: %v", r.bucket, err)
		return false
	}

	totalFiles := len(objs.Results)
	if totalFiles == 0 {
		log.Errorf(c, "Packing failed to find objects found in folder %q: %v", bucket, srcFolder)
		return false
	}

	// create storage file for writing
	log.Infof(c, "Writing new zip file to %v/%v for %v files", bucket, fileName, totalFiles)
	storageWriter := bucket.Object(fileName).NewWriter(c)
	defer storageWriter.Close()

	// add optional content type and meta data
	if len(contentType) > 0 {
		storageWriter.ContentType = contentType
	}
	if metaData != nil {
		storageWriter.Metadata = *metaData
	}

	// Create a new zip archive to memory buffer
	zipWriter := zip.NewWriter(storageWriter)

	// go through each file in the folder
	for _, obj := range objs.Results {

		log.Infof(c, "Packing file %v of size %v to zip file", obj.Name, obj.Size)
		//d.dumpStats(obj)

		// read file in our source folder from storage - io.ReadCloser returned from storage
		storageReader, err := bucket.Object(obj.Name).NewReader(c)
		if err != nil {
			log.Errorf(c, "Packing failed to read from bucket %q file %q: %v", bucket, obj.Name, err)
			return false
		}
		defer storageReader.Close()

		// grab just the filename from directory listing (don't want to store paths in zip)
		_, zipFileName := filepath.Split(obj.Name)
		newFileName := strings.ToLower(zipFileName)

		// add filename to zip
		zipFile, err := zipWriter.Create(newFileName)
		if err != nil {
			log.Errorf(c, "Packing failed to create zip file from bucket %q file %q: %v", bucket, zipFileName, err)
			return false
		}

		// copy from storage reader to zip writer
		_, err = io.Copy(zipFile, storageReader)
		if err != nil {
			log.Errorf(c, "Failed to copy from storage reader to zip file: %v", err)
			return false
		}
	}

	// Make sure to check the error on Close.
	log.Infof(c, "Closing zip writer")
	err = zipWriter.Close()
	if err != nil {
		log.Errorf(c, "Packing failed to close zip file writer from bucket %q file %q : %v", bucket, fileName, err)
		return false
	}

	// success!
	log.Infof(c, "Packed files to new cloud storage file %v successful!", fileName)
	return true
}
