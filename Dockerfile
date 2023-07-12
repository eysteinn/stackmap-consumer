FROM osgeo/gdal:alpine-small-3.6.2 

RUN apk add postgresql-client
RUN apk add postgis


#COPY bin /app/bin
COPY bin/consume /usr/local/bin
#COPY data /app/data

WORKDIR /app

# Run the binary program produced by `go install`
CMD ["/bin/sh", "-c", "while true; do sleep 1; done"]

#ENTRYPOINT ["/bin/sh", "-c", "while true; do sleep 1; done"]
