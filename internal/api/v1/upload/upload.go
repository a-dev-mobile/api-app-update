package upload

import (
	"context"
	"mime/multipart"

	"path/filepath"

	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/slog"

	"github.com/a-dev-mobile/app-update-api/internal/config"
	"github.com/a-dev-mobile/app-update-api/internal/models/db"
	"github.com/a-dev-mobile/app-update-api/internal/models/request"
	"github.com/a-dev-mobile/app-update-api/internal/models/response"

	"github.com/a-dev-mobile/app-update-api/internal/utils"
)

// We declare standard error variables that we can use to correlate and identify problems.
var (
	ErrNotFound           = errors.New("document not found")
	ErrDatabaseError      = errors.New("database error")
	ErrApkFileRequired    = errors.New("apk file is required")
	ErrInvalidRequestBody = errors.New("invalid request body")

	ErrInternalServerError = errors.New("internal server error")
	ErrSaveFile            = errors.New("failed to save the uploaded file")
)

// HandlerContext contains dependencies that will be used by HTTP handlers.

type HandlerContext struct {
	DB     *mongo.Client
	Logger *slog.Logger
	Config *config.Config
}

// NewHandlerContext creates a new handler context with dependencies.
func NewHandlerContext(db *mongo.Client, lg *slog.Logger, cfg *config.Config) *HandlerContext {
	return &HandlerContext{
		DB:     db,
		Logger: lg,
		Config: cfg,
	}
}

// Upload processes an incoming request to update application version information.
// @Summary Update Application Version
// @Description This endpoint handles the uploading of an APK file and updates the application version information in the database.
//	It parses the filename to extract necessary details such as package name, version name, version code, and installer package name.
//	The filename should follow the format: "NameApp-ID_<PackageName>-Ver_<VersionName>-Code_<VersionCode>-Installer_<InstallerPackageName>.apk".
//	After parsing and validation, the APK file is stored, and its details are updated in the database, including file size and update description.
// @Tags Application Update
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "APK File to be uploaded; should follow the specified naming convention."
// @Param checksum formData string true "Checksum of the APK file."
// @Param updateDescription formData string false "Description of what's new in this version."
// @Success 200 {object} response.StatusResponse "Successful response indicating the successful processing and updating of the application version information."
// @Failure 400 {object} response.StatusResponse "Bad Request: Occurs when the request body is invalid, parameters are missing or improperly formatted, or if the file does not follow the expected naming convention."
// @Failure 404 {object} response.StatusResponse "Not Found: The requested document is not found in the database."
// @Failure 500 {object} response.StatusResponse "Internal Server Error: An error occurred during processing, such as a database error, file upload issue, or any other server-side problem."
// @Router /upload [post]
func (hctx *HandlerContext) Upload(c *gin.Context) {
	var req request.UploadRequest
	// Processing multipart/form-data requests
	if err := c.ShouldBind(&req); err != nil {

		c.JSON(http.StatusBadRequest, response.StatusResponse{Message: ErrInvalidRequestBody.Error()})
		return
	}

	// Receive the APK file from the request
	apkFile, err := c.FormFile("file")
	if err != nil {

		c.JSON(http.StatusBadRequest, response.StatusResponse{Message: ErrApkFileRequired.Error()})
		return
	}
	// Calculate the file size in megabytes
	fileSize := apkFile.Size 
	// Parse filename to extract details
	if err := req.ParseFilename(apkFile.Filename); err != nil {
		hctx.Logger.Error("Failed to ParseFilename", "error", err.Error())
		c.JSON(http.StatusBadRequest, response.StatusResponse{Message: err.Error()})
		return
	}
	// Validation of request parameters
	if err := req.Validate(); err != nil {

		c.JSON(http.StatusBadRequest, response.StatusResponse{Message: "invalid request parameters"})
		return
	}

	// Get checksum and update description from request
	providedChecksum := c.PostForm("checksum")
	updateDescription := c.PostForm("updateDescription") // Retrieve the update description

	if err := hctx.verifyChecksum(apkFile, providedChecksum); err != nil {
		hctx.Logger.Error("Checksum verification failed", "error", err.Error())
		c.JSON(http.StatusBadRequest, response.StatusResponse{Message: "Checksum verification failed"})
		return
	}
	savePath := filepath.Join(hctx.Config.FileStorage.ApkPath, filepath.Base(apkFile.Filename))
	if err := c.SaveUploadedFile(apkFile, savePath); err != nil {
		hctx.Logger.Error("Failed to save the uploaded file", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, response.StatusResponse{Message: ErrSaveFile.Error()})
		return
	}

	latestVersion := db.VersionInfo{
		VersionCode:       req.VersionCode,
		VersionName:       req.VersionName,
		Checksum:          providedChecksum,
		UpdateDescription: updateDescription,
		FileSize:          fileSize,
	}

	apkURL := hctx.Config.FileStorage.ApkURL + apkFile.Filename
	if err := hctx.updateLatestVersionInfo(c.Request.Context(), req.PackageName, utils.GetActualStoreName(req.InstallerPackageName), latestVersion, apkURL); err != nil {
		hctx.Logger.Error("Error updating version info", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, response.StatusResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.StatusResponse{Message: "Update information processed successfully"})
}
func (hctx *HandlerContext) verifyChecksum(fileHeader *multipart.FileHeader, expectedChecksum string) error {
	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	calculatedChecksum, err := utils.CalculateChecksum(file)
	if err != nil {
		return err
	}

	if expectedChecksum != calculatedChecksum {
		return errors.New("checksum mismatch")
	}

	return nil
}

func (hctx *HandlerContext) updateLatestVersionInfo(ctx context.Context, packageName, actualStore string, latestVersion db.VersionInfo, url string) error {

	appUpdateCollection := hctx.Config.Database.Collections[string(config.AppUpdate)]
	collection := hctx.DB.Database(hctx.Config.Database.Name).Collection(string(appUpdateCollection))

	// Create a filter to search for the desired document.
	filter := bson.M{"packageName": packageName}

	// Update information about the latest version.
	update := bson.M{
		"$set": bson.M{
			"downloads." + actualStore + ".latestVersion": latestVersion,
			"downloads." + actualStore + ".url":           url,
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		hctx.Logger.Error("Error updating latest version information in the database",
			"packageName", packageName,
			"store", actualStore,
			"error", err.Error())
		return ErrDatabaseError
	}

	if result.MatchedCount == 0 {
		return ErrNotFound
	}

	hctx.Logger.Info("Successfully updated latest version information",
		slog.String("packageName", packageName),
		slog.String("store", actualStore))
	return nil
}
