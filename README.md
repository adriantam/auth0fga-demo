# auth0fga-demo
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

## API Reference
### Authentication
Every endpoint is protected with bearer token based authentication. Use [jwt.io](https://jwt.io) to craft tokens with a `sub` claim. The token's secret should be `mysecret` for the auth middleware to verify it correctly.

### Folders
```
POST http://localhost:8080/folders
{"name": "folderX"}
```

```
GET http://localhost:8080/folders/:id
```

### Documents
```
POST http://localhost:8080/documents
{"parent": "folder:folderX", "name": "mydoc"}
```

```
GET http://localhost:8080/documents/:id
```

### Groups
```
POST http://localhost:8080/groups
{"name": "engineering", "members": ["jill@auth0.com"]}
```

### Share Object
```
POST http://localhost:8080/share
{"object": "folder:folderX", "relation": "viewer", "user": "group:engineering#member"}
```
