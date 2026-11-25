FROM golang:1.27.3 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
ARG COMMIT=unknown
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" -o /server ./cmd/api

FROM scratch
COPY --from=build /server /server
USER 65532:65532
EXPOSE 3003
ENTRYPOINT ["/server"]
