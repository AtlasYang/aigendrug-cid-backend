# aigendrug-cid-backend

2024-2 Creative Integrated Design, Team F

### Architecture

![](https://storage.aigendrug.lighterlinks.io/media/architecture.png)

The aigendrug-cid-backend comprises 6 instances that implement the Aigendrug web service:

- **main-database:** PostgreSQL database for the web service
- **main-storage:** Minio Object Store for managing protein data
- **web-server:** Nest.js web server
- **ml-coordinator:** Go Gin HTTP Server for managing weight files and torch workers
- **torch-worker:** Cluster of Flask HTTP Server for executing PyTorch functions
- **kafka-service(broker, zookeeper):** Kafka services for handling train / inference ML models

### Get Started

Ensure that [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/install/linux/) are installed on your machine.

Next, create a `.env` file in the project root directory using the `.env.template`.

Finally, run the following command in the project root:

```sh
sh ./setup.sh

sh ./run.sh
```

## Contact

If you encounter any issues or have any questions, feel free to reach out to [me](mailto:atlas.yang3598@gmail.com) at any time.
