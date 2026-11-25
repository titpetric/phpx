# The testcov image: the phpx binary the host built and the fixture tree the
# coverage pipeline runs. The checkout is staged into the build context, so the
# image is self-contained and no host path is bind mounted: the docker daemon
# need not see this workspace.
#
# CGO_ENABLED=0 makes the binary static, which is what lets alpine run it.
FROM alpine:3.22
RUN apk add --no-cache bash coreutils tzdata
COPY bin/phpx /usr/local/bin/phpx
COPY bench.sh /usr/local/bin/bench
COPY .env.testing /app/.env.testing
COPY tests /app/tests
COPY phpx/testdata /app/phpx/testdata
WORKDIR /app
ENTRYPOINT ["/usr/local/bin/bench"]
