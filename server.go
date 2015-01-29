package main

import(
	"fmt"
	"net/http"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"gopkg.in/redis.v2"
)

type ServerConfig struct {
	Port string
	Redis string
	Servers map[string][]string
}

var(
	server_config = ServerConfig{}
	redis_client *redis.Client
)

func LoadConfig() {
	buf, err := ioutil.ReadFile("config.yml")
	if err != nil {
		fmt.Println("error: %v", err)
	}
	err = yaml.Unmarshal(buf, &server_config)
	if err != nil {
		fmt.Println("error: %v", err)
	}
}

func ConnectRedis() {
	redis_client = redis.NewTCPClient(&redis.Options{
	Addr:     server_config.Redis,
	Password: "", // no password set
	DB:       0,  // use default DB
	})
}

func SetServers() {
	for server_name, server_params := range server_config.Servers {
		redis_client.Set("servers:" + server_name + ":ip", server_params[0])
		redis_client.Set("servers:" + server_name + ":ssh_user", server_params[1])
		fmt.Println("Node " + server_name + " published in redis store")
	}
}

func main() {
	LoadConfig()
	ConnectRedis()
	SetServers()
	http.HandleFunc("/status", status)
	http.HandleFunc("/api", api_handler)
	http.HandleFunc("/ps", ps)
	http.HandleFunc("/metrics", metrics)
	fmt.Println("Server running on port", server_config.Port)
	http.ListenAndServe(":" + server_config.Port, nil)
}