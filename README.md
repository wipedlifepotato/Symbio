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
- Create a new wallet for electrum: ```
gitpod /workspace/Petri/petri-frontendphp (main) $ docker ps 
CONTAINER ID   IMAGE                   COMMAND                   CREATED          STATUS                         PORTS                      NAMES
109364a43052   petri-electrum-server   "/bin/sh -c 'electru…"    4 minutes ago    Up 30 seconds                  127.0.0.1:7777->7777/tcp   petri-electrum-server-1
d9f3142a8025   petri-php               "docker-php-entrypoi…"    20 minutes ago   Up 30 seconds                  0.0.0.0:8585->80/tcp       mfreelance-php
2846c860fc65   petri-go-server         "/bin/sh -c 'sh -c \"…"   24 hours ago     Restarting (1) 9 seconds ago                              mfreelance-go
c957c58b4101   postgres:15             "docker-entrypoint.s…"    39 hours ago     Up 30 seconds                  0.0.0.0:5432->5432/tcp     mfreelance-postgres
bd5d682ccb69   redis:7                 "docker-entrypoint.s…"    39 hours ago     Up 30 seconds                  0.0.0.0:6379->6379/tcp     mfreelance-redis
gitpod /workspace/Petri/petri-frontendphp (main) $ docker exec -it 109364a43052 /bin/sh
/electrum # electrum create
Daemon not running; try 'electrum daemon -d'
To run this command without a daemon, use --offline
/electrum # electrum --testnet create
Error: Forbidden
/electrum # electrum --testnet create --rpcuser Electrum --rpcpassword Electrum
{
    "msg": "Please keep your seed in a safe place; if you lose it, you will not be able to restore your wallet.",
    "path": "/root/.electrum/testnet/wallets/default_wallet",
    "seed": "..."
}
/electrum # 
```
- run `docker-compose up`
- `docker ps`
- `docker exec -it $DOCKER_CONTAINER_ID /bin/bash # if you need it`
