export MAIN_DB_PORT_BINDING="6804:5432"
export POSTGRES_DB="postgres"
export POSTGRES_USER="admin"
export POSTGRES_PASSWORD="password"

docker volume create aigendruc-cid-db-volume
docker build -t aigendrug-cid-db .
docker run -d --name aigendrug-cid-db \
    --network aigendrug-cid-network \
    -p $MAIN_DB_PORT_BINDING \
    -v aigendruc-cid-db-volume:/var/lib/postgresql/data \
    -e POSTGRES_DB=$POSTGRES_DB \
    -e POSTGRES_USER=$POSTGRES_USER \
    -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD \
    -d \
    aigendrug-cid-db