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

func beforeEvictHook(db *sql.DB) func(map[string]int) {
	return func(eventCntMap map[string]int) {
		query := "INSERT INTO chat_click_event(shop_id, hour, count) values(?, ?, ?) ON DUPLICATE KEY UPDATE count = count + values(count)"
		for k, v := range eventCntMap {
			log.Println("got query: ", query)
			insert, err := db.Prepare(query)
			if err != nil {
				panic(err)
			}

			_, err = insert.Exec(strings.Split(k, "|")[0], strings.Split(k, "|")[1], v)
			if err != nil {
				panic(err)
			}
			insert.Close()
		}
	}
}

func handler(ec ecount.Ecount) gin.HandlerFunc {
	fn := func(ginContext *gin.Context) {
		key := fmt.Sprintf("%s|%s", ginContext.Param("shop"), time.Now().Format("2006010215"))
		ec.Incr(key)
		ginContext.JSON(http.StatusNoContent, nil)
	}
	return gin.HandlerFunc(fn)
}

func main() {
	db, err := sql.Open("mysql", "root:rootpluxpass@tcp(13.233.85.24)/tadpole")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ec := ecount.New(time.Second*1, beforeEvictHook(db))
	defer ec.Stop()

	r := gin.New()
	r.Use(gin.Logger())
	r.GET("/chatlytics/:shop", handler(ec))

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
