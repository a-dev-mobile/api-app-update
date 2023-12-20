package constants

import "time"

const (
	CheckAPIEndpoint     = "app-update-api/v1/check"
	UploadAPIEndpoint    = "app-update-api/v1/upload"
	DownloadAPIEndpoint  = "app-update-api/v1/download/:filename"

	DefaultServerTimeout = 5 * time.Second // Example timeout value
)