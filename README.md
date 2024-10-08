# aigendrug-cid-backend

2024-2 Creative Integrated Design, Team F

The aigendrug-cid-backend comprises 5 instances that implement the Aigendrug web service:

- **main-database:** PostgreSQL database for the web service
- **main-storage:** Minio Object Store for managing protein data
- **web-server:** Nest.js web server
- **ml-server** Flask ml model server that supports dynamic weight reload
- **broker, zookeeper** Kafka services for handling train / inference ML models

### Get Started

Ensure that [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/install/linux/) are installed on your machine.

Next, create a `.env` file in the project root directory using the `.env.template`.

Finally, run the following command in the project root:

```sh
docker compose up -d
```

## Contact

If you encounter any issues or have any questions, feel free to reach out to [me](mailto:atlas.yang3598@gmail.com) at any time.
