package response

// VersionInfo holds information about the latest version of the app.

type VersionInfo struct {
	// existing fields...
	// @Description The URL where the latest version can be downloaded from.
	// In debug mode, this URL is always provided, even when no update is required.
	Url string `json:"url,omitempty"`

	// VersionCode represents the version build of the application.
	VersionCode int `json:"versionCode,omitempty"`

	// VersionName is the version of the latest application version.
	VersionName string `json:"versionName,omitempty"`

	// Checksum is the SHA-256 checksum of the latest application version.
	Checksum string `json:"checksum,omitempty"` // Add Swagger description here
}

// @Description holds the overall update information for an application.
type CheckResponse struct {
	// UpdateType indicates the type of update required ('hard', 'soft', or 'none').
	// When UpdateType is 'none', the LatestVersion field will not be present in the response.
	UpdateType string `json:"updateType,omitempty"`

	// LatestVersion contains the details of the latest version if an update is available.
	// This field is omitted when UpdateType is 'none'.
	LatestVersion *VersionInfo `json:"latestVersion,omitempty"`
}
