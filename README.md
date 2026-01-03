# AvataGo API Guide

AvataGo is a small HTTP service that renders PNG avatars with user initials and rectangular placeholder images. It relies on the `github.com/fogleman/gg` drawing library and embeds Go fonts for crisp text output.

## Quick Start

```bash
go run ./...
```

The server listens on `:8080` by default and exposes the routes below.

## `/avatar/` Endpoint

Generates a square avatar that displays the initials derived from the provided name.

- **Path**: `/avatar/{name}.png` (URL-encoded; optional) or use the `name` query parameter.
- **Size**: `size` query parameter (default `128`), applied to both width and height.
- **Background Color**: `background` query parameter accepts hex (`f0e9e9`) or the literal `random` to derive a deterministic color per name.
- **Text Color**: `color` query parameter (hex, default `8b5d5d`).
- **Rounded**: `rounded=true` draws a circle instead of a square.
- **Bold**: `bold=true` switches to the embedded Go Bold font.

Example:

```bash
curl "http://localhost:8080/avatar/Jane+Doe.png?size=256&rounded=true&bold=true&background=random"
```

## `/placeholder/` Endpoint

Creates a rectangular placeholder image with custom dimensions and optional overlay text.

- **Path Form**: `/placeholder/{width}x{height}.png`. If omitted, fall back to query parameters `w` and `h` (default `128`).
- **Text**: `text` query parameter (defaults to "{width} x {height}").
- **Background Color**: `bg` query parameter (hex, default `cccccc`).
- **Text Color**: `color` query parameter (hex, default `969696`).

Example:

```bash
curl "http://localhost:8080/placeholder/800x400.png?text=Hero+Image&bg=222222&color=f5f5f5"
```

## Response Characteristics

- All responses are PNG images with `Content-Type: image/png`.
- Successful responses include `Cache-Control: public, max-age=31536000, immutable` and an `ETag` keyed by the query parameters.
- Cached entries are stored in an in-memory LRU (`CacheSize = 2000`) to reduce rendering overhead. Cache hits expose the header `X-Cache: HIT`.

## Error Handling

If generation fails (for example due to invalid parameters), the server responds with HTTP `500` and `Failed to generate image`. Invalid dimensions fallback to safe defaults to keep the server responsive.

## Development Tips

- Customize the defaults by editing the constants in `main.go`.
- Extend `drawImage` if you need additional shapes, padding, or font scaling strategies.
- Consider fronting the service with a CDN when deploying to production so the long-lived cache headers are effective.

