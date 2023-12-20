package download

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/a-dev-mobile/app-update-api/internal/config"
	"github.com/a-dev-mobile/app-update-api/internal/models/response"


	"golang.org/x/exp/slog"
)

// We declare standard error variables that we can use to correlate and identify problems.
var (
	ErrInternalServerError = errors.New("internal server error")
	ErrFileNotFound        = errors.New("file not found")
	ErrInvalidFileType     = errors.New("invalid file type")
)

// HandlerContext contains dependencies that will be used by HTTP handlers.

type HandlerContext struct {
	Logger *slog.Logger
	Config *config.Config
}

// NewHandlerContext creates a new handler context with dependencies.
func NewHandlerContext(lg *slog.Logger, cfg *config.Config) *HandlerContext {
	return &HandlerContext{

		Logger: lg,
		Config: cfg,
	}
}

// DownloadApk provides downloading of the APK file.
// @Summary Download APK file
// @Description Provides the ability to download an APK file stored on the server.
// @Tags APK Download
// @Accept json
// @Produce octet-stream
// @Param filename path string true "Filename of the APK to be downloaded"
// @Success 200 {file} file "APK file"
// @Failure 400 {object} response.StatusResponse "Bad Request - Invalid file type or invalid request format."
// @Failure 404 {object} response.StatusResponse "Not Found - The requested APK file is not found."
// @Failure 500 {object} response.StatusResponse "Internal Server Error - An unexpected error occurred."
// @Router /download/{filename} [get]
func (hctx *HandlerContext) DownloadApk(c *gin.Context) {
	filename := c.Param("filename")

	// Restrict access to APK files only for security
	if filepath.Ext(filename) != ".apk" {

		c.JSON(http.StatusBadRequest, response.StatusResponse{Message: ErrInvalidFileType.Error()})
		return
	}

	// Sanitize the filename
	cleanFilename := filepath.Base(filename)
	filePath := filepath.Join(hctx.Config.FileStorage.ApkPath, cleanFilename)

	fileStat, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			hctx.Logger.Error("File not found", slog.String("filename", cleanFilename))
			c.JSON(http.StatusNotFound, response.StatusResponse{Message: ErrFileNotFound.Error()})

		} else {
			hctx.Logger.Error("Internal error", slog.String("error", err.Error()))

			c.JSON(http.StatusInternalServerError, response.StatusResponse{Message: ErrInternalServerError.Error()})
		}
		return
	}

	// Set headers
	c.Header("Content-Disposition", "attachment; filename="+cleanFilename)
	c.Header("Content-Type", "application/vnd.android.package-archive")
	c.Header("Content-Length", fmt.Sprintf("%d", fileStat.Size()))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")

	c.File(filePath)
}
