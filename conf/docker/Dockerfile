FROM hub.17usoft.com/base/ubuntu:18.04.2

ENV WORK_DIR /usr/local/opt/traefik
ENV CONFIG_FILE $WORK_DIR/conf/compass_test.yml
#RUN mkdir -p $WORK_DIR
WORKDIR $WORK_DIR

COPY script/ca-certificates.crt /etc/ssl/certs/
COPY dist/traefik $WORK_DIR
COPY conf $WORK_DIR/conf
COPY static    $WORK_DIR/static

#ENTRYPOINT ["./traefik"]