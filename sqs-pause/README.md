# sqs-pause
localstack
```bash
docker-compose up -d
```

Command
```bash
curl -X POST http://localhost:18080/pause

curl -X POST http://localhost:18080/reset

curl -X POST http://localhost:18080/resume
```