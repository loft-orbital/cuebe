FROM golang:1.17-alpine as build

ENV CGO_ENABLED=0

RUN apk add --no-cache make \
    git

WORKDIR /cuebe

COPY go.mod go.sum ./
RUN go mod download \
 && go mod verify

COPY . .

# Build cuebe
RUN make build

# go install cuelang
RUN go install cuelang.org/go/cmd/cue@v0.4.0

#########################################################
# gcloud cli install
FROM docker:19.03.11 as static-docker-source
FROM alpine:3.15.0 as gcloud-sdk-build

ARG CLOUD_SDK_VERSION=367.0.0
ENV CLOUD_SDK_VERSION=${CLOUD_SDK_VERSION}

RUN apk --no-cache add curl

# Will produce /build/google-cloud-sdk to be copied over the docker image layer
RUN mkdir /build
WORKDIR /build
RUN curl -O https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz
RUN tar xzf google-cloud-sdk-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz && \
    rm google-cloud-sdk-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz

#########################################################
FROM golang:1.17-alpine

# Installing GCP cloud-sdk
ENV PATH /google-cloud-sdk/bin:$PATH

COPY --from=static-docker-source /usr/local/bin/docker /usr/local/bin/docker
COPY --from=gcloud-sdk-build /build/google-cloud-sdk /google-cloud-sdk

# gcloud cli requirements
RUN addgroup -g 1001 -S cloudsdk && \
    adduser -u 1001 -S cloudsdk -G cloudsdk
RUN apk --no-cache add \
        curl \
        python3 \
        py3-crcmod \
        py3-openssl \
        bash \
        libc6-compat \
        openssh-client \
        git \
        gnupg

RUN gcloud config set core/disable_usage_reporting true && \
    gcloud config set component_manager/disable_update_check true && \
    gcloud config set metrics/environment github_docker_image && \
    gcloud --version

RUN git config --system credential.'https://source.developers.google.com'.helper gcloud.sh
VOLUME ["/root/.config"]

# Get cuebe bin from the build layer
COPY --from=build /cuebe/bin/cuebe /usr/bin/cuebe
COPY --from=build /go/bin/cue /usr/bin/cue

WORKDIR /go
ENTRYPOINT ["cuebe"]
