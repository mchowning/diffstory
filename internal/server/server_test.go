package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mchowning/diffguide/internal/model"
	"github.com/mchowning/diffguide/internal/server"
	"github.com/mchowning/diffguide/internal/storage"
)

func setupTestServer(t *testing.T) (*server.Server, *storage.Store, string) {
	t.Helper()
	baseDir := t.TempDir()
	store, err := storage.NewStoreWithDir(baseDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	srv, err := server.New(store, "0", false) // port 0 for ephemeral port
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	return srv, store, baseDir
}

func TestServer_PostReviewReturns200(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	defer srv.Shutdown(context.Background())

	go srv.Run()
	time.Sleep(10 * time.Millisecond) // give server time to start

	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test Review",
	}
	body, _ := json.Marshal(review)

	resp, err := http.Post(
		"http://127.0.0.1:"+srv.Port()+"/review",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 200, got %d: %s", resp.StatusCode, string(body))
	}
}

func TestServer_GetReviewReturns405(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	defer srv.Shutdown(context.Background())

	go srv.Run()
	time.Sleep(10 * time.Millisecond)

	resp, err := http.Get("http://127.0.0.1:" + srv.Port() + "/review")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}

func TestServer_MissingWorkingDirectoryReturns400(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	defer srv.Shutdown(context.Background())

	go srv.Run()
	time.Sleep(10 * time.Millisecond)

	review := map[string]string{"title": "Test"}
	body, _ := json.Marshal(review)

	resp, err := http.Post(
		"http://127.0.0.1:"+srv.Port()+"/review",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestServer_InvalidJSONReturns400(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	defer srv.Shutdown(context.Background())

	go srv.Run()
	time.Sleep(10 * time.Millisecond)

	resp, err := http.Post(
		"http://127.0.0.1:"+srv.Port()+"/review",
		"application/json",
		strings.NewReader("not valid json"),
	)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestServer_OversizedBodyReturns413(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	defer srv.Shutdown(context.Background())

	go srv.Run()
	time.Sleep(10 * time.Millisecond)

	// Create a valid JSON body larger than 10MB
	// Use a JSON string field with large content
	largeContent := strings.Repeat("x", 11*1024*1024)
	largeJSON := `{"workingDirectory":"/test","title":"` + largeContent + `"}`

	resp, err := http.Post(
		"http://127.0.0.1:"+srv.Port()+"/review",
		"application/json",
		strings.NewReader(largeJSON),
	)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", resp.StatusCode)
	}
}

func TestServer_PostCreatesReviewFile(t *testing.T) {
	srv, store, _ := setupTestServer(t)
	defer srv.Shutdown(context.Background())

	go srv.Run()
	time.Sleep(10 * time.Millisecond)

	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test Review",
	}
	body, _ := json.Marshal(review)

	resp, err := http.Post(
		"http://127.0.0.1:"+srv.Port()+"/review",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	// Verify file was created
	path, _ := store.PathForDirectory("/test/project")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected review file to exist at %s", path)
	}
}

func TestServer_NormalizesWorkingDirectory(t *testing.T) {
	srv, _, baseDir := setupTestServer(t)
	defer srv.Shutdown(context.Background())

	go srv.Run()
	time.Sleep(10 * time.Millisecond)

	// Create a real directory to test normalization
	testDir := t.TempDir()

	review := model.Review{
		WorkingDirectory: testDir + "/", // trailing slash should be normalized
		Title:            "Test Review",
	}
	body, _ := json.Marshal(review)

	resp, err := http.Post(
		"http://127.0.0.1:"+srv.Port()+"/review",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	// Read the stored review and verify the path was normalized
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 file, got %d", len(entries))
	}

	data, err := os.ReadFile(filepath.Join(baseDir, entries[0].Name()))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var stored model.Review
	if err := json.Unmarshal(data, &stored); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// The stored path should not have a trailing slash
	if strings.HasSuffix(stored.WorkingDirectory, "/") {
		t.Errorf("workingDirectory should be normalized, got %q", stored.WorkingDirectory)
	}
}

func TestServer_GracefulShutdown(t *testing.T) {
	srv, _, _ := setupTestServer(t)

	go srv.Run()
	time.Sleep(10 * time.Millisecond)

	// Verify server is running
	resp, err := http.Get("http://127.0.0.1:" + srv.Port() + "/review")
	if err != nil {
		t.Fatalf("server not running: %v", err)
	}
	resp.Body.Close()

	// Shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Errorf("shutdown failed: %v", err)
	}

	// Verify server is stopped (connection should fail)
	_, err = http.Get("http://127.0.0.1:" + srv.Port() + "/review")
	if err == nil {
		t.Error("expected connection to fail after shutdown")
	}
}
