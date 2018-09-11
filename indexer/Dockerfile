FROM golang:1.11

# Copy the sources
COPY . /go/src/github.com/SergeyShpak/HNSearch/indexer
WORKDIR /go/src/github.com/SergeyShpak/HNSearch/indexer

RUN mkdir -p /build

# Build the go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/HNIndexer .

COPY config_k8s.json /etc/HNIndexer/config.json
COPY get-dump.sh /etc/get-dump.sh
COPY exec.sh /etc/exec.sh

RUN chmod +x /etc/exec.sh && chmod +x /etc/get-dump.sh

CMD ["bash", "exec.sh"]