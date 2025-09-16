# Petri
### Currently developement not for usage
**Petri — something bigger than code.**  

![Petri Logo](logo.png)

## Installation

- Install go
- Install electrum
- Install redis
- Install postgresql
- Set up electrum to JSON-RPC
- Edit config.yaml
- Run `go build`
- Set up to /usr/local/bin (optional)
- Run `./app --config config.yaml`

### Docker 
- Edit Dockerfile, Dockerfile.electrum, config.yaml, .env, docker-compose.yml
- run `docker-compose build`
- run `docker-compose up`
- `docker ps`
- `docker exec -it $DOCKER_CONTAINER_ID`
