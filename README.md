# go-s3-uploader
A little wrapper for uploading files to s3 via a MultipartReader stream.

<b>Usage example:</b>
<pre>
var r = http.Request
newFilename, originalFilename, fileExtension, mimeType, err := s3Uploader.Upload(r, s3Uploader.AwsSettings{
  AccessKey: "s3-access-key",
  SecretKey: "s3-secret-key",
  Bucket: "s3-bucket",
})
