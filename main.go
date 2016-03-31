package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"gopkg.in/redis.v3"
)

var (
	version       string
	redisDB       int64
	redisAddr     string
	documentation = `Redis Dumper

This script dumps all the entries from one Redis DB into a file in the redis protocol format.
See here (http://redis.io/topics/protocol) and here (http://redis.io/topics/mass-insert).
This allows use to pipe the resulting file directly into redis with pipe command like this

> cat redis_db_0_dump.rdb | redis-cli --pipe

This script is especially created to get contents from AWS Elasticache but works with all Redis instances

`
)

func init() {
	flag.Int64Var(&redisDB, "db", 0, "Indicate which db to process")
	flag.StringVar(&redisAddr, "address", "localhost:6379", "Redis address (url and port)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, documentation)
		fmt.Fprintf(os.Stderr, "Usage of Redis Dumper:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nCurrent Version: %s\n", version)
	}
}

func main() {
	flag.Parse()
	log.Println("Start processing")

	client := redis.NewClient(&redis.Options{
		DB:   redisDB,
		Addr: redisAddr,
	})

	file, writer := createFile()
	defer file.Close()

	var cursor int64
	for {
		var keys []string
		var err error
		cursor, keys, err = client.Scan(cursor, "", 1000).Result()
		if err != nil {
			log.Fatalf("Couldn't iterate through set: %v", err)
		}

		for _, key := range keys {
			dump, err := client.Dump(key).Result()
			if err != nil {
				log.Printf("ERROR: couldn't dump key %s: %v", key, err)
				return
			}
			writer.WriteString(createRestoreCommand(key, dump))
		}
		writer.Flush()

		if cursor == 0 {
			break
		}
	}

	log.Println("End processing")
}

func createRestoreCommand(key, dump string) string {
	proto := "*4\r\n$7\r\nRESTORE\r\n"
	key_proto := "$" + strconv.Itoa(len(key)) + "\r\n" + key + "\r\n"
	ttl_proto := "$1\r\n0\r\n"
	dump_proto := "$" + strconv.Itoa(len(dump)) + "\r\n" + dump + "\r\n"

	return proto + key_proto + ttl_proto + dump_proto
}

func createFile() (*os.File, *bufio.Writer) {
	file, err := os.Create(fmt.Sprintf("redis_db_%d_dump.rdb", redisDB))
	if err != nil {
		log.Fatalf("Couldn't create file: %v", err)
	}

	return file, bufio.NewWriter(file)
}
