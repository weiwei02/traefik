FROM hub.17usoft.com/base/ubuntu:18.04.2
ENV WORK_DIR /usr/local/opt/traefik
RUN mkdir -p $WORK_DIR
WORKDIR $WORK_DIR
COPY script/ca-certificates.crt /etc/ssl/certs/
COPY dist/traefik $WORK_DIR

COPY conf/compass_test.yml $WORK_DIR/conf/compass_test.yml

#ENTRYPOINT ["./traefik"]