# Step 1: Modules caching
FROM golang:1.18-alpine3.14 as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Builder
FROM golang:1.18-alpine3.14 as builder
COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /bin/app ./

# Step 3: Final
FROM scratch
COPY --from=builder /bin/app /app
CMD ["/app"]