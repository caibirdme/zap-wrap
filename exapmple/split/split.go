package main

import (
	"encoding/json"
	"github.com/caibirdme/zap-wrap"
	"go.uber.org/zap"
	"log"
	"math/rand"
	"time"
)

func main() {
	var cfgs []zap_wrap.FileConfig
	err := json.Unmarshal([]byte(jsonCfg), &cfgs)
	if err != nil {
		log.Fatal(err)
	}
	logger, err := zap_wrap.NewLogger(true, cfgs...)
	if err != nil {
		log.Fatal(err)
	}
	tick := time.NewTicker(time.Second)
	count := 0
	for count < 500 {
		<-tick.C
		logger.Debug("", zap.Int("rand", rand.Intn(30)), zap.String("foo", `{"key":"value"}`))
		logger.Error("", zap.Int("rand", rand.Intn(30)), zap.String("foo", `{"key":"value"}`))
		count += 1
	}
}

const jsonCfg = `
[
  {
    "log_dir": "/home/deen/test",
    "filename": "access.log",
    "suffix": "%Y%m%d%H%M",
    "rotate_duration": "1m",
    "retain_age": "168h",
    "soft_link": true,
    "level": "debug"
  },
  {
    "log_dir": "/home/deen/test",
    "filename": "error.log",
    "suffix": "%Y%m%d%H%M",
    "rotate_duration": "1m",
    "retain_age": "168h",
    "soft_link": true,
    "level": "warn"
  }
]
`
