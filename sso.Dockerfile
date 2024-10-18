FROM golang:latest AS build
WORKDIR /app
COPY . /app/
RUN GOOS=linux go build -ldflags '-extldflags "-static"' -o main ./sso/cmd/sso/main.go

FROM scratch
WORKDIR /app
COPY --from=build /app/main /app/sso/config/demo.yaml /app/
COPY --from=build  /app/commands/proto/sso/gen /app/commands/proto/sso/gen
ENV CONFIG_PATH="/app/demo.yaml"
ENTRYPOINT ["/app/main"]


