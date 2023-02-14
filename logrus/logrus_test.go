package logrus

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/joeycumines/go-utilpkg/logiface"
	"github.com/joeycumines/go-utilpkg/logiface/testsuite"
	"github.com/sirupsen/logrus"
	"io"
	"testing"
	"time"
)

var testSuiteConfig = testsuite.Config[*Event]{
	LoggerFactory:    testSuiteLoggerFactory,
	WriteTimeout:     time.Second * 10,
	AlertCallsOsExit: true,
	EmergencyPanics:  true,
	LogsEmptyMessage: true,
}

func testSuiteLoggerFactory(req testsuite.LoggerRequest[*Event]) testsuite.LoggerResponse[*Event] {
	logger := logrus.New()
	logger.Level = logrus.TraceLevel
	logger.Formatter = &logrus.JSONFormatter{
		DisableTimestamp: true,
	}
	logger.Out = req.Writer

	var options []logiface.Option[*Event]

	options = append(options, L.WithLogrus(logger))

	options = append(options, req.Options...)

	return testsuite.LoggerResponse[*Event]{
		Logger:       L.New(options...),
		LevelMapping: testSuiteLevelMapping,
		ParseEvent:   testSuiteParseEvent,
	}
}

func testSuiteLevelMapping(lvl logiface.Level) logiface.Level {
	if !lvl.Enabled() || lvl.Custom() {
		return logiface.LevelDisabled
	}
	switch lvl {
	case logiface.LevelNotice:
		return logiface.LevelWarning
	case logiface.LevelCritical:
		return logiface.LevelError
	default:
		return lvl
	}
}

func testSuiteParseEvent(r io.Reader) ([]byte, *testsuite.Event) {
	d := json.NewDecoder(r)
	var b json.RawMessage
	if err := d.Decode(&b); err != nil {
		if err == io.EOF {
			return nil, nil
		}
		panic(err)
	}

	var data struct {
		Level   *logrus.Level `json:"level"`
		Message *string       `json:"msg"`
		Error   *string       `json:"error"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		panic(err)
	}
	if data.Level == nil {
		panic(`expected logrus message to have a level`)
	}

	var fields map[string]interface{}
	if err := json.Unmarshal(b, &fields); err != nil {
		panic(err)
	}
	delete(fields, `level`)
	delete(fields, `msg`)
	delete(fields, `error`)

	ev := testsuite.Event{
		Message: data.Message,
		Error:   data.Error,
		Fields:  fields,
	}

	switch *data.Level {
	case logrus.TraceLevel:
		ev.Level = logiface.LevelTrace

	case logrus.DebugLevel:
		ev.Level = logiface.LevelDebug

	case logrus.InfoLevel:
		ev.Level = logiface.LevelInformational

	case logrus.WarnLevel:
		ev.Level = logiface.LevelWarning

	case logrus.ErrorLevel:
		ev.Level = logiface.LevelError

	case logrus.FatalLevel:
		ev.Level = logiface.LevelAlert

	case logrus.PanicLevel:
		ev.Level = logiface.LevelEmergency

	default:
		panic(fmt.Errorf(`unexpected logrus level %d`, *data.Level))
	}

	b, err := io.ReadAll(d.Buffered())
	if err != nil {
		panic(err)
	}

	return b, &ev
}

// Test_TestSuite runs the consolidated/shared test suite.
func Test_TestSuite(t *testing.T) {
	t.Parallel()
	testsuite.TestSuite(t, testSuiteConfig)
}

func TestLogger_simple(t *testing.T) {
	t.Parallel()

	type Harness struct {
		L *logiface.Logger[*Event]
		B bytes.Buffer
	}

	newHarness := func(t *testing.T, options ...logiface.Option[*Event]) *Harness {
		var h Harness
		l := logrus.New()
		l.Formatter = &logrus.TextFormatter{
			DisableColors:    true,
			DisableTimestamp: true,
		}
		l.Out = &h.B
		h.L = L.New(append([]logiface.Option[*Event]{L.WithLogrus(l)}, options...)...)
		return &h
	}

	t.Run(`basic log`, func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)

		h.L.Info().
			Log(`hello world`)

		h.L.Debug().
			Log(`wont show`)

		h.L.Warning().
			Log(`is warning`)

		h.L.Trace().
			Log(`wont show`)

		h.L.Err().
			Log(`is err`)

		if s := h.B.String(); s != "level=info msg=\"hello world\"\nlevel=warning msg=\"is warning\"\nlevel=error msg=\"is err\"\n" {
			t.Errorf("unexpected output: %q\n%s", s, s)
		}
	})

	t.Run(`with fields`, func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)

		h.L.Info().
			Field(`one`, 1).
			Field(`two`, 2).
			Field(`three`, 3).
			Log(`hello world`)

		if s := h.B.String(); s != "level=info msg=\"hello world\" one=1 three=3 two=2\n" {
			t.Errorf("unexpected output: %q\n%s", s, s)
		}
	})

	t.Run(`basic context usage`, func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)

		c1 := h.L.Clone().
			Field(`one`, 1).
			Field(`two`, 2).
			Logger()

		c1.Info().
			Field(`three`, 3).
			Field(`four`, 4).
			Log(`case 1`)

		h.L.Clone().
			Field(`six`, 6).
			Logger().
			Clone().
			Field(`seven`, 7).
			Logger().
			Info().
			Field(`eight`, 8).
			Log(`case 2`)

		c1.Info().
			Field(`three`, -3).
			Field(`five`, 5).
			Log(`case 3`)

		if s := h.B.String(); s != "level=info msg=\"case 1\" four=4 one=1 three=3 two=2\nlevel=info msg=\"case 2\" eight=8 seven=7 six=6\nlevel=info msg=\"case 3\" five=5 one=1 three=-3 two=2\n" {
			t.Errorf("unexpected output: %q\n%s", s, s)
		}
	})

	t.Run(`nil logger disabled`, func(t *testing.T) {
		t.Parallel()

		h := &Harness{}

		h.L.Info().
			Log(`hello world`)

		c1 := h.L.Clone().
			Field(`one`, 1).
			Field(`two`, 2).
			Logger()

		c1.Info().
			Field(`three`, 3).
			Field(`four`, 4).
			Log(`case 1`)

		h.L.Clone().
			Field(`six`, 6).
			Logger().
			Clone().
			Field(`seven`, 7).
			Logger().
			Info().
			Field(`eight`, 8).
			Log(`case 2`)

		c1.Info().
			Field(`three`, -3).
			Field(`five`, 5).
			Log(`case 3`)
	})

	t.Run(`add error`, func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)

		h.L.Info().
			Err(errors.New(`some error`)).
			Log(`hello world`)

		if s := h.B.String(); s != "level=info msg=\"hello world\" error=\"some error\"\n" {
			t.Errorf("unexpected output: %q\n%s", s, s)
		}
	})

	t.Run(`add error disabled`, func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)

		h.L.Debug().
			Err(errors.New(`some error`)).
			Log(`hello world`)

		h.L.Clone().
			Err(errors.New(`some error`)).
			Logger().
			Debug().
			Log(`hello world`)

		if s := h.B.String(); s != "" {
			t.Errorf("unexpected output: %q\n%s", s, s)
		}
	})
}
