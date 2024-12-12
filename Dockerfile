FROM golang:1.23.3-alpine3.20 AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum
COPY vendor/ vendor/
RUN go mod verify

COPY api/ api/
COPY cmd/controller cmd/controller
COPY internal/ internal/
COPY pkg/ pkg/

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o x-pdb cmd/controller/main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/x-pdb .
USER 65532:65532
ENTRYPOINT ["/x-pdb"]
