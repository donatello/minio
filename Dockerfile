FROM golang:1.7-alpine

WORKDIR /go/src/app

COPY . /go/src/app

RUN \
	apk add --no-cache git && \
	go-wrapper download && \
	go-wrapper install -ldflags "-X github.com/minio/minio/cmd.Version=2017-02-06T11:40:10Z -X github.com/minio/minio/cmd.ReleaseTag=RELEASE.2017-02-06T11-40-10Z -X github.com/minio/minio/cmd.CommitID=6a45323acc00475c2d513d60ff1f99b9e768c04a" && \
	mkdir -p /export/docker && \
	rm -rf /go/pkg /go/src && \
	apk del git

EXPOSE 9000
ENTRYPOINT ["minio"]
VOLUME ["/export"]
