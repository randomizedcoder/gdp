
The documnetation is all currently in a google doc

https://docs.google.com/document/d/18vrd92pDs2Mj2FwAcCcPdDhAdCwN0lXM1WaoPkyntRY/edit?tab=t.0

My hope is that you can do
```
git clone https://github.com/randomizedcoder/gdp
cd gdp
make
```

It should build all the containers and start working


To reset the Redpanda schema registry and topic, and redeploy gdp, do:
```
make restart_gdp
```

To restart just clickhouse do:
```
make build_clickhouse_and_deploy
```

Browse: http://localhost:8085/topics/topic?p=-1&s=50&o=-2#messages"

Running the following will give you helpful commands to run
```
make help
```

To stop everything:
```
make down
```

Commands to try next:
------
docker logs --follow gdp-gdp-1
docker logs --follow gdp-clickhouse-1
docker exec -ti gdp-clickhouse-1 clickhouse-client
-------
Assuming the gdp distroless:debug is available
docker exec -ti gdp-gdp-1 sh
docker exec -ti gdp-clickhouse-1 bash
clickhouse: sudo apt install iputils-ping
------
docker exec -ti gdp-clickhouse-1 tail -n 30 -f /var/log/clickhouse-server/clickhouse-server.err.log
docker exec -ti gdp-clickhouse-1 tail -n 30 -f /var/log/clickhouse-server/clickhouse-server.log
docker exec -ti gdp-clickhouse-1 clickhouse-client
docker exec -ti gdp-clickhouse-1 clickhouse-client --query "SELECT count(*) FROM gdp.ProtobufSingle;"
docker exec -ti gdp-clickhouse-1 clickhouse-client --query "SELECT * FROM system.kafka_consumers FORMAT Vertical;"
docker exec -ti gdp-clickhouse-1 clickhouse-client --query "SELECT * FROM system.stack_trace ORDER BY 'thread_id' DESC LIMIT 10;"