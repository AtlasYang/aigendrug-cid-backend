export WEB_SERVER_PORT_BINDING="6801:8080"
export POSTGRES_HOST="aigendrug-cid-db"
export POSTGRES_DB="postgres"
export POSTGRES_USER="admin"
export POSTGRES_PASSWORD="password"
export STORAGE_HOST="aigendrug-cid-storage"
export MINIO_URL="https://storage.aigendrug.lighterlinks.io"
export MINIO_ACCESS_KEY_ID="vCaSMSmvHNiQdBn2JNL6"
export MINIO_SECRET_ACCESS_KEY="SX1V0ULzCEfuHhBeuijUpDHiSh3o2NHgpKxqyW7a"
export KAFKA_SERVER="aigendrug-cid-broker:29092"
export KAFKA_GROUP_ID="aigendrug-web-server"
export KAFKA_CLIENT_ID="aigendrug-web-server"
export KAFKA_TOPICS="ModelInferenceResponse,ModelTrainResponse"

docker build -t aigendrug-cid-web-server .
docker run --name aigendrug-cid-web-server \
    -p $WEB_SERVER_PORT_BINDING \
    -e POSTGRES_HOST=$POSTGRES_HOST \
    -e POSTGRES_DB=$POSTGRES_DB \
    -e POSTGRES_USER=$POSTGRES_USER \
    -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD \
    -e STORAGE_HOST=$STORAGE_HOST \
    -e MINIO_URL=$MINIO_URL \
    -e MINIO_ACCESS_KEY_ID=$MINIO_ACCESS_KEY_ID \
    -e MINIO_SECRET_ACCESS_KEY=$MINIO_SECRET_ACCESS_KEY \
    -e KAFKA_SERVER=$KAFKA_SERVER \
    -e KAFKA_GROUP_ID=$KAFKA_GROUP_ID \
    -e KAFKA_CLIENT_ID=$KAFKA_CLIENT_ID \
    -e KAFKA_TOPICS=$KAFKA_TOPICS \
    --network aigendrug-cid-network \
    -d \
    aigendrug-cid-web-server