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

########################################################

FROM google/cloud-sdk:366.0.0-alpine

COPY --from=build /cuebe/bin/cuebe /usr/bin/cuebe
# cuebe needs both go and cue binaries
COPY --from=build /go/bin/cue /usr/bin/cue
COPY --from=build /usr/local/go/bin/go /usr/bin/go

ENTRYPOINT ["cuebe"]
