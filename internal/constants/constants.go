package constants

import "time"

const (
	CheckAPIEndpoint    = "api-app-update/v1/check"
	UploadAPIEndpoint   = "api-app-update/v1/upload"
	DownloadAPIEndpoint = "api-app-update/v1/download/:filename"

	DefaultServerTimeout = 5 * time.Second // Example timeout value
)
