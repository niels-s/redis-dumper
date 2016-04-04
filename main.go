package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

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

const restoreCommand = "*4\r\n$7\r\nRESTORE\r\n"

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
			processKey(client, writer, key)
		}
		writer.Flush()

		if cursor == 0 {
			break
		}
	}

	log.Println("End processing")
}

func processKey(client *redis.Client, writer *bufio.Writer, key string) {
	dump, err := client.Dump(key).Result()
	if err != nil {
		log.Printf("ERROR: couldn't dump key %s: %v", key, err)
		return
	}

	ttl, err := client.TTL(key).Result()
	if err != nil {
		log.Printf("ERROR: couldn't dump key %s: %v", key, err)
		return
	}

	writer.WriteString(createRestoreCommand(key, dump, &ttl))
}

func createRestoreCommand(key, dump string, ttl *time.Duration) string {
	seconds := int(ttl.Seconds() * 1000)
	if seconds < 0 {
		seconds = 0
	}
	ttlString := strconv.Itoa(seconds)

	result := restoreCommand

	for _, val := range [3]string{key, ttlString, dump} {
		result += "$" + strconv.Itoa(len(val)) + "\r\n" + val + "\r\n"
	}

	return result
}

func createFile() (*os.File, *bufio.Writer) {
	file, err := os.Create(fmt.Sprintf("redis_db_%d_dump.rdb", redisDB))
	if err != nil {
		log.Fatalf("Couldn't create file: %v", err)
	}

	return file, bufio.NewWriter(file)
}
