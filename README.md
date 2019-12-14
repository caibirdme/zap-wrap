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
    "rotate_duration": "1m",
    "retain_age": "168h",
    "soft_link": true,
    "level": "debug"
  },
  {
    "log_dir": "/home/deen/test",
    "filename": "error.log",
    "suffix": "%Y%m%d%H",
    "rotate_duration": "1m",
    "retain_age": "168h",
    "soft_link": true,
    "level": "warn"
  }
]
```
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