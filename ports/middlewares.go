package ports

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/google/uuid"
	"github.com/siderustler/go-ecommerce/common/service_logger"
	"github.com/siderustler/go-ecommerce/ports/auth"
)

func ignoreCacheStaticFilesInDev(c fiber.Ctx) error {
	if os.Getenv("ENVIRONMENT") == "DEV" {
		c.Response().Header.Set("Cache-Control", "no-store")
	}
	return c.Next()
}

func isAuthorized(c fiber.Ctx) error {
	sess := session.FromContext(c)
	authorized, ok := sess.Get("authorized").(bool)
	if !ok || !authorized {
		if isHTMXRequest(c) {
			c.Set("HX-Redirect", "/oauth/login")
		} else {
			return c.Redirect().To("/oauth/login")
		}
	}
	return c.Next()
}

const requestIDHeader = "X-Request-ID"
const requestIDLocalKey = "request_id"
const clientIDCookieName = "client_id"
const clientIDLocalKey = "client_id"

func requestIDMiddleware(c fiber.Ctx) error {
	requestID := strings.TrimSpace(c.Get(requestIDHeader))
	if requestID == "" {
		requestID = uuid.NewString()
	}
	c.Locals(requestIDLocalKey, requestID)
	c.Set(requestIDHeader, requestID)
	return c.Next()
}

func clientIDMiddleware(c fiber.Ctx) error {
	clientID := strings.TrimSpace(c.Cookies(clientIDCookieName, ""))
	if _, err := uuid.Parse(clientID); err != nil {
		clientID = uuid.NewString()
		c.Cookie(&fiber.Cookie{
			Name:     clientIDCookieName,
			Value:    clientID,
			Path:     "/",
			HTTPOnly: true,
			SameSite: "Lax",
			Secure:   os.Getenv("ENVIRONMENT") != "DEV",
			MaxAge:   180 * 24 * 60 * 60,
		})
	}
	c.Locals(clientIDLocalKey, clientID)
	return c.Next()
}

func correlationContextMiddleware(c fiber.Ctx) error {
	var ctx context.Context = c.RequestCtx()

	if requestID, _ := c.Locals(requestIDLocalKey).(string); requestID != "" {
		ctx = service_logger.WithRequestID(ctx, requestID)
	}
	if clientID, _ := c.Locals(clientIDLocalKey).(string); clientID != "" {
		ctx = service_logger.WithClientID(ctx, clientID)
	}
	if userID := auth.UserIDFromRequest(c); userID != "" {
		ctx = service_logger.WithUserID(ctx, userID)
	}

	c.SetContext(ctx)
	return c.Next()
}

func requestLoggingMiddleware(baseLogger *slog.Logger) fiber.Handler {
	if baseLogger == nil {
		baseLogger = slog.Default()
	}

	return func(c fiber.Ctx) (err error) {
		startedAt := time.Now()

		defer func() {
			reqLogger := requestLogger(c, baseLogger)
			durationMs := time.Since(startedAt).Milliseconds()

			if rec := recover(); rec != nil {
				reqLogger.ErrorContext(
					c.RequestCtx(),
					"http.request.panic",
					"panic", fmt.Sprint(rec),
					"stack", string(debug.Stack()),
				)
				_ = c.Status(http.StatusInternalServerError).SendString(http.StatusText(http.StatusInternalServerError))
				err = nil
			}

			statusCode := c.Response().StatusCode()
			if !skipRequestCompletedLog(c.Path()) {
				reqLogger.DebugContext(
					c.RequestCtx(),
					"http.request.completed",
					"status", statusCode,
					"duration_ms", durationMs,
					"ip", c.IP(),
				)
			}

			if err != nil {
				reqLogger.ErrorContext(c.RequestCtx(), "http.request.error", "error", err)
				return
			}
			if statusCode >= http.StatusInternalServerError {
				reqLogger.ErrorContext(c.RequestCtx(), "http.request.server_error", "status", statusCode)
			}
		}()

		err = c.Next()
		return err
	}
}

func skipRequestCompletedLog(path string) bool {
	return strings.HasPrefix(path, "/public")
}

func requestLogger(c fiber.Ctx, baseLogger *slog.Logger) *slog.Logger {
	if baseLogger == nil {
		baseLogger = slog.Default()
	}

	requestID, _ := c.Locals(requestIDLocalKey).(string)
	logger := baseLogger.With(
		"request_id", requestID,
		"method", c.Method(),
		"path", c.Path(),
	)
	clientID, _ := c.Locals(clientIDLocalKey).(string)
	if clientID != "" {
		logger = logger.With("client_id", clientID)
	}
	userID := auth.UserIDFromRequest(c)
	if userID != "" {
		logger = logger.With("user_id", userID)
	}

	return logger
}
