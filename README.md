# retrostats
OGame Retro Server statistics.

## Building server
```
cd server
# Get all dependencies
go get -d .
# Get the rice tool
go get github.com/GeertJohan/go.rice/rice
# Embed resources by generating Go source code
rice embed-go
# Build server
go build -o server *.go
```

## Restoring a database dump
Database dump is available [here](https://github.com/patrykstefanski/retrostats-dump).
```
# Create a PostgreSQL database
createdb retrostats
# Restore schema
psql -d retrostats -f schema.sql
# Decompress dump
xz -d 201704301605.sql.xz
# Restore data
psql -d retrostats -f 201704301605.sql
# You need to refresh all materialized views
# e.g. refresh materilized view top_points.
```

## License
This project is licensed under a Simplified BSD license. Please read the [LICENSE file](LICENSE).
