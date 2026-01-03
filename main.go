package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image/color"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

// --- Configuration ---
const (
	DefaultSize      = 128
	DefaultBgColor   = "cccccc"
	DefaultFontColor = "969696"
	CacheSize        = 2000
)

// Global Cache & Fonts
var imageCache *lru.Cache[string, []byte]
var fontRegular *truetype.Font
var fontBold *truetype.Font

func main() {
	var err error
	// 1. Initialize Cache
	imageCache, err = lru.New[string, []byte](CacheSize)
	if err != nil {
		log.Fatal(err)
	}

	// 2. Pre-load fonts (Optimization: Load once on startup)
	fontRegular, _ = truetype.Parse(goregular.TTF)
	fontBold, _ = truetype.Parse(gobold.TTF)

	// 3. Define Routes
	// User Avatars: /avatar/John+Doe
	http.HandleFunc("/avatar/", handleAvatar)

	// Placeholders: /placeholder/600x400
	http.HandleFunc("/placeholder/", handlePlaceholder)

	fmt.Println("AvataGo running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// --- Handlers ---

// handleAvatar generates square initials for users
func handleAvatar(w http.ResponseWriter, r *http.Request) {
	// Parse Name from Path or Query
	name := r.URL.Query().Get("name")
	if strings.HasPrefix(r.URL.Path, "/avatar/") {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 2 && parts[2] != "" {
			name = strings.TrimSuffix(parts[2], ".png")
		}
	}
	if name == "" {
		name = "John Doe"
	}

	// Parse Params
	size := parseInt(r.URL.Query().Get("size"), DefaultSize)
	rounded := r.URL.Query().Get("rounded") == "true"
	bold := r.URL.Query().Get("bold") == "true"

	// Background Color
	bgHex := r.URL.Query().Get("background")
	if bgHex == "" {
		bgHex = "f0e9e9"
	}
	if strings.ToLower(bgHex) == "random" {
		bgHex = generateRandomColor(name)
	}

	// Text Color
	fgHex := r.URL.Query().Get("color")
	if fgHex == "" {
		fgHex = "8b5d5d"
	}

	// Cache Key
	key := fmt.Sprintf("Avatar:%s:%d:%t:%t:%s:%s", name, size, rounded, bold, bgHex, fgHex)

	serveImage(w, r, key, func() ([]byte, error) {
		return drawImage(size, size, bgHex, fgHex, getInitials(name), rounded, bold)
	})
}

// handlePlaceholder generates rectangular placeholder images
func handlePlaceholder(w http.ResponseWriter, r *http.Request) {
	width, height := DefaultSize, DefaultSize

	// Parse Dimensions from Path (e.g., /placeholder/800x400)
	pathMetric := strings.TrimPrefix(r.URL.Path, "/placeholder/")
	pathMetric = strings.TrimSuffix(pathMetric, ".png")

	// Regex to find 123x456
	re := regexp.MustCompile(`^(\d+)x(\d+)$`)
	matches := re.FindStringSubmatch(pathMetric)

	if len(matches) == 3 {
		width = parseInt(matches[1], DefaultSize)
		height = parseInt(matches[2], DefaultSize)
	} else {
		// Fallback to query params
		width = parseInt(r.URL.Query().Get("w"), DefaultSize)
		height = parseInt(r.URL.Query().Get("h"), DefaultSize)
	}

	// Parse Text (Default to dimensions if empty)
	text := r.URL.Query().Get("text")
	if text == "" {
		text = fmt.Sprintf("%d x %d", width, height)
	}

	bgHex := r.URL.Query().Get("bg")
	if bgHex == "" {
		bgHex = DefaultBgColor
	}

	fgHex := r.URL.Query().Get("color")
	if fgHex == "" {
		fgHex = DefaultFontColor
	}

	// Cache Key
	key := fmt.Sprintf("PH:%d:%d:%s:%s:%s", width, height, bgHex, fgHex, text)

	serveImage(w, r, key, func() ([]byte, error) {
		return drawImage(width, height, bgHex, fgHex, text, false, true) // Always bold text for placeholders
	})
}

// --- Core Logic ---

// serveImage handles Caching, ETags, and writing the response
func serveImage(w http.ResponseWriter, r *http.Request, cacheKey string, generator func() ([]byte, error)) {
	etag := fmt.Sprintf(`"%x"`, md5.Sum([]byte(cacheKey)))

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("ETag", etag)

	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if imgData, ok := imageCache.Get(cacheKey); ok {
		w.Header().Set("X-Cache", "HIT")
		w.Write(imgData)
		return
	}

	// Cache Miss - Generate
	imgData, err := generator()
	if err != nil {
		http.Error(w, "Failed to generate image", 500)
		return
	}

	imageCache.Add(cacheKey, imgData)
	w.Header().Set("X-Cache", "MISS")
	w.Write(imgData)
}

// drawImage is the shared drawing engine
func drawImage(w, h int, bgHex, fgHex, text string, rounded, bold bool) ([]byte, error) {
	dc := gg.NewContext(w, h)

	// Draw Background
	dc.SetColor(parseHexColor(bgHex))
	if rounded {
		dc.DrawCircle(float64(w)/2, float64(h)/2, float64(w)/2)
		dc.Fill()
	} else {
		dc.DrawRectangle(0, 0, float64(w), float64(h))
		dc.Fill()
	}

	// Calculate Font Size (Scale to fit)
	// We estimate font size based on the smaller dimension to fit inside
	minDim := float64(w)
	if float64(h) < minDim {
		minDim = float64(h)
	}

	fontSize := minDim * 0.5 // Default 50% of container

	// Adjust font size for long text in placeholders
	if len(text) > 2 {
		fontSize = minDim * 0.15 // Smaller font for longer text
		// Clamp minimum size
		if fontSize < 12 {
			fontSize = 12
		}
	}

	font := fontRegular
	if bold {
		font = fontBold
	}

	face := truetype.NewFace(font, &truetype.Options{Size: fontSize})
	dc.SetFontFace(face)

	// Draw Text
	dc.SetColor(parseHexColor(fgHex))
	dc.DrawStringAnchored(text, float64(w)/2, float64(h)/2, 0.5, 0.5)

	var buf bytes.Buffer
	err := dc.EncodePNG(&buf)
	return buf.Bytes(), err
}

// --- Helpers ---

func getInitials(name string) string {
	parts := strings.Fields(name)
	initials := ""
	for i := 0; i < len(parts) && i < 2; i++ {
		initials += string([]rune(parts[i])[0])
	}
	return strings.ToUpper(initials)
}

func parseHexColor(s string) color.Color {
	s = strings.TrimPrefix(s, "#")
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return color.RGBA{200, 200, 200, 255}
	}
	rgb, _ := hex.DecodeString(s)
	return color.RGBA{R: rgb[0], G: rgb[1], B: rgb[2], A: 255}
}

func generateRandomColor(name string) string {
	hash := md5.Sum([]byte(name))
	return fmt.Sprintf("%02x%02x%02x", hash[0], hash[1], hash[2])
}

func parseInt(s string, def int) int {
	i, err := strconv.Atoi(s)
	if err != nil || i <= 0 {
		return def
	}
	return i
}
