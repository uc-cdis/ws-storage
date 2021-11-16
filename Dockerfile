FROM quay.io/cdis/golang:1.17-bullseye as build-deps

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# Hopefully someday in the future, this will be updated to provide
# the consistent Docker images for Go with consistent paths and structure
# WORKDIR $GOPATH/src/github.com/uc-cdis/ws-storage/
WORKDIR /ws-storage

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN GITCOMMIT=$(git rev-parse HEAD) \
    GITVERSION=$(git describe --always --tags) \
    && go build \
    -ldflags="-X 'github.com/uc-cdis/ws-storage/storage/version.GitCommit=${GITCOMMIT}' -X 'github.com/uc-cdis/ws-storage/storage/version.GitVersion=${GITVERSION}'" \
    -o bin/ws-storage

FROM scratch
COPY --from=build-deps /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# COPY --from=build-deps /ws-storage /ws-storage
COPY --from=build-deps /ws-storage/bin/ws-storage /ws-storage/bin/ws-storage
CMD ["/ws-storage/bin/ws-storage"]
