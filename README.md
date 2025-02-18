# Dogbox

Clone of Catbox for instructive purposes.
~
* Uses Docker to spin up a postgres server and nginx reverse proxy
* Options to store files within a filesystem or remote storage
* Returns information about the images like creation and modification date
* Support for deleting images

# Setup

```sh
docker compose -f infra.yaml up # Starts up the backend services
go run . # Runs application
```

Access the server on port 5050 by default.
