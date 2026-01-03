package handlers

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/golang-lru/v2"

	"go-avatars/internal/config"
	"go-avatars/internal/render"
	"go-avatars/internal/utils"
)

// Service bundles dependencies required by HTTP handlers.
type Service struct {
	renderer *render.Renderer
	cache    *lru.Cache[string, []byte]
	cfg      config.ServerConfig
}

// NewService wires the handler dependencies.
func NewService(renderer *render.Renderer, cache *lru.Cache[string, []byte], cfg config.ServerConfig) *Service {
	return &Service{renderer: renderer, cache: cache, cfg: cfg}
}

// RegisterRoutes attaches handlers to the provided mux.
func (s *Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/avatar/", s.handleAvatar)
	mux.HandleFunc("/placeholder/", s.handlePlaceholder)
}

var placeholderRegex = regexp.MustCompile(`^(\d+)x(\d+)$`)

func (s *Service) handleAvatar(w http.ResponseWriter, r *http.Request) {
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

	size := utils.ParseIntOrDefault(r.URL.Query().Get("size"), config.DefaultSize)
	rounded := r.URL.Query().Get("rounded") == "true"
	bold := r.URL.Query().Get("bold") == "true"

	bgHex := r.URL.Query().Get("background")
	if bgHex == "" {
		bgHex = config.DefaultAvatarBg
	}
	if strings.EqualFold(bgHex, "random") {
		bgHex = render.GenerateColorHash(name)
	}

	fgHex := r.URL.Query().Get("color")
	if fgHex == "" {
		fgHex = render.GetContrastColor(bgHex)
	}

	key := fmt.Sprintf("Avatar:%s:%d:%t:%t:%s:%s", name, size, rounded, bold, bgHex, fgHex)
	s.serveImage(w, r, key, func() ([]byte, error) {
		return s.renderer.DrawImage(size, size, bgHex, fgHex, render.GetInitials(name), rounded, bold)
	})
}

func (s *Service) handlePlaceholder(w http.ResponseWriter, r *http.Request) {
	width, height := config.DefaultSize, config.DefaultSize
	pathMetric := strings.TrimPrefix(r.URL.Path, "/placeholder/")
	pathMetric = strings.TrimSuffix(pathMetric, ".png")

	if matches := placeholderRegex.FindStringSubmatch(pathMetric); len(matches) == 3 {
		width = utils.ParseIntOrDefault(matches[1], config.DefaultSize)
		height = utils.ParseIntOrDefault(matches[2], config.DefaultSize)
	} else {
		width = utils.ParseIntOrDefault(r.URL.Query().Get("w"), config.DefaultSize)
		height = utils.ParseIntOrDefault(r.URL.Query().Get("h"), config.DefaultSize)
	}

	text := r.URL.Query().Get("text")
	if text == "" {
		text = fmt.Sprintf("%d x %d", width, height)
	}

	bgHex := r.URL.Query().Get("bg")
	if bgHex == "" {
		bgHex = config.DefaultBgColor
	}
	fgHex := r.URL.Query().Get("color")
	if fgHex == "" {
		fgHex = render.GetContrastColor(bgHex)
	}

	key := fmt.Sprintf("PH:%d:%d:%s:%s:%s", width, height, bgHex, fgHex, text)
	s.serveImage(w, r, key, func() ([]byte, error) {
		return s.renderer.DrawImage(width, height, bgHex, fgHex, text, false, true)
	})
}

func (s *Service) serveImage(w http.ResponseWriter, r *http.Request, cacheKey string, generator func() ([]byte, error)) {
	etag := fmt.Sprintf("\"%x\"", md5.Sum([]byte(cacheKey)))

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("ETag", etag)

	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if imgData, ok := s.cache.Get(cacheKey); ok {
		w.Header().Set("X-Cache", "HIT")
		_, _ = w.Write(imgData)
		return
	}

	imgData, err := generator()
	if err != nil {
		http.Error(w, "Failed to generate image", http.StatusInternalServerError)
		return
	}

	s.cache.Add(cacheKey, imgData)
	w.Header().Set("X-Cache", "MISS")
	_, _ = w.Write(imgData)
}
