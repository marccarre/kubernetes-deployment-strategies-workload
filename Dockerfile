FROM alpine:3.8

# Install certificates for HTTPS/SSL:
RUN apk --no-cache add ca-certificates

# Add & configure a dedicated user, as
# we do not run as root for security reasons:
RUN addgroup -g 1000 -S service && \
    adduser -u 1000 -S service -G service
USER service
WORKDIR /home/service

EXPOSE 80
ENTRYPOINT ["./service"]

# Copy database migrations & binary:
COPY pkg/db/migrations migrations
COPY cmd/service/service .

# Tag the image with potentially useful metadata.
# Some of these labels change for every build, and should therefore be among the last layers of the image:
ARG BUILD_DATE
ARG REVISION
LABEL maintainer="Marc Carré <marc@weave.works>" \
      org.opencontainers.image.vendor="Weaveworks" \
      org.opencontainers.image.title="service" \
      org.opencontainers.image.description="A sample microservice using a PostgreSQL database and migrations" \
      org.opencontainers.image.url="https://github.com/marccarre/kubernetes-deployment-strategies-workload" \
      org.opencontainers.image.source="git@github.com:marccarre/kubernetes-deployment-strategies-workload.git" \
      org.opencontainers.image.revision="${REVISION}" \
      org.opencontainers.image.created="${BUILD_DATE}"
