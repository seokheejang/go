# fiber-test

```bash
docker compose up -d
```

api test
```bash
curl http://localhost:3000/api

curl http://localhost:3000/api/health
```

db test
```bash
curl -X POST "http://localhost:3000/api/db/123" \
     -H "Content-Type: application/json" \
     -d '{"value": "Hello, MongoDB!"}'

curl -X GET "http://localhost:3000/api/db/123"
```
