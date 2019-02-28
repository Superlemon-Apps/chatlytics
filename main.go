package main

import (
  "os"
  "os/signal"
  "syscall"
  "time"
  "net/http"
  "log"
  "fmt"
  "strings"
  "context"

  "github.com/gin-gonic/gin"
  "github.com/sankalpjonn/ecount"
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
)

func beforeEvictHook(db *sql.DB) func(map[string]int) {
    return func(eventCntMap map[string]int) {
      query := "INSERT INTO chat_click_event(shop_id, hour, count) values(?, ?, ?)"
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
    ec.Incr(fmt.Sprintf("%s|%s", ginContext.Param("event"), time.Now().Format("2006010215")))
		ginContext.Header("content-type", "application/json;charset=utf-8")
		ginContext.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}
	return gin.HandlerFunc(fn)
}

func main() {
  db, err := sql.Open("sqlite3", "/home/sankalpjonna/gowork/src/github.com/sankalpjonn/events/test.db")
  if err != nil {
    panic(err)
  }
  defer db.Close()

  ec := ecount.New(time.Second * 60, beforeEvictHook(db))
  defer ec.Stop()

  r := gin.New()
  r.Use(gin.Logger())
  r.GET("/events/:event", handler(ec))

  srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
  defer func(){
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
