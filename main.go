package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sankalpjonn/ecount"
)

type evicted struct {
	key string
	val int
}

func beforeEvictHook(evictedC chan evicted) func(map[string]int) {
	return func(eventCntMap map[string]int) {
		for k, v := range eventCntMap {
			evictedC <- evicted{key: k, val: v}
		}
	}
}

func handler(ec ecount.Ecount) gin.HandlerFunc {
	fn := func(ginContext *gin.Context) {
		t := time.Now()

		if ginContext.Query("shop_id") == "" || ginContext.Query("url") == "" {
			ginContext.JSON(http.StatusBadRequest, "mandatory elements not present")
			return
		}

		key := fmt.Sprintf(
			"%s|%s|%s|%s",
			ginContext.Query("shop_id"),
			t.Format("20060102"),
			t.Format("15"),
			ginContext.Query("url"))

		ec.Incr(key)
		ginContext.JSON(http.StatusNoContent, nil)
	}
	return gin.HandlerFunc(fn)
}

func ingestToSql(evictedC chan evicted, db *sql.DB, evictionComplete chan bool) {
	query := "INSERT INTO chat_click_event(shop_id, day, hour, url, count) values(?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE count = count + values(count)"

	for ev := range evictedC {
		insert, err := db.Prepare(query)
		if err != nil {
			panic(err)
		}

		_, err = insert.Exec(strings.Split(ev.key, "|")[0], strings.Split(ev.key, "|")[1], strings.Split(ev.key, "|")[2], strings.Split(ev.key, "|")[3], ev.val)
		if err != nil {
			panic(err)
		}
		insert.Close()

		log.Println("ran eviction ", query)
		time.Sleep(time.Second * 3)
	}

	evictionComplete <- true
}

func main() {
	db, err := sql.Open("mysql", "root:rootpluxpass@tcp(13.233.85.24)/tadpole")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	evictedC := make(chan evicted, 1000)
	evictionComplete := make(chan bool)
	go ingestToSql(evictedC, db, evictionComplete)
	defer func() {
		close(evictedC)
		<-evictionComplete
	}()

	ec := ecount.New(time.Second*60, beforeEvictHook(evictedC))
	defer ec.Stop()

	r := gin.New()
	r.Use(gin.Logger())
	r.GET("/chatlytics", handler(ec))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("Server Shutdown:", err)
		} else {
			log.Println("gracefully shut down server")
		}
	}()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	<-sigc
}
