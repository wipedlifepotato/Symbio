# Petri

> **Currently in development, not for usage**  
> Petri — something bigger than code.

![Petri Logo](logo.png)

---

## Installation

### Prerequisites

- [Go](https://golang.org/dl/)
- [Electrum](https://electrum.org/)
- [Redis](https://redis.io/download)
- [PostgreSQL](https://www.postgresql.org/download/)

### Setup

1. Configure Electrum for JSON-RPC access.
2. Edit `config.yaml` to match your environment.
3. Build the Go app:
    ```bash
    go build
    ```
4. Optionally, move the binary to `/usr/local/bin`:
    ```bash
    sudo mv ./app /usr/local/bin/
    ```
5. Run the application:
    ```bash
    ./app --config config.yaml
    ```

---

## Docker Setup

1. Edit the following files as needed:  
   `Dockerfile`, `Dockerfile.electrum`, `config.yaml`, `.env`, `docker-compose.yml`.

2. Build Docker containers:
    ```bash
    docker-compose build
    ```

3. Create a new Electrum wallet:

    ```bash
    docker ps
    # Example output:
    # CONTAINER ID   IMAGE                   STATUS           PORTS                      NAMES
    # 109364a43052   petri-electrum-server   Up 30s           127.0.0.1:7777->7777/tcp   petri-electrum-server-1
    # d9f3142a8025   petri-php               Up 30s           0.0.0.0:8585->80/tcp       mfreelance-php
    # ...

    docker exec -it 109364a43052 /bin/sh

    # Inside the container:
    /electrum # electrum create
    # Daemon not running; try 'electrum daemon -d'
    # To run this command without a daemon, use --offline

    /electrum # electrum --testnet create
    # Error: Forbidden

    /electrum # electrum --testnet create --rpcuser Electrum --rpcpassword Electrum
    {
        "msg": "Please keep your seed in a safe place; if you lose it, you will not be able to restore your wallet.",
        "path": "/root/.electrum/testnet/wallets/default_wallet",
        "seed": "..."
    }
    ```

4. Start all containers:
    ```bash
    docker-compose up
    ```

5. Check running containers:
    ```bash
    docker ps
    ```

6. Enter a container if needed:
    ```bash
    docker exec -it $DOCKER_CONTAINER_ID /bin/bash
    ```

---

## Notes

- Keep your Electrum seed in a safe place. Losing it will make wallet recovery impossible.
- This project is under active development and not intended for production use.
- Follow Docker logs to troubleshoot services:
    ```bash
    docker-compose logs -f
    ```

## License

Petri is released under two types of licenses:

### Non-Commercial License
Petri is free to use for **personal, educational, or research purposes only**.  
Commercial use is **not allowed** without a separate commercial license.  

See [LICENSE](./LICENSE) for full details.

### Commercial License
For **commercial use** of Petri, please contact:

**Email:** existentialglue@proton.me  

See [LICENSE_COMMERCIAL.txt](./LICENSE_COMMERCIAL.txt) for more information.
