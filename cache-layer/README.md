# cache-layer

## Execution Order

1. First, run docker-compose to start the databases:
```bash
docker-compose up -d --build
```

2. Then, run main.go with the appropriate argument:
   - Using memory cache:
     ```bash
     go run main.go --cache=mem
     ```
   - Using Redis cache:
     ```bash
     go run main.go --cache=redis
     ```

## Verification Method

To verify that the cache is working correctly,  
monitor the PostgreSQL server logs:

```bash
docker logs -f postgres-server
```

Check if "execute" appears only once in the logs.  
If the cache is working properly, "execute" should not appear for the second query.

## Key Point

- DDD interface Type & PKG Type
- package: 
  - github.com/go-gorm/caches/v4
  - github.com/patrickmn/go-cache
