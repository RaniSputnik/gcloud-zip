package gcloudz

import (
	"archive/zip"
	"errors"
	"fmt"
	"google.golang.org/cloud/storage"
	"io"
	"path/filepath"
	"strings"
)

var (
	ErrNoFilesFound = errors.New("No files found in folder")
)

// Pack a folder into zip file
func (r *ZipRequest) Zip(srcFolder string, fileName string, contentType string, metaData *map[string]string) error {
	c := r.context
	bucket := r.bucket

	srcFolder = fmt.Sprintf("%v/", srcFolder)
	query := &storage.Query{Prefix: srcFolder, Delimiter: "/"}

	// TODO read all pages of the response, see example
	// https://cloud.google.com/appengine/docs/go/googlecloudstorageclient/read-write-to-cloud-storage

	objs, err := bucket.List(c, query)
	if err != nil {
		return err
	}

	totalFiles := len(objs.Results)
	if totalFiles == 0 {
		return ErrNoFilesFound
	}

	// create storage file for writing
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

		// read file in our source folder from storage - io.ReadCloser returned from storage
		storageReader, err := bucket.Object(obj.Name).NewReader(c)
		if err != nil {
			return err
		}
		defer storageReader.Close()

		// grab just the filename from directory listing (don't want to store paths in zip)
		_, zipFileName := filepath.Split(obj.Name)
		newFileName := strings.ToLower(zipFileName)

		// add filename to zip
		zipFile, err := zipWriter.Create(newFileName)
		if err != nil {
			return err
		}

		// copy from storage reader to zip writer
		_, err = io.Copy(zipFile, storageReader)
		if err != nil {
			return err
		}
	}

	// Make sure to check the error on Close.
	err = zipWriter.Close()
	if err != nil {
		return err
	}

	// success!
	return nil
}
