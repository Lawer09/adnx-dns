# adnx_dns

GoDaddy DNS management service for adnx_dns.

## Features

- Sync root domains from GoDaddy
- Query available / unavailable domains
- Query available domains with binding details
- Create or update A records by IPv4
- Optional unique mode for IPv4 (keep only one bound fqdn)
- Query all domains bound to an IPv4
- Unbind by IPv4 + fqdn
- Enable / disable root domains locally
- Unified response format and error codes

## API Authentication

All `/api/v1/*` requests require:

`X-API-Token: <API_TOKEN>`

## Response Format

```json
{"code":0,"data":{},"message":"ok"}
```

## Error Codes

- `0` success
- `1001` invalid api token
- `1002` invalid parameter
- `1003` invalid ipv4
- `1004` domain not found
- `1005` domain unavailable
- `1006` no available domain
- `1008` fqdn not found
- `1010` ip has no bound domains
- `1011` domain already disabled
- `1012` domain already enabled
- `1014` GoDaddy rate limited
- `1015` GoDaddy provider error
- `1016` database error
- `1020` fqdn and ip do not match

## Endpoints

- `GET /api/v1/domains/available`
- `GET /api/v1/domains/available/detail`
- `GET /api/v1/domains/unavailable`
- `POST /api/v1/records/resolve`
- `GET /api/v1/records/by-ip?ipv4=x.x.x.x`
- `POST /api/v1/records/unbind`
- `POST /api/v1/domains/disable`
- `POST /api/v1/domains/enable`
- `POST /api/v1/domains/sync`

## Build

```bash
go mod tidy
go build -o adnx_dns ./cmd/server
```

## Database

Import `schema.sql` into MySQL database `adnx_dns`.

## Notes

The app prefers `MYSQL_DSN`. If missing, it will build a DSN from `MYSQL_HOST`, `MYSQL_PORT`, `MYSQL_USER`, `MYSQL_PASSWORD`, `MYSQL_DB`.
