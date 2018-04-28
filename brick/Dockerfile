FROM nginx:alpine

RUN mkdir -p /usr/share/nginx/brick /usr/share/nginx/literallycanvas/static
COPY src /usr/share/nginx/brick
COPY default.conf /etc/nginx/conf.d/
COPY literallycanvas-0.4.14 /usr/share/nginx/literallycanvas/static