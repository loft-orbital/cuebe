FROM golang:1.17-alpine as build

ENV CGO_ENABLED=0

RUN apk add --no-cache make \
    git

WORKDIR /cuebe

COPY go.mod go.sum ./
RUN go mod download \
 && go mod verify

COPY . .

RUN make build

########################################################

FROM golang:1.17-alpine

RUN go install cuelang.org/go/cmd/cue@v0.4.0

COPY --from=build /cuebe/bin/cuebe /go/bin/cuebe

ENTRYPOINT ["cuebe"]
