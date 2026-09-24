//go:build ignore

// Sniffing and format helpers: the bytes decide the stored format, and a file
// name may only explain a rejection. A store that let the extension win kept a
// downloaded XML error page as image/jpeg and served it with a 200.
package server

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

// sniffImageType decides the stored format from the bytes alone.
//
// The file name is used only to explain a rejection. It deliberately cannot
// make non-image bytes acceptable: a store that fell back to the extension
// whenever sniffing failed kept a downloaded XML error page as image/jpeg and
// served it with a 200.
func sniffImageType(filename string, data []byte, allowInvalid bool) (ext, mime string, err error) {
	if ext, mime, ok := sniffImageMagic(data); ok {
		return ext, mime, nil
	}
	if ext, mime, ok := sniffContentType(data); ok {
		return ext, mime, nil
	}
	if allowInvalid {
		ext := normalizeImageExt(filepath.Ext(filename))
		if ext == "" {
			ext = "jpg"
		}
		return ext, mimeForExt(ext), nil
	}
	return "", "", &validationError{notAnImageMessage(filename, data)}
}

// notAnImageMessage names what the bytes actually are, so the caller can tell
// a wrong-format photo from a downloaded error page.
func notAnImageMessage(filename string, data []byte) string {
	detected := strings.TrimSpace(strings.SplitN(http.DetectContentType(data), ";", 2)[0])
	if ext := normalizeImageExt(filepath.Ext(filename)); ext != "" {
		return fmt.Sprintf("not an image: detected %s; the file name says .%s", detected, ext)
	}
	return fmt.Sprintf("not an image: detected %s", detected)
}

// sniffImageMagic recognizes the formats whose own magic is authoritative.
func sniffImageMagic(data []byte) (ext, mime string, ok bool) {
	if len(data) >= 3 && data[0] == 0xff && data[1] == 0xd8 && data[2] == 0xff {
		return "jpg", "image/jpeg", true
	}
	if len(data) >= 8 && string(data[:8]) == "\x89PNG\r\n\x1a\n" {
		return "png", "image/png", true
	}
	if len(data) >= 6 && (string(data[:6]) == "GIF87a" || string(data[:6]) == "GIF89a") {
		return "gif", "image/gif", true
	}
	if len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		return "webp", "image/webp", true
	}
	return "", "", false
}

// sniffContentType recognizes image bytes without magic of their own.
func sniffContentType(data []byte) (ext, mime string, ok bool) {
	media := strings.TrimSpace(strings.SplitN(http.DetectContentType(data), ";", 2)[0])
	ext, ok = allowedImageTypes[media]
	if !ok {
		return "", "", false
	}
	return ext, media, true
}

// normalizeImageExt maps a file extension to a supported stored format.
func normalizeImageExt(ext string) string {
	ext = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(ext), "."))
	if ext == "jpeg" {
		ext = "jpg"
	}
	switch ext {
	case "jpg", "png", "gif", "webp":
		return ext
	}
	return ""
}

func mimeForExt(ext string) string {
	switch normalizeImageExt(ext) {
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	default:
		return "image/jpeg"
	}
}

func filenameStem(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	if base == "." || base == "/" || base == "" {
		return ""
	}
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func md5Hex(data []byte) string {
	sum := md5.Sum(data)
	return hex.EncodeToString(sum[:])
}

func isDecimalID(value string) bool {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	return err == nil && parsed > 0
}
