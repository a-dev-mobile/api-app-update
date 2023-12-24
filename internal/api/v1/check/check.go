// Package check provides APIs for checking application updates.
// @title Application Update Check API
// @version 1
// @description This service allows applications to check for updates.
// @BasePath /api/v1
package check

import (
	"context"
	"fmt"

	"errors"
	"net/http"
	"time"

	"github.com/a-dev-mobile/app-update-api/internal/config"
	"github.com/a-dev-mobile/app-update-api/internal/models/db"
	"github.com/a-dev-mobile/app-update-api/internal/models/request"
	"github.com/a-dev-mobile/app-update-api/internal/models/response"
	"github.com/a-dev-mobile/app-update-api/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/slog"
)

// HandlerContext holds dependencies for HTTP handlers.
// @Description Dependencies such as database and logger for HTTP handlers.
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

// Check handles the request for checking application updates.
// @Summary Check for application updates
// @Description Checks if there is an update available for a given application package.
// In debug mode, it always returns the latest version information.
// @Tags Application Update
// @Accept json
// @Produce json
// @Param request body request.CheckRequest true "Check Request - includes details such as package version, current version of the application, and debug mode."
// @Success 200 {object} response.CheckResponse "Success Response - returns the type of update required ('hard', 'soft', or 'none') and the latest version information, including the checksum, if an update is available."
// @Failure 400 {object} response.StatusResponse "Bad Request - Invalid request body or parameters."
// @Failure 404 {object} response.StatusResponse "Not Found - Update information for the requested package is not found."
// @Failure 500 {object} response.StatusResponse "Internal Server Error - Unexpected error occurred during processing."
// @Router /check [post]
func (hctx *HandlerContext) Check(c *gin.Context) {
	var updateReq request.CheckRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {

		c.JSON(http.StatusBadRequest, response.StatusResponse{Message: "Invalid request body"})
		return
	}

	if err := updateReq.Validate(); err != nil {

		c.JSON(http.StatusBadRequest, response.StatusResponse{Message: "Invalid request parameters"})
		return
	}

	updateInfo, err := hctx.processUpdateCheck(c.Request.Context(), updateReq)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {

			c.JSON(http.StatusNotFound, response.StatusResponse{Message: "Update information not found"})
			return
		}
		hctx.Logger.Error("Unhandled error in update check", slog.String("error", err.Error()))
		c.JSON(http.StatusNotFound, response.StatusResponse{Message: err.Error()})

		return
	}

	c.JSON(http.StatusOK, updateInfo)
}

// processUpdateCheck orchestrates the process of checking for updates.
func (hctx *HandlerContext) processUpdateCheck(ctx context.Context, updateReq request.CheckRequest) (*response.CheckResponse, error) {
	// Determine the actual store name using the installer package name only once.
	actualStore := utils.GetActualStoreName(updateReq.InstallerPackageName)

	// Retrieve update information from the database for the specific store.
	appUpdateDb, err := hctx.getUpdateInfoFromDB(ctx, updateReq.PackageName, actualStore)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Determine the type of update that is required.
	updateType, latestVersionInfo := hctx.determineUpdateType(updateReq.VersionCode, appUpdateDb.Downloads[actualStore], updateReq.DebugMode)

	// Construct the response.
	updateResponse := &response.CheckResponse{
		UpdateType:    updateType,
		LatestVersion: latestVersionInfo, 
	}

	return updateResponse, nil
}

// determineUpdateType determines the type of update required based on the current version build, download information, and debug mode.
func (hctx *HandlerContext) determineUpdateType(versionCode int, downloadInfo db.DownloadInfo, debugMode bool) (string, *response.VersionInfo) {
	// If debug mode is enabled, return the latest version information with the actual download URL
	if debugMode {
		return "soft", &response.VersionInfo{
			VersionCode: downloadInfo.LatestVersion.VersionCode,
			VersionName: downloadInfo.LatestVersion.VersionName,
			Url:         downloadInfo.URL, // Actual URL for the latest version
			Checksum:    downloadInfo.LatestVersion.Checksum,
		}
	}

	// Standard logic for determining the type of update required
	if downloadInfo.LatestVersion.VersionCode > versionCode {
		versionInfo := &response.VersionInfo{
			VersionCode: downloadInfo.LatestVersion.VersionCode,
			VersionName: downloadInfo.LatestVersion.VersionName,
			Url:         downloadInfo.URL,
			Checksum:    downloadInfo.LatestVersion.Checksum,
		}

		if downloadInfo.UpdateRequired.HardUpdate.MinimumVersionCode > versionCode {
			return "hard", versionInfo
		} else if downloadInfo.UpdateRequired.SoftUpdate.MinimumVersionCode > versionCode {
			return "soft", versionInfo
		}
	}

	return "none", nil // No updates available
}

// getUpdateInfoFromDB Retrieves update information from the database.

func (hctx *HandlerContext) getUpdateInfoFromDB(ctx context.Context, packageName, actualStore string) (*db.AppUpdateDb, error) {
	// Create a context with a timeout.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	appUpdateCollection := hctx.Config.Database.Collections[string(config.AppUpdate)]
	collection := hctx.DB.Database(hctx.Config.Database.Name).Collection(string(appUpdateCollection))

	var appUpdateDb db.AppUpdateDb

	// Setting up a filter to search for a document by PackageName.
	filter := bson.M{"packageName": packageName}

	// Find the first document that matches the filter.
	err := collection.FindOne(ctx, filter).Decode(&appUpdateDb)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			hctx.Logger.Error("Database query timed out", slog.String("packageName", packageName))
			return nil, fmt.Errorf("database query timed out")
		}
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("update information not found")
		}
		hctx.Logger.Error("Error retrieving update information from the database", slog.String("packageName", packageName))
		return nil, fmt.Errorf("database error")
	}

	// Filter out all stores from the Downloads map except for the actual store.
	for storeKey := range appUpdateDb.Downloads {
		if storeKey != actualStore {
			delete(appUpdateDb.Downloads, storeKey)
		}
	}

	if len(appUpdateDb.Downloads) == 0 {
		hctx.Logger.Info("No update information found for the actual store",
			slog.String("packageName", packageName),
			slog.String("store", actualStore))
		return nil, fmt.Errorf("no update information found for the actual store")
	}

	// Log the successful retrieval and adjustment of the data.
	hctx.Logger.Info("Successfully retrieved and filtered update information from the database",
		slog.String("packageName", packageName),
		slog.String("store", actualStore))

	return &appUpdateDb, nil
}
