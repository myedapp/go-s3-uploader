package s3uploader

import (
	"bytes"
	"fmt"
	"github.com/kr/s3/s3util"
	"github.com/satori/go.uuid"
	"github.com/sdemontfort/go-mimemagic"
	"io"
	"net/http"
	"strings"
)

type AwsSettings struct {
	AccessKey string
	SecretKey string
	Bucket    string
}

/**
 * Get file extension from name
 */
func getFileExtension(file string) string {
	split := strings.Split(file, ".")

	// Return the last part of the split for the extension
	return strings.ToLower(split[len(split)-1])
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

		origFile := part.FileName()
		// Generate a unique file name
		ext := getFileExtension(origFile)
		file := fmt.Sprintf("%s", uuid.NewV4())
		fullFile := file + "." + ext

		// Check the mime type
		cBuff := make([]byte, 512)
		part.Read(cBuff)

		mime := mimemagic.Match("", cBuff)
		joinedBody := io.MultiReader(bytes.NewReader(cBuff), part)

		s3util.DefaultConfig.AccessKey = settings.AccessKey
		s3util.DefaultConfig.SecretKey = settings.SecretKey

		s3File, err := s3util.Create("http://"+settings.Bucket+".s3.amazonaws.com/"+fullFile, nil, nil)
		if err != nil {
			return "", "", "", "", err
		}

		io.Copy(s3File, joinedBody)
		defer s3File.Close()

		return file, origFile, ext, mime, nil
	}

	return "", "", "", "", nil
}
