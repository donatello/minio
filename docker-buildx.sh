#!/bin/bash

sudo sysctl net.ipv6.conf.all.disable_ipv6=0

release=$(git describe --abbrev=0 --tags)

echo "Current commit is:"
git log -n 1
echo "Building container images for release ${release}"

set -e
read -p "Press [Enter] to continue or Ctrl+C to abort." -r
set +e


docker buildx build --push --no-cache \
	--build-arg RELEASE="${release}" \
	-t "minio/minio:latest" \
	-t "quay.io/minio/minio:latest" \
	-t "minio/minio:${release}" \
	-t "quay.io/minio/minio:${release}" \
	--platform=linux/arm64,linux/amd64,linux/ppc64le,linux/s390x \
	-f Dockerfile.release .

docker buildx prune -f

docker buildx build --push --no-cache \
	--build-arg RELEASE="${release}" \
	-t "minio/minio:${release}-cpuv1" \
	-t "quay.io/minio/minio:${release}-cpuv1" \
	--platform=linux/arm64,linux/amd64,linux/ppc64le,linux/s390x \
	-f Dockerfile.release.old_cpu .

docker buildx prune -f

docker buildx build --push --no-cache \
	--build-arg RELEASE="${release}" \
	-t "minio/minio:${release}.fips" \
	-t "quay.io/minio/minio:${release}.fips" \
	--platform=linux/amd64 -f Dockerfile.release.fips .

docker buildx prune -f

sudo sysctl net.ipv6.conf.all.disable_ipv6=0
