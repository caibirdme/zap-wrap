package zap_wrap

import (
	"encoding/base64"
	"encoding/json"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"math"
	"time"
)

type plainEncoder struct {
	buf *buffer.Buffer
	zapcore.EncoderConfig
}

func NewPlainEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	enc := &plainEncoder{buf:bufferPool.Get(), EncoderConfig: cfg}
	enc.buf.Reset()
	return enc
}

var bufferPool = buffer.NewPool()

func (enc *plainEncoder) addKey(key string) {
	enc.buf.AppendString(key)
}

var connector = []byte("||")

func (enc *plainEncoder) addConnector() {
	enc.buf.Write(connector)
}

func (enc *plainEncoder) addElementSeparator() {
	enc.buf.AppendByte('=')
}

func (enc *plainEncoder) AddArray(key string, arr zapcore.ArrayMarshaler) error {
	enc.addKey(key)
	enc.addElementSeparator()
	return enc.AppendArray(arr)
}

func (enc *plainEncoder) AddObject(key string, obj zapcore.ObjectMarshaler) error {
	enc.addKey(key)
	enc.addElementSeparator()
	return enc.AppendObject(obj)
}

func (enc *plainEncoder) AddBinary(key string, val []byte) {
	enc.AddString(key, base64.StdEncoding.EncodeToString(val))
}

func (enc *plainEncoder) AddByteString(key string, val []byte) {
	enc.addKey(key)
	enc.addElementSeparator()
	enc.AppendByteString(val)
}

func (enc *plainEncoder) AddBool(key string, val bool) {
	enc.addKey(key)
	enc.addElementSeparator()
	enc.AppendBool(val)
}

func (enc *plainEncoder) AddComplex128(key string, val complex128) {
	enc.addKey(key)
	enc.addElementSeparator()
	enc.AppendComplex128(val)
}

func (enc *plainEncoder) AddDuration(key string, val time.Duration) {
	enc.addKey(key)
	enc.addElementSeparator()
	enc.AppendDuration(val)
}

func (enc *plainEncoder) AddFloat64(key string, val float64) {
	enc.addKey(key)
	enc.addElementSeparator()
	enc.AppendFloat64(val)
}

func (enc *plainEncoder) AddInt64(key string, val int64) {
	enc.addKey(key)
	enc.addElementSeparator()
	enc.AppendInt64(val)
}

var nullLiteralBytes = []byte("null")

// Only invoke the standard JSON encoder if there is actually something to
// encode; otherwise write JSON null literal directly.
func (enc *plainEncoder) encodeReflected(obj interface{}) ([]byte, error) {
	if obj == nil {
		return nullLiteralBytes, nil
	}
	return json.Marshal(obj)
}

func (enc *plainEncoder) AddReflected(key string, obj interface{}) error {
	valueBytes, err := enc.encodeReflected(obj)
	if err != nil {
		return err
	}
	enc.addKey(key)
	enc.addElementSeparator()
	_, err = enc.buf.Write(valueBytes)
	return err
}

func (enc *plainEncoder) OpenNamespace(key string) {
	// todo:
	return
}

func (enc *plainEncoder) AddString(key, val string) {
	enc.addKey(key)
	enc.addElementSeparator()
	enc.AppendString(val)
}

func (enc *plainEncoder) AddTime(key string, val time.Time) {
	enc.addKey(key)
	enc.addElementSeparator()
	enc.AppendTime(val)
}

func (enc *plainEncoder) AddUint64(key string, val uint64) {
	enc.addKey(key)
	enc.addElementSeparator()
	enc.AppendUint64(val)
}

func (enc *plainEncoder) AppendArray(arr zapcore.ArrayMarshaler) error {
	err := arr.MarshalLogArray(enc)
	return err
}

func (enc *plainEncoder) AppendObject(obj zapcore.ObjectMarshaler) error {
	err := obj.MarshalLogObject(enc)
	return err
}

func (enc *plainEncoder) AppendBool(val bool) {
	enc.buf.AppendBool(val)
}

func (enc *plainEncoder) AppendByteString(val []byte) {
	enc.buf.Write(val)
}

func (enc *plainEncoder) AppendComplex128(val complex128) {
	// Cast to a platform-independent, fixed-size type.
	r, i := float64(real(val)), float64(imag(val))
	enc.buf.AppendByte('"')
	// Because we're always in a quoted string, we can use strconv without
	// special-casing NaN and +/-Inf.
	enc.buf.AppendFloat(r, 64)
	enc.buf.AppendByte('+')
	enc.buf.AppendFloat(i, 64)
	enc.buf.AppendByte('i')
	enc.buf.AppendByte('"')
}

func (enc *plainEncoder) AppendDuration(val time.Duration) {
	cur := enc.buf.Len()
	enc.EncodeDuration(val, enc)
	if cur == enc.buf.Len() {
		// User-supplied EncodeDuration is a no-op. Fall back to nanoseconds to keep
		// JSON valid.
		enc.AppendInt64(int64(val))
	}
}

