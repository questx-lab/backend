# QuestX

# Get started

Start the database
```shell
make start-db
```

Start the web server
```shell
make start-server
```

# Migration

Create migration db file:
```shell
migrate create -ext sql -dir migration/mysql -seq <migration_name>
```

```cql
docker exec -it scylladb cqlsh

CREATE KEYSPACE xquest WITH replication = {'class' : 'NetworkTopologyStrategy', 'datacenter1' : 1};
```

Revert migration
```shell
SELECT * FROM gocqlx_migrate;
DELETE FROM xquest.gocqlx_migrate WHERE name = 'x';
```
