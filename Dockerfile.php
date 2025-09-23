FROM php:8.2-apache


COPY apache-custom.conf /etc/apache2/sites-available/000-default.conf


RUN apt-get update && apt-get install -y \
        unzip \
        git \
        curl \
    && curl -sS https://getcomposer.org/installer | php \
    && mv composer.phar /usr/local/bin/composer


COPY ./symbio-php-layer /var/www/html/


RUN cd /var/www/html/ && COMPOSER_ALLOW_SUPERUSER=1 composer install 


RUN mkdir -p /var/log/apache2 \
    && chown -R www-data:www-data /var/log/apache2 \
    && chown -R www-data:www-data /var/www/html
