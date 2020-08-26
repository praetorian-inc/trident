FROM golang AS build
WORKDIR /app
ADD go.* ./
RUN go mod download
ADD . .
RUN go build ./cmd/orchestrator

FROM golang
COPY --from=build /app/orchestrator /bin/
ENTRYPOINT ["/bin/orchestrator"]
