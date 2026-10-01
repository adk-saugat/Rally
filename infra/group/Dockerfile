FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY gen gen
COPY shared shared
COPY services/group services/group
RUN CGO_ENABLED=0 go build -o /out/server ./services/group/cmd/server

FROM golang:1.26 AS goose
RUN CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose@v3.26.0

FROM gcr.io/distroless/static-debian12:nonroot AS migrate
COPY --from=goose /go/bin/goose /goose
COPY services/group/migrations /migrations
ENV GOOSE_DRIVER=postgres
ENV GOOSE_MIGRATION_DIR=/migrations
ENTRYPOINT ["/goose"]
CMD ["up"]

FROM gcr.io/distroless/static-debian12:nonroot AS server
COPY --from=build /out/server /server
ENTRYPOINT ["/server"]
