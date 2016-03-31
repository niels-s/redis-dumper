# Redis Dumper

This script dumps all the entries from one Redis DB into a file in the redis protocol format.
See here (http://redis.io/topics/protocol) and here(http://redis.io/topics/mass-insert).
This allows use to pipe the resulting file directly into redis with pipe command like this

    > redis-dumper -address=127.0.0.1:6379 -db=0
    > cat redis_db_0_dump.rdb | redis-cli --pipe

This script is especially created to get contents from AWS Elasticache but works with all Redis instances
