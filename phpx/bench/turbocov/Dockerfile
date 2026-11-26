# The turbocov benchmark image: php 8.5 with turbocov and xdebug built from
# source, composer, and the commonmark suite turbocov's own benchmark runs
# (thephpleague/commonmark 2.10.7, 4,447 tests). The phpx binary is mounted at
# run time, so the image holds one stack and the comparison binds no php build.
FROM php:8.5-cli AS builder

RUN apt-get update && apt-get install -y --no-install-recommends libxml2-dev \
    && rm -rf /var/lib/apt/lists/*

COPY turbocov /src/turbocov
WORKDIR /src/turbocov
RUN phpize && ./configure --enable-turbocov && make -j"$(nproc)" && make install
RUN pecl install xdebug-3.5.3

FROM php:8.5-cli

RUN apt-get update && apt-get install -y --no-install-recommends git unzip libxml2 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/local/lib/php/extensions/ /usr/local/lib/php/extensions/
# turbocov refuses to load beside xdebug, so only turbocov is enabled here and
# the xdebug run names its extension with -d.
RUN echo "extension=turbocov.so" > /usr/local/etc/php/conf.d/turbocov.ini
COPY --from=composer:2 /usr/bin/composer /usr/bin/composer

# The suite turbocov benchmarks, at the tag its results name.
RUN git clone --depth 1 --branch 2.10.7 https://github.com/thephpleague/commonmark.git /app \
    && cd /app \
    && composer require --dev --no-interaction shipmonk/turbocov-phpunit \
    && composer install --no-interaction
COPY phpunit-turbocov.xml phpunit-xdebug.xml phpunit-plain.xml /app/
# The phpx leg runs the same checkout: the binary is staged in the context
# beside the Dockerfile, because the daemon cannot mount a host path.
COPY phpx /usr/local/bin/phpx
COPY phpx-shim.sh /app/phpx-shim.sh
RUN chmod +x /app/phpx-shim.sh
WORKDIR /app
