FROM ubuntu:latest

WORKDIR /opt/tib/

RUN \
  apt-get update && \
  apt-get -y upgrade && \
  apt-get install -y bzr golang git

RUN mkdir -p /opt/tib/build/go/src/tyk-identity-broker
COPY . /opt/tib/build/go/src/tyk-identity-broker
COPY tib.conf /opt/tib/
COPY profiles.json /opt/tib/

ENV GOPATH /opt/tib/build/go

RUN \
  cd /opt/tib/build/go/src/tyk-identity-broker && \
  go get && \
  go build && \
  cp /opt/tib/build/go/src/tyk-identity-broker/tyk-identity-broker /opt/tib/tib && \
  rm -rf /opt/tib/build

CMD ["./tib"]
