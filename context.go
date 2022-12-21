package logiface

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

type (
	// Context is used to build a sub-logger, see Logger.Field.
	Context[E Event] struct {
		Modifiers ModifierSlice[E]
		methods   modifierMethods[E]
		logger    *Logger[E]
	}

	// Builder is used to build a log event, see Logger.Build, Logger.Info, etc.
	Builder[E Event] struct {
		Event   E
		methods modifierMethods[E]
		shared  *loggerShared[E]
	}

	modifierMethods[E Event] struct{}
)

func (x *Context[E]) Logger() *Logger[E] {
	if x == nil {
		return nil
	}
	return x.logger
}

func (x *Context[E]) add(fn ModifyFunc[E]) {
	x.Modifiers = append(x.Modifiers, fn)
}

func (x *Builder[E]) Call(fn func(b *Builder[E])) *Builder[E] {
	fn(x)
	return x
}

func (x *Builder[E]) Log(msg string) {
	if x == nil {
		return
	}
	defer x.release()
	if x.Event.Level().Enabled() {
		x.log(msg)
	}
}

func (x *Builder[E]) Logf(format string, args ...any) {
	if x == nil {
		return
	}
	defer x.release()
	if x.Event.Level().Enabled() {
		x.log(fmt.Sprintf(format, args...))
	}
}

func (x *Builder[E]) LogFunc(fn func() string) {
	if x == nil {
		return
	}
	defer x.release()
	if x.Event.Level().Enabled() {
		x.log(fn())
	}
}

func (x *Builder[E]) log(msg string) {
	if !x.Event.AddMessage(msg) {
		x.Event.AddField(`msg`, msg)
	}
	_ = x.shared.writer.Write(x.Event)
}

func (x *Builder[E]) release() {
	if x.shared != nil {
		x.shared.pool.Put(x)
	}
}

func (x modifierMethods[E]) Field(event E, key string, val any) error {
	if !event.Level().Enabled() {
		return ErrDisabled
	}
	switch val := val.(type) {
	case string:
		x.str(event, key, val)
	case []byte:
		x.bytes(event, key, val)
	case time.Time:
		x.timestamp(event, key, val)
	case time.Duration:
		x.duration(event, key, val)
	case int:
		x.int(event, key, val)
	case float32:
		x.float32(event, key, val)
	default:
		event.AddField(key, val)
	}
	return nil
}

// Field adds a field to the log context, making an effort to choose the most
// appropriate handler for the value.
//
// WARNING: The behavior of this method may change without notice.
//
// Use the Interface method if you want a direct pass-through to the
// Event.AddField implementation.
func (x *Context[E]) Field(key string, val any) *Context[E] {
	if x != nil && x.logger != nil {
		x.add(func(event E) error { return x.methods.Field(event, key, val) })
	}
	return x
}

// Field adds a field to the log event, making an effort to choose the most
// appropriate handler for the value.
//
// WARNING: The behavior of this method may change without notice.
//
// Use the Interface method if you want a direct pass-through to the
// Event.AddField implementation.
func (x *Builder[E]) Field(key string, val any) *Builder[E] {
	if x != nil && x.shared != nil {
		_ = x.methods.Field(x.Event, key, val)
	}
	return x
}

func (x modifierMethods[E]) Interface(event E, key string, val any) error {
	if !event.Level().Enabled() {
		return ErrDisabled
	}
	event.AddField(key, val)
	return nil
}
func (x *Context[E]) Interface(key string, val any) *Context[E] {
	if x != nil && x.logger != nil {
		x.add(func(event E) error { return x.methods.Interface(event, key, val) })
	}
	return x
}
func (x *Builder[E]) Interface(key string, val any) *Builder[E] {
	if x != nil && x.shared != nil {
		_ = x.methods.Interface(x.Event, key, val)
	}
	return x
}

func (x modifierMethods[E]) Err(event E, err error) error {
	if !event.Level().Enabled() {
		return ErrDisabled
	}
	if !event.AddError(err) {
		event.AddField(`err`, err)
	}
	return nil
}
func (x *Context[E]) Err(err error) *Context[E] {
	if x != nil && x.logger != nil {
		x.add(func(event E) error { return x.methods.Err(event, err) })
	}
	return x
}
func (x *Builder[E]) Err(err error) *Builder[E] {
	if x != nil && x.shared != nil {
		_ = x.methods.Err(x.Event, err)
	}
	return x
}

