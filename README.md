# openfga-demo
This project demonstrates a sample Google Docs API with access control enforced via an integration with [Auth0 FGA](https://fga.dev).

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
export FGA_STORE_ID=<storeID>
export FGA_CLIENT_ID=<clientID>
export FGA_CLIENT_SECRET=<secret>
go run main.go
```
The `FGA_STORE_ID`, `FGA_CLIENT_ID`, and `FGA_CLIENT_SECRET` can be found in the [Auth0 FGA Dashboard](https://dashboard.fga.dev) for your FGA account.
