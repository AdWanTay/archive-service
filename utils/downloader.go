package utils

import (
	"archive/zip"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func ValidExtension(url string, allowed []string) bool {
	for _, ext := range allowed {
		if strings.HasSuffix(strings.ToLower(url), ext) {
			return true
		}
	}
	return false
}
func GenerateID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", b)
}

var DownloadAndZip = downloadAndZip

func downloadAndZip(id string, urls []string) (zipPath string, bad []string, err error) {
	zipPath = fmt.Sprintf("archives/%s.zip", id)
	if err = os.MkdirAll("archives", 0755); err != nil {
		return "", urls, err
	}

	out, err := os.Create(zipPath)
	if err != nil {
		return "", urls, err
	}
	defer func() {
		if cerr := out.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	zipWriter := zip.NewWriter(out)
	defer func() {
		if cerr := zipWriter.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	existing := make(map[string]int)

	for _, u := range urls {
		resp, reqErr := http.Get(u)
		if reqErr != nil || resp.StatusCode != http.StatusOK {
			bad = append(bad, u)
			continue
		}

		func() {
			defer func() {
				if cerr := resp.Body.Close(); cerr != nil && err == nil {
					err = cerr
				}
			}()

			filename := filepath.Base(u)
			count := existing[filename]
			if count > 0 {
				filename = fmt.Sprintf("%s (%d)%s",
					strings.TrimSuffix(filename, filepath.Ext(filename)),
					count,
					filepath.Ext(filename),
				)
			}
			existing[filepath.Base(u)]++

			writer, zipErr := zipWriter.Create(filename)
			if zipErr != nil {
				bad = append(bad, u)
				return
			}

			if _, copyErr := io.Copy(writer, resp.Body); copyErr != nil {
				bad = append(bad, u)
				return
			}
		}()
	}

	return zipPath, bad, nil
}
