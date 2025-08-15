package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

func ValidateFileExtension(filename string, allowedExtensions []string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return false
	}
	ext = ext[1:]
	for _, allowed := range allowedExtensions {
		if ext == allowed {
			return true
		}
	}
	return false
}

func SaveUploadedFile(file *multipart.FileHeader, destDir string) (string, error) {
	ext := filepath.Ext(file.Filename)
	newFileName := uuid.New().String() + ext

	localBasePath := viper.GetString("storage.base_path")
	localDestPath := filepath.Join(localBasePath, destDir)

	if err := os.MkdirAll(localDestPath, os.ModePerm); err != nil {
		return "", err
	}
	localFilePath := filepath.Join(localDestPath, newFileName)

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(localFilePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}

	relativePath := filepath.ToSlash(filepath.Join(destDir, newFileName))
	return relativePath, nil
}

func DeleteFile(relativePath string) error {
	localBasePath := viper.GetString("storage.base_path")
	fullPath := filepath.Join(localBasePath, relativePath)

	absBasePath, err := filepath.Abs(localBasePath)
	if err != nil {
		return err
	}
	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(absFullPath, absBasePath) {
		return fmt.Errorf("invalid file path for deletion: %s", relativePath)
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil
	}

	return os.Remove(fullPath)
}

func FormatSizeUnits(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func GetUploadToken(path string) (string, error) {
	apiURL := viper.GetString("file_server.api_url")
	fullURL := fmt.Sprintf("%s/create/upload?path=%s", apiURL, url.QueryEscape(path))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(fullURL)
	if err != nil {
		return "", fmt.Errorf("failed to connect to file server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("file server returned non-200 status: %d", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode file server response: %w", err)
	}

	if result.Token == "" {
		return "", fmt.Errorf("file server response did not contain a token")
	}

	return result.Token, nil
}

func GetDownloadToken(path string) (string, error) {
	apiURL := viper.GetString("file_server.api_url")
	fullURL := fmt.Sprintf("%s/create/download?path=%s", apiURL, url.QueryEscape(path))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(fullURL)
	if err != nil {
		return "", fmt.Errorf("failed to connect to file server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("file server returned non-200 status: %d", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode file server response: %w", err)
	}

	if result.Token == "" {
		return "", fmt.Errorf("file server response did not contain a token")
	}

	return result.Token, nil
}
