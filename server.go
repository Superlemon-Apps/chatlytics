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

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func getServer(ec ecount.Ecount) *http.Server {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(CORSMiddleware())
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
