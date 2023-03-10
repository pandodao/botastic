# Installatin

## Requirements

Please make sure you have installed the following services:

- Milvus: follow the [install guide](./install-mivus.md) to install milvus.
- Tokencal: follow the [install guide](./install-tokencal.md) to install tokencal.

## Edit botastic config file

### Config dependencies

The configuration file of botastic contains the connection addresses for the above services, milvus and tokencal.

```yaml
milvus:
  address: "localhost:19530"

tokencal:
  addr: "http://localhost:9092"
```

## Initalize Storage and Database

Run the following command to initialize Milvus: 

```bash
botastic migrate milvus
```

Run the following command to initialize Postgres:

```bash
botastic migrate 
```

