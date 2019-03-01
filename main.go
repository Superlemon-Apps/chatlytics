package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"flag"

	"github.com/sankalpjonn/ecount"
)

var connStr string

const (
	CHAT_CLICKS_INSERT_QUERY = "INSERT INTO chat_click_event(shop_id, day, hour, url_path, count) values(?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE count = count + values(count)"
)

func main() {

	flag.StringVar(&connStr, "conn", "", "mysql connection string")
	flag.Parse()

	// start database connection
	db := newDb(CHAT_CLICKS_INSERT_QUERY, connStr)

	// start data ingestor
	ingestor := newIngestor(db)
	go ingestor.Start()

	// start counter
	ec := ecount.New(
		time.Second*60,
		func(eventCntMap map[string]int) {
			for k, v := range eventCntMap {
				ingestor.In() <- evicted{key: k, val: v}
			}
		},
	)

	// start http server
	srv := getServer(ec)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	// to be run on kill signal
	defer func(){
		// stop accountsepting new connections
		stopServer(srv)

		// flush current counts
		ec.Stop()

		// finish ingesting current data
		ingestor.Stop()

		// close database connection
		db.Close()
	}()


	// wait for kill signal
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	<-sigc
}
