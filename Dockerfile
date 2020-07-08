# Build
FROM prologic/go-builder:latest AS build

# Runtime
FROM golang:alpine

RUN apk --no-cache -U add git build-base

COPY --from=build /src/rss2twtxt /rss2twtxt

ENTRYPOINT ["/rss2twtxt"]
CMD [""]
