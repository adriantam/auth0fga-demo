# openfga-demo
This project demonstrates a sample Google Docs API with access control enforced via an integration with [openfga/openfga](https://github.com/openfga/openfga).

## Running
1. Start a Postgres container
```console
docker run -e POSTGRES_PASSWORD=password -p 5432:5432  -d postgres:14
```

2. Bootstrap the database tables
```console
PGPASSWORD=password psql -h localhost -p 5432 -U postgres -d postgres -f schema.sql
```

3. Start the app
```console
go run main.go
```
