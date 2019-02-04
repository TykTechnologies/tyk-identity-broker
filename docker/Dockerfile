FROM debian:jessie-slim

ARG TYKVERSION
ARG REPOSITORY
LABEL Description="Tyk Identity Broker docker image" Vendor="Tyk" Version=$TYKVERSION

RUN apt-get update \
 && apt-get upgrade -y \
 && apt-get install -y --no-install-recommends \
            curl ca-certificates apt-transport-https debian-archive-keyring \
 && curl -L https://packagecloud.io/tyk/$REPOSITORY/gpgkey | apt-key add - \
 && apt-get autoremove -y \
 && rm -rf /root/.cache

RUN echo "deb https://packagecloud.io/tyk/$REPOSITORY/debian/ jessie main" | tee /etc/apt/sources.list.d/tyk_tyk-identity-broker.list \
 && apt-get update \
 && apt-get install -y tyk-identity-broker=$TYKVERSION \
 && rm -rf /var/lib/apt/lists/*

COPY ./tib_sample.conf /opt/tyk-identity-broker/tib.conf

WORKDIR /opt/tyk-identity-broker

CMD ["/opt/tyk-identity-broker/tyk-identity-broker", "-c", "/opt/tyk-identity-broker/tib.conf"]
