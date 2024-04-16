package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/guonaihong/bench-ws/config"
	"github.com/guonaihong/bench-ws/core"
	"github.com/guonaihong/clop"
	"github.com/lesismal/nbio/logging"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

var upgrader = websocket.NewUpgrader()

type Config struct {
	Addr           string `clop:"short;long" usage:"websocket server address" default:":4444""`
	LimitPortRange int    `clop:"short;long" usage:"limit port range" default:"1"`
}

func main() {
	var cnf Config
	clop.Bind(&cnf)
	logging.SetLevel(logging.LevelError)

	upgrader.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {
		c.WriteMessage(messageType, data)
	})
	upgrader.KeepaliveTime = 0
	upgrader.BlockingModAsyncWrite = false

	addrs, err := config.GetFrameworkServerAddrs(config.NbioStd, cnf.LimitPortRange)
	if err != nil {
		log.Fatalf("GetFrameworkBenchmarkAddrs(%v) failed: %v", config.NbioStd, err)
	}
	lns := core.StartServers(addrs, onWebsocket)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	for _, ln := range lns {
		ln.Close()
	}
}

func onWebsocket(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade failed: %v", err)
		return
	}
	c.SetReadDeadline(time.Time{})
}
