FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o issue-upvote-board .

FROM gcr.io/distroless/static-debian13:nonroot

WORKDIR /app

COPY --from=builder /app/issue-upvote-board .

EXPOSE 8080

ENTRYPOINT ["/app/issue-upvote-board"]
