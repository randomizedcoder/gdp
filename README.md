
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

Running the following will give you helpful commands to run
```
make help
```