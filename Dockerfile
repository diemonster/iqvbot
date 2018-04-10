FROM alpine
RUN apk add --no-cache ca-certificates
ADD ./iqvbot /
CMD ["/iqvbot"]
