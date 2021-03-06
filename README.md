# Redis Dumper

This script dumps all the entries from one Redis DB into a file in the redis protocol
format. See [here](http://redis.io/topics/protocol) and [here](http://redis.io/topics/mass-insert).
This allows use to pipe the resulting file directly into redis with pipe command
like this.

```bash
# dump redis database
$ redis-dumper -h 127.0.0.1 -p 6379 -n 0

# restore redis database from file
$ cat redis_db_0_dump.rdb | redis-cli --pipe
```

This script is especially created to get contents from AWS Elasticache but works
with all Redis instances.
