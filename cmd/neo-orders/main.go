package main

import (
	"flag"

	"github.com/dynamicgo/aliyunlog"
	"github.com/dynamicgo/config"
	"github.com/dynamicgo/slf4go"
	orderservice "github.com/inwecrypto/neo-order-service"
	_ "github.com/lib/pq"
)

var logger = slf4go.Get("neo-order-service")
var configpath = flag.String("conf", "./neo-order-service.json", "neo order service config file")

func main() {
	flag.Parse()

	neocnf, err := config.NewFromFile(*configpath)

	if err != nil {
		logger.ErrorF("load neo config err , %s", err)
		return
	}

	factory, err := aliyunlog.NewAliyunBackend(neocnf)

	if err != nil {
		logger.ErrorF("create aliyun log backend err , %s", err)
		return
	}

	slf4go.Backend(factory)

	watcher, err := orderservice.NewTxWatcher(neocnf)

	if err != nil {
		logger.ErrorF("create tx watcher err , %s", err)
		return
	}

	go watcher.Run()

	server, err := orderservice.NewHTTPServer(neocnf)

	if err != nil {
		logger.ErrorF("create http server err , %s", err)
		return
	}

	server.Run()
}
