FROM --platform=$BUILDPLATFORM golang:1.26@sha256:0f063af2d465d8dcae54cce04278ada488b96f77b42449c8d071e47d016cc65a AS builder

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

FROM gcr.io/distroless/static:nonroot@sha256:f7f8f729987ad0fdf6b05eeeae94b26e6a0f613bdf46feea7fc40f7bd72953e6
WORKDIR /
COPY --from=builder /workspace/kir .
USER 65532:65532

ENTRYPOINT ["/kir"]
