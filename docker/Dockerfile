FROM golang:1.12
WORKDIR /opt/pbm
COPY . .
RUN make install

FROM registry.access.redhat.com/ubi7/ubi-minimal

LABEL org.opencontainers.image.title="Percona Backup for MongoDB"
LABEL org.opencontainers.image.vendor="Percona"
LABEL org.opencontainers.image.description="Percona Backup for MongoDB is a distributed, \
low-impact solution for achieving consistent backups of MongoDB Sharded Clusters and Replica Sets."
LABEL org.opencontainers.image.authors="info@percona.com"

COPY LICENSE /licenses/

COPY --from=0 /go/bin/pbm /go/bin/pbm-agent /usr/local/bin/

USER nobody

CMD ["pbm-agent"]
