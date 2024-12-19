# aigendrug-cid-backend

2024-2 Creative Integrated Design, Team F

### Architecture

The `aigendrug-cid-backend` comprises six components that implement the Aigendrug web service:

- **main-database:** PostgreSQL database for the web service
- **main-storage:** Minio Object Store for managing protein data
- **web-server:** Nest.js web server
- **ml-coordinator:** Go Gin HTTP Server for managing weight files and Torch workers
- **torch-worker:** Cluster of Flask HTTP Servers for executing PyTorch functions
- **kafka-service (broker, zookeeper):** Kafka services for handling training and inference of ML models

### Get Started

Ensure that [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/install/linux/) are installed on your machine.

Run the following command to complete the setup. For now, it simply creates a bridge network in your Docker engine.

```sh
sh ./setup.sh
```

Each service directory contains a `run.sh` script that sets up environment variables and runs the Docker container for that service. Navigate to each directory and execute the following command:

```sh
sh ./run.sh
```

Make sure to replace the environment variables with values appropriate for your machine.

⚠️ **Important**: The kafka-service must be executed before starting any other services.

### Contact

If you encounter any issues or have questions, feel free to reach out to [me](mailto:atlas.yang3598@gmail.com) at any time.
