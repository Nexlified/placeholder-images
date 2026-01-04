package handlers

import (
	"crypto/md5"
	"encoding/json"
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
	mux.HandleFunc("/", s.handleHome)
	mux.HandleFunc("/avatar/", s.handleAvatar)
	mux.HandleFunc("/placeholder/", s.handlePlaceholder)
	mux.HandleFunc("GET /health", s.HandleHealth)
}

var placeholderRegex = regexp.MustCompile(`^(\d+)x(\d+)$`)

// formatExtensions maps file extensions to image formats
var formatExtensions = map[string]render.ImageFormat{
	".png":  render.FormatPNG,
	".jpg":  render.FormatJPG,
	".jpeg": render.FormatJPEG,
	".gif":  render.FormatGIF,
	".webp": render.FormatWebP,
}

// extractFormat extracts the image format from a filename, returning the format and the name without extension
func extractFormat(filename string) (render.ImageFormat, string) {
	// Check for known extensions
	for ext, format := range formatExtensions {
		if strings.HasSuffix(filename, ext) {
			return format, strings.TrimSuffix(filename, ext)
		}
	}

	// Default to WebP if no extension found
	return render.FormatWebP, filename
}

// getContentType returns the MIME type for the given format
func getContentType(format render.ImageFormat) string {
	switch format {
	case render.FormatPNG:
		return "image/png"
	case render.FormatJPG, render.FormatJPEG:
		return "image/jpeg"
	case render.FormatGIF:
		return "image/gif"
	case render.FormatWebP:
		return "image/webp"
	default:
		return "image/webp"
	}
}

func (s *Service) handleAvatar(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	format := render.FormatWebP // Default to WebP

	if strings.HasPrefix(r.URL.Path, "/avatar/") {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 2 && parts[2] != "" {
			format, name = extractFormat(parts[2])
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

	key := fmt.Sprintf("Avatar:%s:%d:%t:%t:%s:%s:%s", name, size, rounded, bold, bgHex, fgHex, format)
	s.serveImage(w, r, key, format, func() ([]byte, error) {
		return s.renderer.DrawImageWithFormat(size, size, bgHex, fgHex, render.GetInitials(name), rounded, bold, format)
	})
}

func (s *Service) handlePlaceholder(w http.ResponseWriter, r *http.Request) {
	width, height := config.DefaultSize, config.DefaultSize
	pathMetric := strings.TrimPrefix(r.URL.Path, "/placeholder/")

	// Extract format from path
	format, pathMetric := extractFormat(pathMetric)

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

	key := fmt.Sprintf("PH:%d:%d:%s:%s:%s:%s", width, height, bgHex, fgHex, text, format)
	s.serveImage(w, r, key, format, func() ([]byte, error) {
		return s.renderer.DrawImageWithFormat(width, height, bgHex, fgHex, text, false, true, format)
	})
}

func (s *Service) serveImage(w http.ResponseWriter, r *http.Request, cacheKey string, format render.ImageFormat, generator func() ([]byte, error)) {
	etag := fmt.Sprintf("\"%x\"", md5.Sum([]byte(cacheKey)))

	w.Header().Set("Content-Type", getContentType(format))
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

func (s *Service) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"version": "1.0.0",
	})
}

func (s *Service) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(homePageHTML))
}

const homePageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AvataGo - Avatar and Placeholder Image Generator</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            line-height: 1.6;
            color: #333;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            overflow: hidden;
        }
        header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px;
            text-align: center;
        }
        header h1 {
            font-size: 2.5rem;
            margin-bottom: 10px;
        }
        header p {
            font-size: 1.2rem;
            opacity: 0.9;
        }
        .content {
            padding: 40px;
        }
        .section {
            margin-bottom: 50px;
        }
        .section h2 {
            color: #667eea;
            font-size: 1.8rem;
            margin-bottom: 20px;
            border-bottom: 3px solid #667eea;
            padding-bottom: 10px;
        }
        .section p {
            margin-bottom: 15px;
            color: #555;
        }
        .examples {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 30px;
            margin-top: 30px;
        }
        .example-card {
            background: #f8f9fa;
            border-radius: 8px;
            padding: 20px;
            text-align: center;
            border: 2px solid #e9ecef;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        .example-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 10px 25px rgba(0,0,0,0.1);
        }
        .example-card img {
            max-width: 100%;
            height: auto;
            border-radius: 4px;
            margin-bottom: 15px;
        }
        .example-card h3 {
            color: #333;
            margin-bottom: 10px;
            font-size: 1.1rem;
        }
        .example-card code {
            background: white;
            padding: 10px;
            border-radius: 4px;
            display: block;
            font-size: 0.85rem;
            color: #d63384;
            word-wrap: break-word;
            margin-top: 10px;
            border: 1px solid #dee2e6;
        }
        .params-table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 20px;
            background: white;
        }
        .params-table th,
        .params-table td {
            padding: 12px;
            text-align: left;
            border: 1px solid #dee2e6;
        }
        .params-table th {
            background: #667eea;
            color: white;
            font-weight: 600;
        }
        .params-table tr:nth-child(even) {
            background: #f8f9fa;
        }
        .params-table code {
            background: #e9ecef;
            padding: 2px 6px;
            border-radius: 3px;
            font-size: 0.9rem;
            color: #d63384;
        }
        footer {
            background: #2d3748;
            color: white;
            padding: 30px 40px;
            text-align: center;
        }
        footer a {
            color: #667eea;
            text-decoration: none;
            font-weight: 600;
            transition: color 0.2s;
        }
        footer a:hover {
            color: #764ba2;
        }
        .github-link {
            margin-top: 15px;
            font-size: 1.1rem;
        }
        .github-link svg {
            vertical-align: middle;
            margin-right: 8px;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>🎨 AvataGo</h1>
            <p>Fast, lightweight avatar and placeholder image generator</p>
        </header>
        
        <div class="content">
            <div class="section">
                <h2>About AvataGo</h2>
                <p>AvataGo is a high-performance HTTP service written in Go that generates images on-demand. Whether you need user avatars with initials or placeholder images for your designs, AvataGo delivers them instantly with smart caching and multiple format support.</p>
                <p><strong>Key Features:</strong></p>
                <ul style="margin-left: 20px; color: #555;">
                    <li>Generate circular or square avatars with user initials</li>
                    <li>Create custom placeholder images with any dimensions</li>
                    <li>Support for multiple formats: WebP, PNG, JPG, GIF</li>
                    <li>Smart in-memory LRU caching for optimal performance</li>
                    <li>ETag support for conditional requests</li>
                    <li>Customizable colors, sizes, and styles</li>
                </ul>
            </div>

            <div class="section">
                <h2>Avatar Examples</h2>
                <div class="examples">
                    <div class="example-card">
                        <img src="/avatar/John+Doe?size=128&rounded=false" alt="Square Avatar">
                        <h3>Square Avatar</h3>
                        <code>/avatar/John+Doe?size=128</code>
                    </div>
                    <div class="example-card">
                        <img src="/avatar/Jane+Smith?size=128&rounded=true&background=random" alt="Round Avatar">
                        <h3>Round Avatar (Random Color)</h3>
                        <code>/avatar/Jane+Smith?size=128&rounded=true&background=random</code>
                    </div>
                    <div class="example-card">
                        <img src="/avatar/Alex+Johnson?size=128&rounded=true&bold=true&background=3498db&color=ffffff" alt="Custom Avatar">
                        <h3>Custom Colors & Bold</h3>
                        <code>/avatar/Alex+Johnson?size=128&rounded=true&bold=true&background=3498db&color=ffffff</code>
                    </div>
                </div>
            </div>

            <div class="section">
                <h2>Avatar URL Parameters</h2>
                <table class="params-table">
                    <thead>
                        <tr>
                            <th>Parameter</th>
                            <th>Type</th>
                            <th>Default</th>
                            <th>Description</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr>
                            <td><code>name</code></td>
                            <td>string</td>
                            <td>"John Doe"</td>
                            <td>Name to extract initials from (can be in path or query)</td>
                        </tr>
                        <tr>
                            <td><code>size</code></td>
                            <td>integer</td>
                            <td>128</td>
                            <td>Size in pixels (width and height)</td>
                        </tr>
                        <tr>
                            <td><code>background</code></td>
                            <td>hex/random</td>
                            <td>f0e9e9</td>
                            <td>Background color (hex without #) or "random" for deterministic color</td>
                        </tr>
                        <tr>
                            <td><code>color</code></td>
                            <td>hex</td>
                            <td>auto-contrast</td>
                            <td>Text color (hex without #), auto-calculated if not provided</td>
                        </tr>
                        <tr>
                            <td><code>rounded</code></td>
                            <td>boolean</td>
                            <td>false</td>
                            <td>Set to "true" for circular avatars</td>
                        </tr>
                        <tr>
                            <td><code>bold</code></td>
                            <td>boolean</td>
                            <td>false</td>
                            <td>Set to "true" for bold font</td>
                        </tr>
                        <tr>
                            <td>extension</td>
                            <td>.png/.jpg/.gif/.webp</td>
                            <td>.webp</td>
                            <td>Image format (append to name: /avatar/Name.png)</td>
                        </tr>
                    </tbody>
                </table>
            </div>

            <div class="section">
                <h2>Placeholder Examples</h2>
                <div class="examples">
                    <div class="example-card">
                        <img src="/placeholder/300x200?bg=cccccc" alt="Basic Placeholder">
                        <h3>Basic Placeholder</h3>
                        <code>/placeholder/300x200</code>
                    </div>
                    <div class="example-card">
                        <img src="/placeholder/300x200?text=Hero+Image&bg=2c3e50&color=ecf0f1" alt="Custom Text">
                        <h3>Custom Text & Colors</h3>
                        <code>/placeholder/300x200?text=Hero+Image&bg=2c3e50&color=ecf0f1</code>
                    </div>
                    <div class="example-card">
                        <img src="/placeholder/300x200?bg=e74c3c,3498db&text=Gradient" alt="Gradient Background">
                        <h3>Gradient Background</h3>
                        <code>/placeholder/300x200?bg=e74c3c,3498db&text=Gradient</code>
                    </div>
                </div>
            </div>

            <div class="section">
                <h2>Placeholder URL Parameters</h2>
                <table class="params-table">
                    <thead>
                        <tr>
                            <th>Parameter</th>
                            <th>Type</th>
                            <th>Default</th>
                            <th>Description</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr>
                            <td>dimensions</td>
                            <td>path</td>
                            <td>128x128</td>
                            <td>Image dimensions in format: /placeholder/{width}x{height}</td>
                        </tr>
                        <tr>
                            <td><code>w</code></td>
                            <td>integer</td>
                            <td>128</td>
                            <td>Width in pixels (alternative to path dimensions)</td>
                        </tr>
                        <tr>
                            <td><code>h</code></td>
                            <td>integer</td>
                            <td>128</td>
                            <td>Height in pixels (alternative to path dimensions)</td>
                        </tr>
                        <tr>
                            <td><code>text</code></td>
                            <td>string</td>
                            <td>"{width} x {height}"</td>
                            <td>Text to display on the placeholder</td>
                        </tr>
                        <tr>
                            <td><code>bg</code></td>
                            <td>hex</td>
                            <td>cccccc</td>
                            <td>Background color (hex without #). Use comma-separated values for gradients (e.g., ff0000,0000ff)</td>
                        </tr>
                        <tr>
                            <td><code>color</code></td>
                            <td>hex</td>
                            <td>auto-contrast</td>
                            <td>Text color (hex without #), auto-calculated if not provided</td>
                        </tr>
                        <tr>
                            <td>extension</td>
                            <td>.png/.jpg/.gif/.webp</td>
                            <td>.webp</td>
                            <td>Image format (append to dimensions: /placeholder/300x200.png)</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>

        <footer>
            <p>Made with love in Nexlified Lab</p>
            <p class="github-link">
                <svg height="24" width="24" viewBox="0 0 16 16" fill="currentColor">
                    <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"></path>
                </svg>
                <a href="https://github.com/Nexlified/grout" target="_blank">View on GitHub</a>
            </p>
        </footer>
    </div>
</body>
</html>`
