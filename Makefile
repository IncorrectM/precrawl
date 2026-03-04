linux:
	export GOOS=linux
	export GOARCH=amd64    # 目标 CPU 架构
	export CGO_ENABLED=0   # 避免依赖 C 库，简化二进制
	go build -o precrawl-linux-amd64 main.go

linux-arm64:
	export GOOS=linux
	export GOARCH=arm64    # 目标 CPU 架构
	export CGO_ENABLED=0   # 避免依赖 C 库，简化二进制
	go build -o precrawl-linux-arm64 main.go