// this code makes the following assumptions:
// redis was deployed inside kubernetes using the bitnami-redis helm chart
// redis is running within a namespace called "redis"
// the name of the redis sentinal service is "primary-redis"
// the sentinel master is named "mymaster"
// this code is running inside that same kubernetes cluster

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

func defaultPath(w http.ResponseWriter, r *http.Request) {
	var ctx = context.Background()
	rdb := redis.NewFailoverClusterClient(&redis.FailoverOptions{
		MasterName:    "mymaster",
		SentinelAddrs: []string{"primary-redis.redis.svc.cluster.local:26379"},
		RouteRandomly: true,
		Password:      "",
		DB:            0,
	})

	status := rdb.Info(ctx, "all")
	fmt.Fprintf(w, "%v", status)
}

func getAllNodes(w http.ResponseWriter, r *http.Request) {
	var ctx = context.Background()
	rdb := redis.NewFailoverClusterClient(&redis.FailoverOptions{
		MasterName:    "mymaster",
		SentinelAddrs: []string{"primary-redis.redis.svc.cluster.local:26379"},
		RouteRandomly: true,
		Password:      "",
		DB:            0,
	})

	val, err := rdb.Keys(ctx, "*").Result()
	if err != nil {
		fmt.Fprintf(w, "error retrieving nodes")
	} else {
		jsonVal, _ := json.MarshalIndent(val, "", " ")
		fmt.Fprintf(w, "{ \"nodes\": %v }", string(jsonVal))
	}
}

func getNode(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["id"]

	var ctx = context.Background()
	rdb := redis.NewFailoverClusterClient(&redis.FailoverOptions{
		MasterName:    "mymaster",
		SentinelAddrs: []string{"primary-redis.redis.svc.cluster.local:26379"},
		RouteRandomly: true,
		Password:      "",
		DB:            0,
	})

	exists, err := rdb.Exists(ctx, key).Result()
	if err != nil {
		fmt.Fprintf(w, "%v", err)
	} else if exists == 0 {
		fmt.Fprintf(w, "node does not exist")
	} else if exists == 1 {
		val, err := rdb.Get(ctx, key).Result()
		if err != nil {
			fmt.Fprintf(w, "error reading node data")
		} else {
			fmt.Fprintf(w, "%v", val)
		}
	}
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", defaultPath)
	router.HandleFunc("/nodes", getAllNodes).Methods("GET")
	router.HandleFunc("/nodes/{id}", getNode).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", router))
}
