FROM golang:1.7-alpine

WORKDIR /go/src/app

COPY . /go/src/app

RUN \
	apk add --no-cache git && \
	go-wrapper download && \
	go-wrapper install -ldflags "-X github.com/minio/minio/cmd.Version=2017-01-28T13:24:41Z -X github.com/minio/minio/cmd.ReleaseTag=RELEASE.2017-01-28T13-24-41Z -X github.com/minio/minio/cmd.CommitID=4c88787c114b29112c48e4edc5e9bbbf423e5c48" && \
	mkdir -p /export/docker && \
	rm -rf /go/pkg /go/src && \
	apk del git

EXPOSE 9000
ENTRYPOINT ["minio"]
VOLUME ["/export"]
