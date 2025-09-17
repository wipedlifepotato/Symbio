FROM php:8.2-apache

# Настройка Apache
COPY apache-custom.conf /etc/apache2/sites-available/000-default.conf

# Установка зависимостей для Composer и PHP расширений
RUN apt-get update && apt-get install -y \
        unzip \
        git \
        curl \
    && curl -sS https://getcomposer.org/installer | php \
    && mv composer.phar /usr/local/bin/composer

# Копирование проекта
COPY ./petri-frontendphp /var/www/html/

# Установка зависимостей PHP через Composer
RUN cd /var/www/html/ && COMPOSER_ALLOW_SUPERUSER=1 composer install 

# Права на папки
RUN mkdir -p /var/log/apache2 \
    && chown -R www-data:www-data /var/log/apache2 \
    && chown -R www-data:www-data /var/www/html
