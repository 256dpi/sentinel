FROM golang:alpine AS build
RUN apk add --no-cache git mercurial
ADD . /src
RUN cd /src && go build -o sentinel

FROM alpine:latest
COPY --from=build /src/sentinel /usr/bin/
CMD ["sentinel"]
