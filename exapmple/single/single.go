package main

import (
	"encoding/json"
	zap_wrap "github.com/caibirdme/zap-wrap"
	"go.uber.org/zap"
	"log"
	"math/rand"
	"time"
)

const nosplitCfg = `
{
	"log_dir": "/home/deen/test",
	"filename": "some.log",
	"suffix": "%Y%m%d%H%M",
	"rotate_duration": "1m",
	"retain_age": "168h",
	"soft_link": true,
	"level": "debug"
}
`

func main() {
	var cfgs zap_wrap.FileConfig
	err := json.Unmarshal([]byte(nosplitCfg), &cfgs)
	if err != nil {
		log.Fatal(err)
	}
	logger, err := zap_wrap.NewLogger(true, cfgs)
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
