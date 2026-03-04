# Go build settings
GOOS_LINUX := linux
GOARCH_AMD64 := amd64
GOARCH_ARM64 := arm64
CGO := 0

linux:
	GOOS=$(GOOS_LINUX) GOARCH=$(GOARCH_AMD64) CGO_ENABLED=$(CGO) go build -o precrawl-linux-amd64 main.go

linux-arm64:
	GOOS=$(GOOS_LINUX) GOARCH=$(GOARCH_ARM64) CGO_ENABLED=$(CGO) go build -o precrawl-linux-arm64 main.go