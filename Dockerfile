FROM golang:1.15 as build-deps

WORKDIR /ws-storage
COPY . /ws-storage

RUN useradd -d /ws-storage -s /bin/bash gen3 && chown -R gen3: /ws-storage
USER gen3

# Populate git version info into the code
RUN echo "package storage\n\nconst (" > storage/gitversion.go \
    && COMMIT=`git rev-parse HEAD` && echo "    gitcommit=\"${COMMIT}\"" >> storage/gitversion.go \
    && VERSION=`git describe --always --tags` && echo "    gitversion=\"${VERSION}\"" >> storage/gitversion.go \
    && echo ")" >> storage/gitversion.go
    
RUN go build -o bin/ws-storage

# Store only the resulting binary in the final image
# Resulting in significantly smaller docker image size
#FROM scratch
#COPY --from=build-deps /ws-storage/ws-storage /ws-storage

CMD ["/ws-storage/bin/ws-storage"]
