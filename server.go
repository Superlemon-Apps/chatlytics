package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sankalpjonn/ecount"
)

const (
	NETWORK_ADDR = "0.0.0.0:8080"
)

func handler(ec ecount.Ecount) gin.HandlerFunc {
	fn := func(ginContext *gin.Context) {
		t := time.Now()

		if ginContext.Query("shop_id") == "" || ginContext.Query("url_path") == "" {
			ginContext.JSON(http.StatusBadRequest, "mandatory elements not present")
			return
		}

		key := fmt.Sprintf(
			"%s|%s|%s|%s|%s",
			ginContext.Query("shop_id"),
			t.Format("200601"),
			t.Format("20060102"),
			t.Format("15"),
			ginContext.Query("url_path"))

		ec.Incr(key)
		ginContext.JSON(http.StatusNoContent, nil)
	}
	return gin.HandlerFunc(fn)
}

func getServer(ec ecount.Ecount) *http.Server {
	r := gin.New()
	r.Use(gin.Logger())
	r.GET("/chatlytics/chat", handler(ec))

	// start server
	return &http.Server{
		Addr:    NETWORK_ADDR,
		Handler: r,
	}
}

func stopServer(srv *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	} else {
		log.Println("gracefully shut down server")
	}
}
