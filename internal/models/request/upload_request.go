package request

import (
	"errors"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

// UploadRequest represents the request structure for uploading an APK file.
// It parses the filename in the format "NameApp-ID_<PackageName>-Ver_<VersionName>-Code_<VersionCode>-Installer_<InstallerPackageName>-Var_<Variable>.apk"
// to extract the package name, version code, version name, and installer package name.
type UploadRequest struct {
    File *multipart.FileHeader `form:"file" validate:"required"` // The APK file to be uploaded.
    
    // Extracted fields
    PackageName          string // Unique identifier for the app, extracted from the filename.
    VersionCode          int    // Build number of the app, extracted from the filename.
    VersionName          string // Version of the app, extracted from the filename.
    InstallerPackageName string // Name of the installer package, extracted from the filename.
}
// ParseFilename parses the given filename to extract the package name, version code,
// version name, and installer package name. It expects a filename in the specific format
// "NameApp-ID_<PackageName>-Ver_<VersionName>-Code_<VersionCode>-Installer_<InstallerPackageName>-Var_<Variable>.apk".
// Returns an error if the filename does not conform to the expected format.

func (a *UploadRequest) ParseFilename(filename string) error {
    // Split the filename by "-"
    parts := strings.Split(filename, "-")
    if len(parts) != 6 {
        return errors.New("invalid filename format")
    }

    // Parse the PackageName
    a.PackageName = strings.TrimPrefix(parts[1], "ID_")

    // Parse the VersionName
    a.VersionName = strings.TrimPrefix(parts[2], "Ver_")

    // Parse the VersionCode
    versionCode, err := strconv.Atoi(strings.TrimPrefix(parts[3], "Code_"))
    if err != nil {
        return errors.New("invalid version code format")
    }
    a.VersionCode = versionCode


    a.InstallerPackageName = strings.TrimPrefix(parts[4], "Installer_")



    return nil
}

// Validate uses a validator to check the fields of the UploadRequest structure.

func (a *UploadRequest) Validate() error {
    validate := validator.New()
    return validate.Struct(a)
}
