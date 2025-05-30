FROM stevekieu/golang-script:20220602 AS BUILD_BASE
#FROM localhost/build-golang-ubuntu20:20210807-1 AS BUILD_BASE
# You can use the standard golang:alpine but then uncomment the apk below to install sqlite3 depends
# The above image is just a cache image of golang:alpine to save download time
RUN mkdir /app && mkdir /imagetmp && chmod 1777 /imagetmp
    # apk add musl-dev gcc sqlite-dev
ADD . /app/
WORKDIR /app
ENV CGO_ENABLED=1 PATH=/usr/local/go/bin:/opt/go/bin:/usr/bin:/usr/sbin:/bin:/sbin

ARG APP_VERSION
RUN go build -trimpath -ldflags="-X main.version=v1.0 -extldflags=-static -w -s" --tags "osusergo,netgo,sqlite_stat4,sqlite_foreign_keys,sqlite_json"
CMD ["/app/webnote-go"]

FROM scratch
# the ca files is from my current ubuntu 20 /etc/ssl/certs/ca-certificates.crt - it should provide all current root certs
ADD ca-certificates.crt /etc/ssl/certs/
COPY --from=BUILD_BASE /app/webnote-go /webnote-go
COPY --from=BUILD_BASE /imagetmp /tmp
COPY --from=BUILD_BASE /app/assets /assets
ENV TZ=Australia/Brisbane
EXPOSE 80
ENTRYPOINT [ "/webnote-go" ]