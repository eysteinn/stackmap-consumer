FROM osgeo/gdal:alpine-small-3.6.2 as build_base

ENV CGO_ENABLED=1
ENV GO111MODULE=on
RUN apk add --no-cache git  git gcc g++
RUN apk add --no-cache git make musl-dev go
RUN apk add --no-cache pkgconfig
# Set the Current Working Directory inside the container
WORKDIR /app/src
ENV PATH /go/bin:$PATH
COPY src/ /app/src/
RUN go build -o main main.go
#CMD ./main
#CMD exec /bin/sh -c "trap : TERM INT; sleep 9999999999d & wait"

# We want to populate the module cache based on the go.{mod,sum} files.
#COPY go.mod .
#COPY go.sum .
#RUN go mod download

#COPY . .

# Build the Go app
#RUN go build -o ./out/app ./cmd/api/main.go


# Start fresh from a smaller image
#FROM alpine:3.12
FROM osgeo/gdal:alpine-small-3.6.2
#RUN apk add ca-certificates

#WORKDIR /app

COPY --from=build_base /app/src/main /usr/local/bin/stackmap-consumer
CMD stackmap-consumer
#oCOPY --from=build_base /src/data /app/data

#RUN chmod +x restapi

# This container exposes port 8080 to the outside world
#EXPOSE 3000

# Run the binary program produced by `go install`
#ENTRYPOINT ./restapi


