package db

type DownloadInfo struct {
	URL            string             `json:"url" bson:"url"`
	LatestVersion  VersionInfo        `json:"latestVersion" bson:"latestVersion"`
	UpdateRequired UpdateRequirements `json:"updateRequirements" bson:"updateRequirements"`
}

type VersionInfo struct {
	VersionCode int    `json:"versionCode" bson:"versionCode"`
	VersionName  string `json:"versionName" bson:"versionName"`
	Checksum    string `json:"checksum" bson:"checksum"` // New field for checksum
}



type UpdateRequirements struct {
	HardUpdate UpdatePolicy `json:"hardUpdate" bson:"hardUpdate"`
	SoftUpdate UpdatePolicy `json:"softUpdate" bson:"softUpdate"`
}

type UpdatePolicy struct {
	MinimumVersionCode int `json:"minimumVersionCode" bson:"minimumVersionCode"`}

type AppUpdateDb struct {
	Name        string                  `json:"name" bson:"name"`
	PackageName string                  `json:"packageName" bson:"packageName"`
	Downloads   map[string]DownloadInfo `json:"downloads" bson:"downloads"`
}
