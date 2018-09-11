FROM golang:1.11

# Copy the sources
COPY . /go/src/github.com/SergeyShpak/HNSearch/server
WORKDIR /go/src/github.com/SergeyShpak/HNSearch/server

RUN mkdir -p /build

# Build the go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/HNServer .

COPY config_k8s.json /etc/HNIndexer/config.json

CMD ["/build/HNServer"]