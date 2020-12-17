FROM ruby as template
ARG nginx_template_data=nginx.example.yaml

RUN    echo "---" \
    && echo "template_file: ${nginx_template_data}" \
    && echo "---"
 
ADD "./nginx.tmpl.conf" ${nginx_template_data} "/build/"
RUN gem install mustache \
      && cd /build \
      && cat ${nginx_template_data} \
      && mustache ${nginx_template_data} nginx.tmpl.conf > nginx.conf


FROM nginx:latest
#RUN apt-get update && apt-get install --no-install-recommends --no-install-suggests -y libnginx-mod-stream
COPY --from=template /build/nginx.conf /etc/nginx/nginx.conf


