export STORAGE_VOLUME=aigendrug-cid-storage-volume
export STORAGE_PORT=6802
export STORAGE_CONSOLE_PORT=6803
export MINIO_ROOT_USER=admin
export MINIO_ROOT_PASSWORD=password

docker volume create $STORAGE_VOLUME
docker run --name aigendrug-cid-storage \
    -v $STORAGE_VOLUME:/data \
    -e MINIO_ROOT_USER=$MINIO_ROOT_USER \
    -e MINIO_ROOT_PASSWORD=$MINIO_ROOT_PASSWORD \
    -p $STORAGE_PORT:9000 \
    -p $STORAGE_CONSOLE_PORT:9001 \
    --network aigendrug-cid-network \
    -d \
    minio/minio server /data --console-address ":9001"
