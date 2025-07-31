package utils

import (
	"archive/zip"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestValidExtension(t *testing.T) {
	tests := []struct {
		url      string
		allowed  []string
		expected bool
	}{
		{"http://example.com/file.pdf", []string{".pdf", ".jpeg"}, true},
		{"http://example.com/file.jpeg", []string{".pdf", ".jpeg"}, true},
		{"http://example.com/file.jpg", []string{".pdf", ".jpeg"}, false},
		{"http://example.com/file.PDF", []string{".pdf", ".jpeg"}, true}, // проверка регистра
		{"http://example.com/file", []string{".pdf", ".jpeg"}, false},
	}

	for _, tt := range tests {
		result := ValidExtension(tt.url, tt.allowed)
		if result != tt.expected {
			t.Errorf("ValidExtension(%q, %v) = %v; want %v", tt.url, tt.allowed, result, tt.expected)
		}
	}
}

func TestGenerateID(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := GenerateID()
		if len(id) == 0 {
			t.Errorf("GenerateID() returned empty string")
		}
		if ids[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		ids[id] = true
	}
}

func TestDownloadAndZip_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		io.WriteString(w, "test file content")
	}))
	defer server.Close()

	os.RemoveAll("archives")
	defer os.RemoveAll("archives")

	id := "test-task"
	urls := []string{server.URL + "/file1.pdf", server.URL + "/file2.pdf"}
	zipPath, badLinks, err := DownloadAndZip(id, urls)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(badLinks) != 0 {
		t.Errorf("Expected no bad links, got: %v", badLinks)
	}

	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		t.Fatalf("Expected archive %s to exist", zipPath)
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("Failed to open zip: %v", err)
	}
	defer r.Close()

	var filenames []string
	for _, f := range r.File {
		filenames = append(filenames, f.Name)
	}

	if len(filenames) != 2 {
		t.Errorf("Expected 2 files in zip, got %d", len(filenames))
	}
}

func TestDownloadAndZip_BadURL(t *testing.T) {
	os.RemoveAll("archives")
	defer os.RemoveAll("archives")

	id := "bad-task"
	urls := []string{"http://localhost:9999/404.pdf"}

	zipPath, badLinks, err := DownloadAndZip(id, urls)

	if err != nil {
		t.Errorf("Expected no fatal error, got: %v", err)
	}

	if len(badLinks) != 1 || !strings.Contains(badLinks[0], "http") {
		t.Errorf("Expected bad link to be captured, got: %v", badLinks)
	}

	if zipPath == "" {
		t.Errorf("Expected zipPath to be returned even if content is empty")
	}
}
