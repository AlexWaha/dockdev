FROM php:8.3-fpm-bookworm AS php

WORKDIR /var/www/html

COPY php-entrypoint.sh /usr/local/bin/php-entrypoint.sh
RUN chmod +x /usr/local/bin/php-entrypoint.sh

# Install necessary libraries and PHP extensions
RUN apt-get update && apt-get upgrade -y && \
    apt-get install -y --no-install-recommends \
        $PHPIZE_DEPS \
        cmake \
        wget \
        libfreetype6-dev \
        libfontconfig1-dev \
        libpng-dev \
        libjpeg-dev \
        libc-dev \
        jpegoptim optipng pngquant gifsicle \
        unzip \
        curl \
        libzip-dev \
        libpq-dev \
        pkg-config \
        libssl-dev \
        zlib1g zlib1g-dev \
        libxml2-dev \
        libonig-dev \
        libmemcached-dev \
        libgmp-dev \
        default-mysql-client \
        libsqlite3-dev && \
    docker-php-ext-configure gd --with-freetype=/usr/include/ --with-jpeg=/usr/include/ && \
    docker-php-ext-install \
        bcmath \
        gd \
        pcntl \
        pdo_mysql \
        pdo_sqlite \
        mbstring \
        mysqli \
        exif \
        zip \
        intl \
        soap \
        gmp \
        opcache && \
    pecl install redis pcov && docker-php-ext-enable redis pcov && \
    curl -sS https://getcomposer.org/installer | php -- --install-dir=/usr/local/bin --filename=composer && \
    /usr/local/bin/composer config --global repo.packagist composer https://packagist.org && \
    apt-get remove --purge -y $PHPIZE_DEPS \
    && rm -rf /tmp/* \
    && apt -y autoremove \
    && apt-get -y clean \
    && rm -rf /var/lib/apt/lists/* \
    \
    && rm -rf /var/lib/apt/lists/*
    
# Set permissions for Composer
RUN mkdir -p /.composer/ && chmod -R 777 /.composer/

# Expose PHP-FPM port
EXPOSE 9000

# Start PHP-FPM
CMD ["php-fpm"]