# Installatin

## Requirements

Please make sure you have installed the following services:

- Milvus: follow the [install guide](./install-mivus.md) to install milvus.

## Edit botastic config file

### Config dependencies

The configuration file of botastic contains the connection addresses for the milvus service.

```yaml
milvus:
  address: "localhost:19530"
```

## Initalize Storage and Database

Run the following command to initialize Milvus: 

```bash
botastic migrate milvus
```

Run the following command to initialize Postgres:

```bash
botastic migrate up
```
