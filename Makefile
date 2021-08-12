build:
	go build -o bin/docker-pull -ldflags="-s -w" docker-pull/docker-pull.go
