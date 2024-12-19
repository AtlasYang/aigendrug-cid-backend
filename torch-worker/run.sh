docker build -t aigendrug-cid-torch-worker -f ./worker/Dockerfile ./worker
docker compose up --build -d