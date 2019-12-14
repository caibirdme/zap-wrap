# zap-wrap
zap-wrap intergrates zap and log rotate, make zap easier to use

zap is excellent, but its configuration is too complicated. Though zap provides
a default configuration, but it's not enough for most projects.

### use case
If you:
* want to rotate the log file by time
* want clear old log files routinely
* want to write normal log to access.log and error info to error.log, just like nginx

this is for you!

### quick start
```json
[
  {
    "log_dir": "/home/deen/test",
    "filename": "access.log",
    "suffix": "%Y%m%d%H",
    "rotate_duration": "1h",
    "retain_age": "168h",
    "soft_link": true,
    "level": "debug"
  },
  {
    "log_dir": "/home/deen/test",
    "filename": "error.log",
    "suffix": "%Y%m%d%H",
    "rotate_duration": "1h",
    "retain_age": "168h",
    "soft_link": true,
    "level": "warn"
  }
]
```
config above means:
* rotate once an hour
* old log file will be named as access.log.2019121420
    * %Y: year
    * %m: Month
    * %d: day
    * %H: hour
    * %M: minute
* clear old log files which exist over 168 hours
* create symbol link named access.log to access.log.{current}, so that you can find the current log right away
* level >= warn -> error.log
* level >= debug && level < warn -> access.log

then
```go
func main() {
	var cfgs []zap_wrap.FileConfig
	_ = json.Unmarshal(jsonConfig, &cfgs)

        // So easy
	logger, _ := zap_wrap.NewLogger(cfgs...)

	logger.Debug("123", zap.Int("rand", rand.Intn(30)), zap.String("foo", `{"key":"value"}`))
        // In access.log: {"level":"debug","time":"2019-12-14T19:52:56+08:00","caller":"triple/main.go:28","msg":"123","rand":20}

	logger.Error("ttt", zap.Int("rand", rand.Intn(30)), zap.String("foo", `{"key":"value"}`))
        // In error.log: {"level":"warn","time":"2019-12-14T19:52:56+08:00","caller":"triple/main.go:28","msg":"ttt","rand":5}

}
```