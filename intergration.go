package zap_wrap

import (
	"bytes"
	"encoding/json"
	"errors"
	rotate "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type FileConfig struct {
	// LogDir absolute dir path
	LogDir string `json:"log_dir"`
	// FileName such as access.log etc.
	FileName string `json:"filename"`
	// Suffix if set "%Y%m%d%H%M"(year month day hour minute), the rotated file
	// will be named as "filename.201912141922".
	Suffix string `json:"suffix"`
	// RotatePeriod time period to rotate the file
	RotatePeriod Duration `json:"rotate_duration,omitempty"`
	// RetainMaxAge, this will purge old files those whom exceeds this duration
	RetainMaxAge Duration `json:"retain_age,omitempty"`
	// SoftLink if set softlink to current log file, so you can always tail the same file
	SoftLink bool `json:"soft_link"`
	// Level, set log level
	Level LogLevel `json:"level"`
	// EncodeCfg, ignore this if you don't know about it. This lib provides a default config that may meet most
	// users' requirement
	EncodeCfg *zapcore.EncoderConfig
}

type Duration time.Duration

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))
		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(tmp)
		return nil
	default:
		return errors.New("invalid duration")
	}
}

type LogLevel struct {
	zapcore.Level
}

func (l *LogLevel) UnmarshalJSON(data []byte) error {
	var level zapcore.Level
	data = bytes.Trim(data, `"`)
	err := level.UnmarshalText(data)
	if err != nil {
		return err
	}
	l.Level = level
	return nil
}

type sortConfig []FileConfig

func (s sortConfig) Len() int {
	return len(s)
}

func (s sortConfig) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortConfig) Less(i, j int) bool {
	return s[i].Level.Level < s[j].Level.Level
}

// NewLogger combines multiple writers and creates an uniform zap.Logger
func NewLogger(addCaller bool, cfgs ...FileConfig) (*zap.Logger, error) {
	sort.Sort(sortConfig(cfgs))
	n := len(cfgs)
	levelEnablers := make([]zapcore.LevelEnabler, 0, n)
	for i := 0; i < n-1; i++ {
		t := i
		levelEnablers = append(levelEnablers, zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return cfgs[t].Level.Enabled(lvl) && !cfgs[t+1].Level.Enabled(lvl)
		}))
	}
	levelEnablers = append(levelEnablers, cfgs[n-1].Level.Level)
	writers := make([]io.WriteCloser, 0, n)
	for _, cfg := range cfgs {
		w, err := newRotateWriter(cfg)
		if err != nil {
			return nil, err
		}
		writers = append(writers, w)
	}
	cores := make([]zapcore.Core, 0, n)
	defaultCfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: milliSecondsEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	for idx, cfg := range cfgs {
		var enc zapcore.Encoder
		if cfg.EncodeCfg == nil {
			enc = zapcore.NewJSONEncoder(defaultCfg)
		} else {
			enc = zapcore.NewJSONEncoder(*cfg.EncodeCfg)
		}
		cores = append(cores, zapcore.NewCore(enc, zapcore.AddSync(writers[idx]), levelEnablers[idx]))
	}
	if addCaller {
		return zap.New(zapcore.NewTee(cores...), zap.AddCaller()), nil
	}
	return zap.New(zapcore.NewTee(cores...)), nil
}

func milliSecondsEncoder(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendInt64(d.Milliseconds())
}

func newRotateWriter(cfg FileConfig) (io.WriteCloser, error) {
	var options []rotate.Option
	absLog, err := filepath.Abs(cfg.LogDir)
	if err != nil {
		return nil, err
	}
	if _, err = os.Stat(absLog); err != nil && os.IsNotExist(err) {
		err = os.Mkdir(absLog, 0755)
		if err != nil {
			return nil, err
		}
	}
	absFileName, err := filepath.Abs(filepath.Join(cfg.LogDir, cfg.FileName))
	if err != nil {
		return nil, err
	}
	if cfg.SoftLink {
		options = append(options, rotate.WithLinkName(absFileName))
	}
	if cfg.RotatePeriod != 0 {
		options = append(options, rotate.WithRotationTime(time.Duration(cfg.RotatePeriod)))
	}
	if cfg.RetainMaxAge != 0 {
		options = append(options, rotate.WithMaxAge(time.Duration(cfg.RetainMaxAge)))
	} else {
		options = append(options, rotate.WithMaxAge(7*24*time.Hour))
	}

	if cfg.Suffix != "" {
		absFileName += "." + cfg.Suffix
	}
	logger, err := rotate.New(
		absFileName,
		options...,
	)
	if err != nil {
		return nil, err
	}
	return logger, nil
}