func (enc *plainEncoder) AppendInt64(val int64) {
	enc.buf.AppendInt(val)
}

func (enc *plainEncoder) AppendReflected(val interface{}) error {
	valueBytes, err := enc.encodeReflected(val)
	if err != nil {
		return err
	}
	_, err = enc.buf.Write(valueBytes)
	return err
}

func (enc *plainEncoder) AppendString(val string) {
	enc.buf.AppendString(val)
}

func (enc *plainEncoder) AppendTime(val time.Time) {
	cur := enc.buf.Len()
	enc.EncodeTime(val, enc)
	if cur == enc.buf.Len() {
		// User-supplied EncodeTime is a no-op. Fall back to nanos since epoch to keep
		// output JSON valid.
		enc.AppendInt64(val.UnixNano()/1e6)
	}
}

func (enc *plainEncoder) AppendUint64(val uint64) {
	enc.buf.AppendUint(val)
}

func (enc *plainEncoder) appendFloat(val float64, bitSize int) {
	switch {
	case math.IsNaN(val):
		enc.buf.AppendString(`"NaN"`)
	case math.IsInf(val, 1):
		enc.buf.AppendString(`"+Inf"`)
	case math.IsInf(val, -1):
		enc.buf.AppendString(`"-Inf"`)
	default:
		enc.buf.AppendFloat(val, bitSize)
	}
}

func (enc *plainEncoder) AddComplex64(k string, v complex64) { enc.AddComplex128(k, complex128(v)) }
func (enc *plainEncoder) AddFloat32(k string, v float32)     { enc.AddFloat64(k, float64(v)) }
func (enc *plainEncoder) AddInt(k string, v int)             { enc.AddInt64(k, int64(v)) }
func (enc *plainEncoder) AddInt32(k string, v int32)         { enc.AddInt64(k, int64(v)) }
func (enc *plainEncoder) AddInt16(k string, v int16)         { enc.AddInt64(k, int64(v)) }
func (enc *plainEncoder) AddInt8(k string, v int8)           { enc.AddInt64(k, int64(v)) }
func (enc *plainEncoder) AddUint(k string, v uint)           { enc.AddUint64(k, uint64(v)) }
func (enc *plainEncoder) AddUint32(k string, v uint32)       { enc.AddUint64(k, uint64(v)) }
func (enc *plainEncoder) AddUint16(k string, v uint16)       { enc.AddUint64(k, uint64(v)) }
func (enc *plainEncoder) AddUint8(k string, v uint8)         { enc.AddUint64(k, uint64(v)) }
func (enc *plainEncoder) AddUintptr(k string, v uintptr)     { enc.AddUint64(k, uint64(v)) }
func (enc *plainEncoder) AppendComplex64(v complex64)        { enc.AppendComplex128(complex128(v)) }
func (enc *plainEncoder) AppendFloat64(v float64)            { enc.appendFloat(v, 64) }
func (enc *plainEncoder) AppendFloat32(v float32)            { enc.appendFloat(float64(v), 32) }
func (enc *plainEncoder) AppendInt(v int)                    { enc.AppendInt64(int64(v)) }
func (enc *plainEncoder) AppendInt32(v int32)                { enc.AppendInt64(int64(v)) }
func (enc *plainEncoder) AppendInt16(v int16)                { enc.AppendInt64(int64(v)) }
func (enc *plainEncoder) AppendInt8(v int8)                  { enc.AppendInt64(int64(v)) }
func (enc *plainEncoder) AppendUint(v uint)                  { enc.AppendUint64(uint64(v)) }
func (enc *plainEncoder) AppendUint32(v uint32)              { enc.AppendUint64(uint64(v)) }
func (enc *plainEncoder) AppendUint16(v uint16)              { enc.AppendUint64(uint64(v)) }
func (enc *plainEncoder) AppendUint8(v uint8)                { enc.AppendUint64(uint64(v)) }
func (enc *plainEncoder) AppendUintptr(v uintptr)            { enc.AppendUint64(uint64(v)) }

func (enc *plainEncoder) Clone() zapcore.Encoder {
	cloned := &plainEncoder{
		buf:           bufferPool.Get(),
		EncoderConfig: enc.EncoderConfig,
	}
	cloned.buf.Write(enc.buf.Bytes())
	return cloned
}

func (enc *plainEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	if enc.LevelKey != "" && enc.EncodeLevel != nil {
		enc.buf.AppendByte('[')
		enc.buf.AppendString(ent.Level.String())
		enc.buf.AppendByte(']')
	}

	enc.buf.AppendByte(' ')

	enc.addFields(fields)
	buf := enc.buf
	enc.buf = nil
	return buf, nil
}

func (enc *plainEncoder) addFields(fields []zapcore.Field) {
	n := len(fields)
	fields[0].AddTo(enc)
	for i := 1; i<n; i++ {
		enc.addConnector()
		fields[i].AddTo(enc)
	}
}





