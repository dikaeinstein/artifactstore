FROM golang:1.16-alpine as build
WORKDIR /app
# Download necessary Go modules
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . ./
RUN go build -o /artifactstore-server cmd/artifactstore-server.go

FROM golang:1.16-alpine
COPY --from=build /artifactstore-server /artifactstore-server
USER 1000
CMD ["/artifactstore-server"]
