FROM golang:1.7-alpine

WORKDIR /go/src/app

COPY . /go/src/app

RUN \
	apk add --no-cache git && \
	go-wrapper download && \
	go-wrapper install -ldflags "-X github.com/minio/minio/cmd.Version=2017-02-06T11:53:25Z -X github.com/minio/minio/cmd.ReleaseTag=RELEASE.2017-02-06T11-53-25Z -X github.com/minio/minio/cmd.CommitID=4446a09ff4a8158108ef3c7e7dcf986bd74da545" && \
	mkdir -p /export/docker && \
	rm -rf /go/pkg /go/src && \
	apk del git

EXPOSE 9000
ENTRYPOINT ["minio"]
VOLUME ["/export"]
