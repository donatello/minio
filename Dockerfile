FROM golang:1.7-alpine

WORKDIR /go/src/app

COPY . /go/src/app

RUN \
	apk add --no-cache git && \
	go-wrapper download && \
	go-wrapper install -ldflags "-X github.com/minio/minio/cmd.Version=2017-02-01T15:16:06Z -X github.com/minio/minio/cmd.ReleaseTag=RELEASE.2017-02-01T15-16-06Z -X github.com/minio/minio/cmd.CommitID=27fef28e811db2d1c4ff3392044a98e6fa86e717" && \
	mkdir -p /export/docker && \
	rm -rf /go/pkg /go/src && \
	apk del git

EXPOSE 9000
ENTRYPOINT ["minio"]
VOLUME ["/export"]
