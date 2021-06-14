build:
	go build -o bin/docker-pull -ldfalgs="-s -w" docker-pull/docker-pull.go
