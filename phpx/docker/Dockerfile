FROM alpine:latest

ADD ./bin/phpx-linux-amd64 /usr/local/bin/phpx

# The two binaries have to work the same way in use, and a compose file or a
# script written against phpscript runs here unchanged.
RUN ln -s /usr/local/bin/phpx /usr/local/bin/phpscript

RUN /usr/local/bin/phpx version

WORKDIR /app
ENTRYPOINT ["phpx"]
CMD ["server"]
