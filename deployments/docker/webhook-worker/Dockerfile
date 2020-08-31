FROM golang AS build
WORKDIR /app
ADD go.* ./
RUN go mod download
ADD . .
RUN CGO_ENABLED=0 go build -trimpath ./cmd/webhook-worker

FROM alpine
COPY --from=build /app/webhook-worker /bin/
ENTRYPOINT ["/bin/webhook-worker"]
