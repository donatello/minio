FROM golang:1.7-alpine

WORKDIR /go/src/app

COPY . /go/src/app

RUN \
	apk add --no-cache git && \
	go-wrapper download && \
	go-wrapper install -ldflags "-X github.com/minio/minio/cmd.Version=2017-02-03T04:08:34Z -X github.com/minio/minio/cmd.ReleaseTag=RELEASE.2017-02-03T04-08-34Z -X github.com/minio/minio/cmd.CommitID=ce2410a0b00f8a4ecaf844bdfba6bef8c5c833e7" && \
	mkdir -p /export/docker && \
	rm -rf /go/pkg /go/src && \
	apk del git

EXPOSE 9000
ENTRYPOINT ["minio"]
VOLUME ["/export"]
