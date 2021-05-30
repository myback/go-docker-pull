# go-docker-pull
The script makes it possible to download a docker-image without docker

!!! Not working with v1 registry
## Use
```bash
> git clone https://github.com/myback/go-docker-pull.git
> cd go-docker-pull
> go build -o bin/docker-pull -ldfalgs="-s -w" docker-pull/docker-pull.go
> bin/docker-pull -h
Usage:
  docker-pull image [image ...] [flags]

Flags:
  -a, --arch string       CPU architecture platform image (default "amd64")
  -h, --help              help for docker-pull
  -d, --only-download     Only download layers
  -o, --os string         OS platform image (default "linux")
  -p, --password string   Registry password
  -s, --save-cache        Do not delete the temp folder
  -u, --user string       Registry user

>
> bin/docker-pull alpine:3.10
3.10: Pulling from library/alpine
21c83c524219: Pull complete
Digest: sha256:a143f3ba578f79e2c7b3022c488e6e12a35836cd4a6eb9e363d7f3a07d848590
> docker pull alpine:3.10
> docker save alpine:3.10 -o alpine_3.10.tar
> sha256sum *.tar
d59b494721c87e7536ad6b68d9066b82b55b9697d89239adb56a6ba2878a042d  alpine_3.10.tar
d59b494721c87e7536ad6b68d9066b82b55b9697d89239adb56a6ba2878a042d  library_alpine_3.10.tar
```
Fetch multiple images
```bash
> bin/docker-pull alpine:3.10 ubuntu:18.04 bitnami/redis:5.0
```
Fetch image from private registry
```bash
> bin/docker-pull --user username --password 'P@$$w0rd' private-registry.mydomain.com/my_image:1.2.3
```
