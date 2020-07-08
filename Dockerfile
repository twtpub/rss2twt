# Build
FROM prologic/go-builder:latest AS build

# Runtime
FROM alpine:latest

RUN apk --no-cache -U add ca-certificates

WORKDIR /
VOLUME /feeds

COPY .dockerfiles/config.yaml /config.yaml
COPY --from=build /src/rss2twtxt /rss2twtxt

ENTRYPOINT ["/rss2twtxt"]
CMD [""]
