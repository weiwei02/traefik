 rm -rf static/ autogen/; make generate-webui
 go generate
 docker build -t hub.17usoft.com/cncf/traefik_proxy:1.0.1 -f conf/docker/Dockerfile ./
 docker push hub.17usoft.com/cncf/traefik_proxy:1.0
 docker tag hub.17usoft.com/cncf/traefik_proxy hub.17usoft.com/cncf/traefik_proxy:1.0