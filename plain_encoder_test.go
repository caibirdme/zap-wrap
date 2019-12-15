package zap_wrap

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"testing"
)

func TestLog(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	core := zapcore.NewCore(NewPlainEncoder(zap.NewDevelopmentEncoderConfig()), zapcore.AddSync(buf), zapcore.DebugLevel)
	logger := zap.New(core)
	intArr := zap.Ints("ints", []int{1,2,3})
	logger.Warn("hello", intArr)
	should := assert.New(t)
	should.Equal("[warn] ints=1,2,3", buf.String())
}
