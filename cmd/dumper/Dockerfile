FROM golang:1.21.1

LABEL org.opencontainers.image.source='https://github.com/Jacobbrewer1/dumpster'
LABEL org.opencontainers.image.description="This is a Database Dumping app written in Go that creates a dump of a database."
LABEL org.opencontainers.image.licenses='GNU General Public License v3.0'

WORKDIR /dumpster

# Copy the binary from the build
COPY ./bin/linux /dumpster/app

RUN ["chmod", "+x", "./app"]

ENTRYPOINT ["/dumpster/app"]
