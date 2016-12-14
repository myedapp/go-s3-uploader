package s3uploader

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/kr/s3/s3util"
	"github.com/satori/go.uuid"
	"gopkg.in/h2non/filetype.v1"
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
	var (
		isBase64     = false
		base64Prefix = ";base64,"
		file         string
		origFile     string
		ext          string
		mime         string
	)

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

		// Check the mime type
		cBuff := make([]byte, 512)
		part.Read(cBuff)

		// Check if the file part is base 64 encoded.
		// If so, strip the type prefix string and re-check
		// it's mime type.
		if strings.Contains(string(cBuff), base64Prefix) {
			isBase64 = true

			s := strings.Split(string(cBuff), base64Prefix)

			// Remove the prefix slice element
			s = append(s[:0], s[1:]...)
			sBuff := strings.Join(s, "")

			// Add back to buffer with removed prefix
			cBuff = []byte(sBuff)

			// base 64 decode
			decode, _ := base64.StdEncoding.DecodeString(sBuff)

			// Check the mime type
			kind, _ := filetype.Match([]byte(decode))
			mime = kind.MIME.Value
			origFile = "file." + kind.Extension
		} else {
			kind, _ := filetype.Match(cBuff)
			mime = kind.MIME.Value
			origFile = part.FileName()
		}

		ext = getFileExtension(origFile)

		// Generate a unique file name
		file = fmt.Sprintf("%s", uuid.NewV4())

		joinedBody := io.MultiReader(bytes.NewReader(cBuff), part)

		if isBase64 {
			// base64 decode the stream
			joinedBody = base64.NewDecoder(base64.StdEncoding, joinedBody)
		}

		s3util.DefaultConfig.AccessKey = settings.AccessKey
		s3util.DefaultConfig.SecretKey = settings.SecretKey

		s3File, err := s3util.Create(fmt.Sprintf("http://%s.s3.amazonaws.com/%s.%s", settings.Bucket, file, ext), nil, nil)
		if err != nil {
			return "", "", "", "", err
		}
		defer s3File.Close()

		io.Copy(s3File, joinedBody)

		return file, origFile, ext, mime, nil
	}

	return file, origFile, ext, mime, nil
}
