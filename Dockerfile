FROM golang:1.7-alpine

WORKDIR /go/src/app

COPY . /go/src/app

RUN \
	apk add --no-cache git && \
	go-wrapper download && \
	go-wrapper install -ldflags "-X github.com/minio/minio/cmd.Version=2017-02-06T11:45:39Z -X github.com/minio/minio/cmd.ReleaseTag=RELEASE.2017-02-06T11-45-39Z -X github.com/minio/minio/cmd.CommitID=94867e7d642b7eec01166a60c292fc01912b0190" && \
	mkdir -p /export/docker && \
	rm -rf /go/pkg /go/src && \
	apk del git

EXPOSE 9000
ENTRYPOINT ["minio"]
VOLUME ["/export"]
