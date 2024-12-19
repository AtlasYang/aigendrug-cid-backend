export WEIGHT_VOLUME_MOUNT="./.weights:/app/weights"
export CSV_VOLUME_MOUNT="./.csv:/app/csv"
export POSTGRES_CONNECTION="postgres://admin:password@aigendrug-cid-db:5432/postgres"
export MINIO_ENDPOINT="aigendrug-cid-storage:9000"
export MINIO_ACCESS_KEY_ID="vCaSMSmvHNiQdBn2JNL6"
export MINIO_SECRET_ACCESS_KEY="SX1V0ULzCEfuHhBeuijUpDHiSh3o2NHgpKxqyW7a"
export KAFKA_BROKER_HOST="aigendrug-cid-broker:29092"
export KAFKA_CONSUMER_GROUP="aigendrug-cid-ml-coordinator"
export KAFKA_CONSUMER_TOPICS="ModelInferenceRequest,ModelTrainRequest,ModelInitializeRequest,ModelProcessRequest"
export WORKER_COUNT=10
export TORCH_WORKER_COUNT=3
export WEIGHT_BUCKET="weights"
export WEIGHT_CACHE_SIZE=10

docker build -t aigendrug-cid-ml-coordinator .
docker run --name aigendrug-cid-ml-coordinator \
    -p 6809:8080 \
    -v $WEIGHT_VOLUME_MOUNT \
    -v $CSV_VOLUME_MOUNT \
    -e POSTGRES_CONNECTION=$POSTGRES_CONNECTION \
    -e MINIO_ENDPOINT=$MINIO_ENDPOINT \
    -e MINIO_ACCESS_KEY_ID=$MINIO_ACCESS_KEY_ID \
    -e MINIO_SECRET_ACCESS_KEY=$MINIO_SECRET_ACCESS_KEY \
    -e KAFKA_BROKER_HOST=$KAFKA_BROKER_HOST \
    -e KAFKA_CONSUMER_GROUP=$KAFKA_CONSUMER_GROUP \
    -e KAFKA_CONSUMER_TOPICS=$KAFKA_CONSUMER_TOPICS \
    -e WORKER_COUNT=$WORKER_COUNT \
    -e TORCH_WORKER_COUNT=$TORCH_WORKER_COUNT \
    -e WEIGHT_BUCKET=$WEIGHT_BUCKET \
    -e WEIGHT_CACHE_SIZE=$WEIGHT_CACHE_SIZE \
    --network aigendrug-cid-network \
    -d \
    aigendrug-cid-ml-coordinator