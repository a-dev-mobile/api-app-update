package download

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

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
	hctx.Logger.Info("Download request for APK", slog.String("filename", filename))

	if filepath.Ext(filename) != ".apk" {
		hctx.Logger.Warn("Attempt to download a non-apk file", slog.String("filename", filename))
		c.JSON(http.StatusBadRequest, response.StatusResponse{Message: ErrInvalidFileType.Error()})
		return
	}

	cleanFilename := filepath.Base(filename)
	filePath := filepath.Join(hctx.Config.FileStorage.ApkPath, cleanFilename)

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			hctx.Logger.Error("APK file not found", slog.String("filename", cleanFilename))
			c.JSON(http.StatusNotFound, response.StatusResponse{Message: ErrFileNotFound.Error()})
		} else {
			hctx.Logger.Error("Internal server error while accessing file", slog.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, response.StatusResponse{Message: ErrInternalServerError.Error()})
		}
		return
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		hctx.Logger.Error("Error getting file info", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, response.StatusResponse{Message: ErrInternalServerError.Error()})
		return
	}

	fileSize := fi.Size()
	c.Header("Content-Disposition", "attachment; filename="+cleanFilename)
	c.Header("Content-Type", "application/vnd.android.package-archive")
	c.Header("Accept-Ranges", "bytes")

	rangeHeader := c.GetHeader("Range")
	if rangeHeader == "" {
		// No range header, send the entire file
		c.Header("Content-Length", fmt.Sprintf("%d", fileSize))
		c.File(filePath)
	} else {
		start, end, err := parseRange(rangeHeader, fileSize)
		if err != nil {
			hctx.Logger.Error("Invalid range", slog.String("range", rangeHeader), slog.String("error", err.Error()))
			c.JSON(http.StatusRequestedRangeNotSatisfiable, gin.H{"error": "Invalid range"})
			return
		}

		if start >= fileSize {
			// The range requested starts beyond the end of the file
			hctx.Logger.Warn("Requested range not satisfiable", slog.String("range", rangeHeader))
			c.JSON(http.StatusRequestedRangeNotSatisfiable, gin.H{"error": "Requested range not satisfiable"})
			return
		}

		c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
		c.Header("Content-Length", fmt.Sprintf("%d", end-start+1))
		c.Status(http.StatusPartialContent)

		// Use io.CopyN to send the specified part of the file
		_, err = file.Seek(start, 0)
		if err != nil {
			hctx.Logger.Error("Error seeking in file", slog.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, response.StatusResponse{Message: ErrInternalServerError.Error()})
			return
		}

		_, err = io.CopyN(c.Writer, file, end-start+1)
		if err != nil {
			hctx.Logger.Error("Error during partial file transfer", slog.String("error", err.Error()))
			// Note: Cannot send a new HTTP status code here since the headers are already written
		}
	}

	hctx.Logger.Info("APK file transfer completed", slog.String("filename", cleanFilename))
}

func parseRange(rangeHeader string, fileSize int64) (int64, int64, error) {
	// Example range header: "bytes=0-499"
	rangeParts := strings.Split(rangeHeader, "=")
	if len(rangeParts) != 2 || rangeParts[0] != "bytes" {
		return 0, 0, errors.New("invalid range header")
	}

	rangeValParts := strings.Split(rangeParts[1], "-")
	if len(rangeValParts) != 2 {
		return 0, 0, errors.New("invalid range value")
	}

	start, err := strconv.ParseInt(rangeValParts[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	var end int64
	if rangeValParts[1] == "" {
		// Open-ended range
		end = fileSize - 1
	} else {
		end, err = strconv.ParseInt(rangeValParts[1], 10, 64)
		if err != nil {
			return 0, 0, err
		}
	}

	if start > end || end >= fileSize {
		return 0, 0, errors.New("invalid range span")
	}

	return start, end, nil
}
