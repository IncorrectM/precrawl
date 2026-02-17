# precrawl

A small Go service that pre-renders pages through a headless browser and returns HTML. It proxies request paths to a configured base target URL, waits for a selector, optionally sleeps, then applies post-processing transformers before returning the HTML.

## Features

- Proxy-style rendering: GET /path?query=1 renders ${PRECRAWL_BASE_TARGET_URL}/path?query=1
- Selector wait with timeout (returns HTML even if the selector wait times out)
- Optional post-wait sleep to let the page settle
- Built-in transformers: ImageURLPruner and ClassPruner
- Simple worker pool over a shared chromedp allocator

## Requirements

- Go 1.25+ (see go.mod)
- A Chrome/Chromium installation available to chromedp

## Quick start

1. Set the base target URL.
2. Run the service.
3. Send a GET request to the local server.

Example:

- PRECRAWL_BASE_TARGET_URL=https://example.com go run .
- curl "http://localhost:8080/pages/index?id=212"

## Configuration

Environment variables:

- PRECRAWL_BASE_TARGET_URL (required)
  The origin used to build the target URL for rendering.
- PRECRAWL_DEFAULT_SELECTOR (optional)
  Fallback CSS selector to wait for when the request does not specify a selector header.
  Default: body
- PRECRAWL_RENDER_TIMEOUT (optional)
  Selector wait timeout, e.g. 5s, 200ms. Default: 5s

Runtime behavior:

- The request path and query are appended to PRECRAWL_BASE_TARGET_URL.
- Only GET is supported.
- Selector wait timeout returns HTML with a warning in logs.

## Request headers

- X-Render-Selector: override the default selector for this request
- X-Render-Wait: sleep after selector is visible (Go duration string)
- X-Render-Wait-Ms: sleep after selector is visible (milliseconds)

Only one of X-Render-Wait or X-Render-Wait-Ms should be used.

## Transformers

After prerendering, HTML is passed through these transformers in order:

1. ImageURLPruner (removes img src attributes)
2. ClassPruner (removes class attributes)
3. StylePruner (removes style attributes)

These are currently enabled by default and not configurable.

## Testing

- go test ./...

## Notes

- If the selector wait times out, the service still returns the HTML captured after timeout.
- Long running pages may need a larger PRECRAWL_RENDER_TIMEOUT.
