package s3uploader

import (
	"bytes"
	"fmt"
	"github.com/kr/s3/s3util"
	"github.com/satori/go.uuid"
	"io"
	"net/http"
	"strings"
)

/**
 * Get file extension from name
 */
func getFileExtension(filename string) string {
	split := strings.Split(filename, ".")

	// Return the last part of the split for the extension
	return strings.ToLower(split[len(split)-1])
}

type AwsSettings struct {
	AccessKey string
	SecretKey string
	Bucket    string
}

/**
 * Upload a file straight to s3 via the multipartreader stream
 */
func Upload(r *http.Request, settings AwsSettings) (string, string, string, string, error) {
	reader, readerErr := r.MultipartReader()
	if readerErr != nil {
		return "", "", "", "", readerErr
	}

	// Copy each part to an s3 file (assumes only 1 part)
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		originalFilename := part.FileName()
		// Generate a unique file name
		fileExtension := getFileExtension(originalFilename)
		filename := fmt.Sprintf("%s", uuid.NewV4())
		fullfilename := filename + "." + fileExtension

		// Check the mime type
		checkBuffer := make([]byte, 512)
		part.Read(checkBuffer)

		mimeType, _ := detectMimeType(checkBuffer)
		joinedBody := io.MultiReader(bytes.NewReader(checkBuffer), part)

		s3util.DefaultConfig.AccessKey = settings.AccessKey
		s3util.DefaultConfig.SecretKey = settings.SecretKey

		s3File, err := s3util.Create("http://"+settings.Bucket+".s3.amazonaws.com/"+fullfilename, nil, nil)
		if err != nil {
			return "", "", "", "", err
		}

		io.Copy(s3File, joinedBody)
		defer s3File.Close()

		return filename, originalFilename, fileExtension, mimeType, nil
	}

	return "", "", "", "", nil
}
