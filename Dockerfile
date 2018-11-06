FROM alpine

COPY ./resolver /

CMD ["/resolver", "-v"]
