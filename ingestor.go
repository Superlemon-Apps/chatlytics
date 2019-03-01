package main

import (
  "log"
  "strings"
  "strconv"
)

type Ingestor interface{
  In() chan evicted
  Start()
  Stop()
}

type ingestor struct {
  db DBHandler
  done chan bool
  in   chan evicted
}

func newIngestor(db DBHandler) Ingestor {
  return &ingestor{
    db: db,
    done: make(chan bool),
    in: make(chan evicted, 1000),
  }
}

func(self *ingestor) In() chan evicted {
  return self.in
}

func(self *ingestor) Start() {
  for ev := range self.in {
    keys := strings.Split(ev.key, "|")
		shopId := keys[0]
		day, _ := strconv.Atoi(keys[1])
		hour, _ := strconv.Atoi(keys[2])
		url_path := keys[3]
		count := ev.val
		err := self.db.Exec(
			shopId,
			day,
			hour,
			url_path,
			count)

		if err != nil {
			panic(err)
		}
  }

  log.Println("ingestion complete")
  self.done <- true
}

func(self *ingestor) Stop() {
  close(self.in)
  <-self.done
}
