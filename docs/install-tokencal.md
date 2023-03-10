
# Install tokencal

Botastic uses tokencal calculate the token consumption of OpenAI.

## Edit `docker-compose.yaml`

Create a folder named `tokecal` and put a `docker-compose.yaml` file with the following content in it. 

It exposes a port, you can replace it with a port that is easy to manage.

* 9092 - tokencal HTTP service port.

```yaml
version: "3"
services:
  api:
    image: ghcr.io/pandodao/tokencal:latest
    command: uvicorn app.main:app --host 0.0.0.0 --port 80
    ports:
      - 9092:80
```

## Start tokencal

Then start it,

```bash
cd tokencal
docker compose up -d
docker compose ps
```
