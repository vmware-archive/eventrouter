# Kafka Sink Integration Test

### Start kafka
```bash
docker-compose up -d kafka
```

### Running the consumer
```bash
# give kafka time to start before running the consumer
docker-compose up consumer
```

### Running the integration test
```bash
# one time test
docker-compose up app

# repeated testing, much faster
docker-compose run --entrypoint=/bin/sh app
$ cd tests/kafka
$ go install
$ go run main.go

```
