FROM golang AS build
WORKDIR /app
ADD go.* ./
RUN go mod download
ADD . .
RUN CGO_ENABLED=0 go build -trimpath ./cmd/dispatcher

FROM alpine
COPY --from=build /app/dispatcher /bin/
ENTRYPOINT ["/bin/dispatcher"]