func (x modifierMethods[E]) Str(event E, key string, val string) error {
	if !event.Level().Enabled() {
		return ErrDisabled
	}
	x.str(event, key, val)
	return nil
}
func (x *Context[E]) Str(key string, val string) *Context[E] {
	if x != nil && x.logger != nil {
		x.add(func(event E) error { return x.methods.Str(event, key, val) })
	}
	return x
}
func (x *Builder[E]) Str(key string, val string) *Builder[E] {
	if x != nil && x.shared != nil {
		_ = x.methods.Str(x.Event, key, val)
	}
	return x
}

func (x modifierMethods[E]) Int(event E, key string, val int) error {
	if !event.Level().Enabled() {
		return ErrDisabled
	}
	x.int(event, key, val)
	return nil
}
func (x *Context[E]) Int(key string, val int) *Context[E] {
	if x != nil && x.logger != nil {
		x.add(func(event E) error { return x.methods.Int(event, key, val) })
	}
	return x
}
func (x *Builder[E]) Int(key string, val int) *Builder[E] {
	if x != nil && x.shared != nil {
		_ = x.methods.Int(x.Event, key, val)
	}
	return x
}

func (x modifierMethods[E]) Float32(event E, key string, val float32) error {
	if !event.Level().Enabled() {
		return ErrDisabled
	}
	x.float32(event, key, val)
	return nil
}
func (x *Context[E]) Float32(key string, val float32) *Context[E] {
	if x != nil && x.logger != nil {
		x.add(func(event E) error { return x.methods.Float32(event, key, val) })
	}
	return x
}
func (x *Builder[E]) Float32(key string, val float32) *Builder[E] {
	if x != nil && x.shared != nil {
		_ = x.methods.Float32(x.Event, key, val)
	}
	return x
}

func (x modifierMethods[E]) str(event E, key string, val string) {
	if !event.AddString(key, val) {
		event.AddField(key, val)
	}
}

func (x modifierMethods[E]) bytes(event E, key string, val []byte) {
	// TODO allow custom handling via an optional method
	x.str(event, key, base64.StdEncoding.EncodeToString(val))
}

func (x modifierMethods[E]) timestamp(event E, key string, val time.Time) {
	// TODO allow custom handling via an optional method
	x.str(event, key, formatTimestamp(val))
}

func (x modifierMethods[E]) duration(event E, key string, val time.Duration) {
	// TODO allow custom handling via an optional method
	x.str(event, key, formatDuration(val))
}

func (x modifierMethods[E]) int(event E, key string, val int) {
	if !event.AddInt(key, val) {
		event.AddField(key, val)
	}
}

func (x modifierMethods[E]) float32(event E, key string, val float32) {
	if !event.AddFloat32(key, val) {
		event.AddField(key, val)
	}
}

// formatTimestamp uses the same behavior as protobuf's timestamp.
// "1972-01-01T10:00:20.021Z"	Uses RFC 3339, where generated output will always be Z-normalized and uses 0, 3, 6 or 9 fractional digits. Offsets other than "Z" are also accepted.
func formatTimestamp(t time.Time) string {
	t = t.UTC()
	x := t.Format("2006-01-02T15:04:05.000000000") // RFC 3339
	x = strings.TrimSuffix(x, "000")
	x = strings.TrimSuffix(x, "000")
	x = strings.TrimSuffix(x, ".000")
	return x + "Z"
}

// formatDuration uses the same behavior as protobuf's duration.
// "1.000340012s", "1s"	Generated output always contains 0, 3, 6, or 9 fractional digits, depending on required precision, followed by the suffix "s". Accepted are any fractional digits (also none) as long as they fit into nano-seconds precision and the suffix "s" is required.
func formatDuration(d time.Duration) string {
	nanos := d.Nanoseconds()
	secs := nanos / 1e9
	nanos -= secs * 1e9
	//if nanos <= -1e9 || nanos >= 1e9 || (secs > 0 && nanos < 0) || (secs < 0 && nanos > 0) {
	//	panic("invalid duration")
	//}
	sign := ""
	if secs < 0 || nanos < 0 {
		sign, secs, nanos = "-", -1*secs, -1*nanos
	}
	x := fmt.Sprintf("%s%d.%09d", sign, secs, nanos)
	x = strings.TrimSuffix(x, "000")
	x = strings.TrimSuffix(x, "000")
	x = strings.TrimSuffix(x, ".000")
	return x + "s"
}