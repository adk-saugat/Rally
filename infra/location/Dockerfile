FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY gen gen
COPY shared shared
COPY services/location services/location
RUN CGO_ENABLED=0 go build -o /out/server ./services/location/cmd/server

FROM gcr.io/distroless/static-debian12:nonroot AS server
COPY --from=build /out/server /server
ENTRYPOINT ["/server"]
