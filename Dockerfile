FROM scratch
# local carina must be the Linux binary, rely on the Makefile for this
ADD ca-certificates.crt /etc/ssl/certs/
ADD carina-linux /carina
ENTRYPOINT ["/carina"]
CMD ["--version"]
