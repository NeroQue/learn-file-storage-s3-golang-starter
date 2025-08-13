package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func (cfg apiConfig) getObjectURL(key string) string {
	return fmt.Sprintf("https://d3emfweo15k14y.cloudfront.net/%s", key)
}

func getAssetPath(mediaType string) string {
	base := make([]byte, 32)
	_, err := rand.Read(base)
	if err != nil {
		panic("failed to generate random bytes")
	}
	id := base64.RawURLEncoding.EncodeToString(base)

	ext := mediaTypeToExt(mediaType)
	return fmt.Sprintf("%s%s", id, ext)
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetPath)
}

func mediaTypeToExt(mediaType string) string {
	parts := strings.Split(mediaType, "/")
	if len(parts) != 2 {
		return ".bin"
	}
	return "." + parts[1]
}

func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	var buffer bytes.Buffer
	cmd.Stdout = &buffer
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	type Stream struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}
	type FFProbe struct {
		Streams []Stream `json:"streams"`
	}
	var output FFProbe
	err = json.Unmarshal(buffer.Bytes(), &output)
	if err != nil {
		return "", err
	}
	if len(output.Streams) == 0 {
		return "", fmt.Errorf("no streams found in video file")
	}

	width := output.Streams[0].Width
	height := output.Streams[0].Height

	if width == 0 || height == 0 {
		return "", fmt.Errorf("invalid dimensions: width=%d, height=%d", width, height)
	}

	ratio := float64(width) / float64(height)

	if ratio >= 1.7 && ratio <= 1.8 {
		return "landscape", nil
	}

	if ratio >= 0.5 && ratio <= 0.6 {
		return "portrait", nil
	}

	return "other", nil
}

func processVideoForFastStart(filePath string) (string, error) {
	outputFilePath := filePath + ".processing"
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputFilePath)
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return outputFilePath, nil
}
