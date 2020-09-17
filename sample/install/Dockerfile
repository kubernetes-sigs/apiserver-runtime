
FROM ubuntu
RUN apt update -y && apt install fortune fortunes -y && apt clean
COPY ./bin/apiserver /usr/local/bin/apiserver
ENTRYPOINT ["/usr/local/bin/apiserver"]
