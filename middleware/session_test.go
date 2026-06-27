package middleware

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/Fortuna-Lab/ihy-loglib/logger"
	"github.com/gofiber/fiber/v2"
)

func TestSessionIDGeneratesWhenHeaderMissing(t *testing.T) {
	app := fiber.New()
	app.Use(SessionID(SessionConfig{}))
	app.Get("/", func(c *fiber.Ctx) error {
		if logger.SessionIDFromContext(Ctx(c)) == "" {
			t.Fatal("expected session ID on context")
		}
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	sessionID := resp.Header.Get(DefaultSessionHeader)
	if sessionID == "" {
		t.Fatal("expected response session header")
	}
}

func TestSessionIDPreservesClientHeader(t *testing.T) {
	app := fiber.New()
	app.Use(SessionID(SessionConfig{}))
	app.Get("/", func(c *fiber.Ctx) error {
		got := logger.SessionIDFromContext(Ctx(c))
		if got != "client-session-abc" {
			t.Fatalf("context session = %q, want client-session-abc", got)
		}
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(DefaultSessionHeader, "client-session-abc")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if got := resp.Header.Get(DefaultSessionHeader); got != "client-session-abc" {
		t.Fatalf("response session = %q, want client-session-abc", got)
	}
}
