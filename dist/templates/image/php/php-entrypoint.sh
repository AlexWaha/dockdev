#!/bin/sh

# Install Composer dependencies
if [ ! -d "/var/www/html/vendor" ]; then
    composer install --no-interaction --prefer-dist --optimize-autoloader
fi

groupadd -g 1000 laravel
useradd laravel -g laravel -s /bin/sh

chown -R laravel:laravel /var/www/html
chmod 777 -R /var/www/html

touch /var/www/html/storage/logs/laravel.log
chown laravel:laravel /var/www/html/storage/logs/laravel.log
chown laravel:laravel /var/www/html/vendor/

chmod -R 777 /var/www/html/storage/logs/laravel.log

php-fpm