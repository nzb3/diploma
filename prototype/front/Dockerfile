FROM nginx:alpine

WORKDIR /usr/share/nginx/html

RUN rm -rf ./*

COPY index.html ./
COPY index.js ./
COPY index.css ./

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
