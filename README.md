# kajappka-rating-service

This application is game rating service working as a part of Kajappka
game rating tool. This application provides three endpoints:

* `GET /` which returns average rating for all games
* `GET /{id}` which returns rating for a game `{id}` and a current user
* `PUT /{id}` which updates rating for a game `{id}` and a current user

The following environment variables can be used to configure the service,
some of those are required.

* `HOST` - the host which the verifier listens on, defaults to
  `0.0.0.0`,
* `PORT` - the port which the verifier listens on, defaults to `8080`,
* `VERIFIER_URI` - the URL of authentication service, required,
   e.g. `http://localhost:8000/success`
* `MONGO_CONNECTION_STRING` - the MongoDB connection string, required,
   e.g. `mongodb://localhost:27017/`
* `MONGO_DB_NAME` - the MongoDB database name, required (e.g. `ratings`)
* `MONGO_COLLECTION_NAME` - the MongoDB collection name,
   required (e.g. `ratings`)

## Development

For development, you can use provided Docker Compose startup file as follows:

```
$ docker-compose -f docker/development/docker-compose.yml up
```

it boots the MongoDB container, the Mock Verifier container and the rating
service itself for local testing and exposes the service at `http://localhost:8080`

## Production image

The production Docker image can be built with

```
$ docker build -f docker/production/Dockerfile -t kajappka-rating-service:latest .
```
