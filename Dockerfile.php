FROM php:8.2-apache

COPY apache-custom.conf /etc/apache2/sites-available/000-default.conf

RUN mkdir -p /var/log/apache2 && chown -R www-data:www-data /var/log/apache2 && chown www-data:www-data -R /var/www/html
