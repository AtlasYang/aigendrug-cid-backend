FROM postgres:16-alpine

COPY ./init.sql /docker-entrypoint-initdb.d/init.sql

EXPOSE 5432

CMD ["postgres"]