FROM golang:1.7-alpine

WORKDIR /go/src/app

COPY . /go/src/app

RUN \
	apk add --no-cache git && \
	go-wrapper download && \
	go-wrapper install -ldflags "-X github.com/minio/minio/cmd.Version=2017-02-01T15:31:38Z -X github.com/minio/minio/cmd.ReleaseTag=RELEASE.2017-02-01T15-31-38Z -X github.com/minio/minio/cmd.CommitID=33f63d4b2a785940620c6a85203a8f458c358a84" && \
	mkdir -p /export/docker && \
	rm -rf /go/pkg /go/src && \
	apk del git

EXPOSE 9000
ENTRYPOINT ["minio"]
VOLUME ["/export"]
