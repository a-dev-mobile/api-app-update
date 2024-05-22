package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"

	"golang.org/x/exp/slog"
)

func RecoveryMiddleware(lg *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Логирование паники
				lg.Error("Server panic", slog.String("error", fmt.Sprintf("%v", err)))

				// Ответ сервера
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"message": "Internal Server Error",
				})
			}
		}()
		c.Next() // обработка запроса
	}
}
