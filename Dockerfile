FROM scratch
# local carina must be the Linux binary, rely on the Makefile for this
ADD bin/carina-linux-amd64 /carina
ENTRYPOINT ["/carina"]
CMD ["--version"]
