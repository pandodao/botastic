# Install Milvus

Botastic uses Milvus as its vector database. To simplify the installation process, we provide a docker-compose file to deploy Milvus.

## Edit `docker-compose.yaml`

Create a folder named `milvus` and put a `docker-compose.yaml` file with the following content in it.

It exposes two ports, which you can replace with easy-to-manage ports. 

* 19530 - milvus gRPC service port.
* 19531 - port for management GUI.


```yaml
version: "3.5"

services:
  etcd:
    container_name: milvus-etcd
    image: quay.io/coreos/etcd:v3.5.5
    environment:
      - ETCD_AUTO_COMPACTION_MODE=revision
      - ETCD_AUTO_COMPACTION_RETENTION=1000
      - ETCD_QUOTA_BACKEND_BYTES=4294967296
      - ETCD_SNAPSHOT_COUNT=50000
    volumes:
      - ${DOCKER_VOLUME_DIRECTORY:-.}/volumes/etcd:/etcd
    command: etcd -advertise-client-urls=http://127.0.0.1:2379 -listen-client-urls http://0.0.0.0:2379 --data-dir /etcd

  standalone:
    container_name: milvus-standalone
    image: milvusdb/milvus:v2.2.3
    command: ["milvus", "run", "standalone"]
    environment:
      ETCD_ENDPOINTS: etcd:2379
    volumes:
      - ${DOCKER_VOLUME_DIRECTORY:-.}/volumes/milvus:/var/lib/milvus
      - ${DOCKER_VOLUME_DIRECTORY:-.}/volumes/milvus_config/milvus.yaml:/milvus/configs/milvus.yaml
    ports:
      - "19530:19530"
      # - "9091:9091"
    depends_on:
      - "etcd"

  attu:
    container_name: milvus-attu
    image: zilliz/attu:latest
    environment:
      MILVUS_URL: standalone:19530
    ports:
      - "19531:3000"
    depends_on:
      - "standalone"

networks:
  default:
    name: milvus

```

## Create the `milvus.yaml` config file.

Execute the folling command to download the `milvus.yaml` file.

```shell
cd milvus
mkdir -p volumes/milvus_config
cd volumes/milvus_config
wget https://raw.githubusercontent.com/milvus-io/milvus/v2.2.3/configs/milvus.yaml
```

The `tree` command output as following,

```shell
cd milvus
tree
.
├── docker-compose.yml
└── volumes
    └── milvus_config
        └── milvus.yaml
```

## Edit the `milvus.yaml` file

Edit the `minio` section of the `milvus.yaml` file to config the S3 storage. 

If you are not sure about the configuration you can refer to MinIO's [documentation](https://docs.min.io/docs/minio-docker-quickstart-guide.html).

```yaml
# ...

minio:
  address: CHANGE_ME # Address of MinIO/S3
  port: 80 # Port of MinIO/S3
  accessKeyID: CHANGE_ME # accessKeyID of MinIO/S3
  secretAccessKey: CHANGE_ME # MinIO/S3 encryption string
  useSSL: false # Access to MinIO/S3 with SSL
  bucketName: CHANGE_ME # Bucket name in MinIO/S3
  rootPath:
    "files" # The root path where the message is stored in MinIO/S3
  useIAM: false
  cloudProvider: "aws"
  iamEndpoint: ""

# ...
```


## Start the service

check the service status to ensure that the service started successfully.

```bash
cd milvus
docker compose up -d
docker compose ps
```

