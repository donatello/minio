FROM golang:1.7-alpine

WORKDIR /go/src/app

COPY . /go/src/app

RUN \
	apk add --no-cache git && \
	go-wrapper download && \
	go-wrapper install -ldflags "-X github.com/minio/minio/cmd.Version=2017-02-01T15:51:28Z -X github.com/minio/minio/cmd.ReleaseTag=RELEASE.2017-02-01T15-51-28Z -X github.com/minio/minio/cmd.CommitID=7b1016ca00b89a3c0af0692646be37f68a26b854" && \
	mkdir -p /export/docker && \
	rm -rf /go/pkg /go/src && \
	apk del git

EXPOSE 9000
ENTRYPOINT ["minio"]
VOLUME ["/export"]
