FROM ruby as template
ARG nginx_template_data=nginx.example.yaml

ADD "nginx.conf.tmpl" ${nginx_template_data} "/build/"
RUN gem install mustache \
      && cd /build \
      && mustache ${nginx_template_data} nginx.conf.tmpl > nginx.conf


FROM nginx:latest
#RUN apt-get update && apt-get install --no-install-recommends --no-install-suggests -y libnginx-mod-stream
COPY --from=template /build/nginx.conf /etc/nginx/nginx.conf


