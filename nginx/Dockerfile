FROM ruby as template

ADD "nginx.conf.tmpl" "nginx.yaml" "/build/"
RUN gem install mustache \
      && cd /build \
      && mustache nginx.yaml nginx.conf.tmpl > nginx.conf


FROM nginx:latest
#RUN apt-get update && apt-get install --no-install-recommends --no-install-suggests -y libnginx-mod-stream
COPY --from=template /build/nginx.conf /etc/nginx/nginx.conf


