package ports

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/google/uuid"
)

func TestRequestIDMiddleware_PreservesIncomingID(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(requestIDMiddleware)
	app.Get("/", func(c fiber.Ctx) error {
		requestID, _ := c.Locals(requestIDLocalKey).(string)
		return c.SendString(requestID)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(requestIDHeader, "rid-123")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	body, _ := io.ReadAll(resp.Body)

	if got := resp.Header.Get(requestIDHeader); got != "rid-123" {
		t.Fatalf("expected response request_id rid-123, got %q", got)
	}
	if got := string(body); got != "rid-123" {
		t.Fatalf("expected body request_id rid-123, got %q", got)
	}
}

func TestRequestIDMiddleware_GeneratesWhenMissing(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(requestIDMiddleware)
	app.Get("/", func(c fiber.Ctx) error {
		requestID, _ := c.Locals(requestIDLocalKey).(string)
		return c.SendString(requestID)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	body, _ := io.ReadAll(resp.Body)

	requestID := resp.Header.Get(requestIDHeader)
	if requestID == "" {
		t.Fatal("expected generated request_id header")
	}
	if _, err := uuid.Parse(requestID); err != nil {
		t.Fatalf("expected valid uuid request_id, got %q", requestID)
	}
	if got := string(body); got != requestID {
		t.Fatalf("expected body request_id %q, got %q", requestID, got)
	}
}

func TestClientIDMiddleware_GeneratesCookieWhenMissing(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(clientIDMiddleware)
	app.Get("/", func(c fiber.Ctx) error {
		clientID, _ := c.Locals(clientIDLocalKey).(string)
		return c.SendString(clientID)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	body, _ := io.ReadAll(resp.Body)

	cookies := resp.Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected client_id cookie in response")
	}
	var clientIDCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == clientIDCookieName {
			clientIDCookie = cookie
			break
		}
	}
	if clientIDCookie == nil {
		t.Fatalf("expected %s cookie in response", clientIDCookieName)
	}
	if clientIDCookie.Value == "" {
		t.Fatalf("expected %s cookie value", clientIDCookieName)
	}
	if _, err := uuid.Parse(clientIDCookie.Value); err != nil {
		t.Fatalf("expected valid uuid client_id cookie, got %q", clientIDCookie.Value)
	}
	if got := string(body); got != clientIDCookie.Value {
		t.Fatalf("expected body client_id %q, got %q", clientIDCookie.Value, got)
	}
}

func TestClientIDMiddleware_UsesIncomingCookie(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(clientIDMiddleware)
	app.Get("/", func(c fiber.Ctx) error {
		clientID, _ := c.Locals(clientIDLocalKey).(string)
		return c.SendString(clientID)
	})

	expectedClientID := uuid.NewString()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: clientIDCookieName, Value: expectedClientID})

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	body, _ := io.ReadAll(resp.Body)

	if got := string(body); got != expectedClientID {
		t.Fatalf("expected body client_id %q, got %q", expectedClientID, got)
	}
}

func TestRequestLogger_CorrelatesRequestAndUserID(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer
	baseLogger := slog.New(slog.NewJSONHandler(&buffer, &slog.HandlerOptions{Level: slog.LevelDebug}))

	app := fiber.New()
	sessionStore := session.New()
	app.Use(requestIDMiddleware)
	app.Use(clientIDMiddleware)
	app.Use(requestLoggingMiddleware(baseLogger))
	app.Get("/", sessionStore, func(c fiber.Ctx) error {
		session.FromContext(c).Set("user_id", "user-123")
		requestLogger(c, baseLogger).ErrorContext(c.RequestCtx(), "handler.failed", "error", errors.New("boom"))
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(requestIDHeader, "rid-correlated")
	clientID := uuid.NewString()
	req.AddCookie(&http.Cookie{Name: clientIDCookieName, Value: clientID})

	if _, err := app.Test(req); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	entries := parseJSONLogLines(t, buffer.String())

	var foundHandlerLog bool
	var foundAccessLog bool
	for _, entry := range entries {
		msg, _ := entry["msg"].(string)
		if msg == "handler.failed" {
			foundHandlerLog = true
			if got := entry["request_id"]; got != "rid-correlated" {
				t.Fatalf("expected handler request_id rid-correlated, got %v", got)
			}
			if got := entry["client_id"]; got != clientID {
				t.Fatalf("expected handler client_id %s, got %v", clientID, got)
			}
			if got := entry["user_id"]; got != "user-123" {
				t.Fatalf("expected handler user_id user-123, got %v", got)
			}
		}
		if msg == "http.request.completed" {
			foundAccessLog = true
			if got := entry["request_id"]; got != "rid-correlated" {
				t.Fatalf("expected access request_id rid-correlated, got %v", got)
			}
			if got := entry["client_id"]; got != clientID {
				t.Fatalf("expected access client_id %s, got %v", clientID, got)
			}
			if got := entry["user_id"]; got != "user-123" {
				t.Fatalf("expected access user_id user-123, got %v", got)
			}
		}
	}

	if !foundHandlerLog {
		t.Fatal("expected handler.failed log entry")
	}
	if !foundAccessLog {
		t.Fatal("expected http.request.completed log entry")
	}
}

func TestRequestLoggingMiddleware_LogsReturnedErrorAndPanic(t *testing.T) {
	t.Parallel()

	t.Run("returned error", func(t *testing.T) {
		var buffer bytes.Buffer
		baseLogger := slog.New(slog.NewJSONHandler(&buffer, &slog.HandlerOptions{Level: slog.LevelDebug}))

		app := fiber.New()
		app.Use(requestIDMiddleware)
		app.Use(clientIDMiddleware)
		app.Use(requestLoggingMiddleware(baseLogger))
		app.Get("/", func(c fiber.Ctx) error {
			return errors.New("failed handler")
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(requestIDHeader, "rid-error")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", resp.StatusCode)
		}

		entries := parseJSONLogLines(t, buffer.String())
		var found bool
		for _, entry := range entries {
			if entry["msg"] == "http.request.error" {
				found = true
				if got := entry["request_id"]; got != "rid-error" {
					t.Fatalf("expected request_id rid-error, got %v", got)
				}
				break
			}
		}
		if !found {
			t.Fatal("expected http.request.error log entry")
		}
	})

	t.Run("panic", func(t *testing.T) {
		var buffer bytes.Buffer
		baseLogger := slog.New(slog.NewJSONHandler(&buffer, &slog.HandlerOptions{Level: slog.LevelDebug}))

		app := fiber.New()
		app.Use(requestIDMiddleware)
		app.Use(clientIDMiddleware)
		app.Use(requestLoggingMiddleware(baseLogger))
		app.Get("/", func(c fiber.Ctx) error {
			panic("boom")
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(requestIDHeader, "rid-panic")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", resp.StatusCode)
		}

		entries := parseJSONLogLines(t, buffer.String())
		var found bool
		for _, entry := range entries {
			if entry["msg"] == "http.request.panic" {
				found = true
				if got := entry["request_id"]; got != "rid-panic" {
					t.Fatalf("expected request_id rid-panic, got %v", got)
				}
				stack, _ := entry["stack"].(string)
				if stack == "" {
					t.Fatal("expected panic stack trace in log")
				}
				break
			}
		}
		if !found {
			t.Fatal("expected http.request.panic log entry")
		}
	})
}

func TestRequestLoggingMiddleware_SkipsCompletedLogForIconPaths(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer
	baseLogger := slog.New(slog.NewJSONHandler(&buffer, &slog.HandlerOptions{Level: slog.LevelDebug}))

	app := fiber.New()
	app.Use(requestIDMiddleware)
	app.Use(clientIDMiddleware)
	app.Use(requestLoggingMiddleware(baseLogger))
	app.Get("/public/icons/logo.svg", func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/public/icons/logo.svg", nil)
	if _, err := app.Test(req); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	entries := parseJSONLogLines(t, buffer.String())
	for _, entry := range entries {
		if entry["msg"] == "http.request.completed" {
			t.Fatalf("expected no completed log for icon path, got entry %v", entry)
		}
	}
}

func parseJSONLogLines(t *testing.T, output string) []map[string]any {
	t.Helper()

	entries := make([]map[string]any, 0)
	scanner := bufio.NewScanner(bytes.NewBufferString(output))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		entry := make(map[string]any)
		if err := json.Unmarshal(line, &entry); err != nil {
			t.Fatalf("failed to parse log line %q: %v", string(line), err)
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("failed to scan log output: %v", err)
	}
	return entries
}
