FROM --platform=$BUILDPLATFORM golang:1.24 AS builder

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum

COPY main.go main.go
COPY cmd/ cmd/
COPY fileutil/ fileutil/
COPY k8s/ k8s/
COPY processor/ processor/
COPY yamlparser/ yamlparser/

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -o kir main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/kir .
USER 65532:65532

ENTRYPOINT ["/kir"]
