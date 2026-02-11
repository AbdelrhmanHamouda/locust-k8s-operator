package main

import (
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
)

// testArrayEncoder is a minimal PrimitiveArrayEncoder that captures appended strings.
type testArrayEncoder struct {
	strings []string
}

func (e *testArrayEncoder) AppendString(s string)                      { e.strings = append(e.strings, s) }
func (e *testArrayEncoder) AppendBool(bool)                            {}
func (e *testArrayEncoder) AppendByteString([]byte)                    {}
func (e *testArrayEncoder) AppendComplex128(complex128)                {}
func (e *testArrayEncoder) AppendComplex64(complex64)                  {}
func (e *testArrayEncoder) AppendFloat64(float64)                      {}
func (e *testArrayEncoder) AppendFloat32(float32)                      {}
func (e *testArrayEncoder) AppendInt(int)                              {}
func (e *testArrayEncoder) AppendInt64(int64)                          {}
func (e *testArrayEncoder) AppendInt32(int32)                          {}
func (e *testArrayEncoder) AppendInt16(int16)                          {}
func (e *testArrayEncoder) AppendInt8(int8)                            {}
func (e *testArrayEncoder) AppendUint(uint)                            {}
func (e *testArrayEncoder) AppendUint64(uint64)                        {}
func (e *testArrayEncoder) AppendUint32(uint32)                        {}
func (e *testArrayEncoder) AppendUint16(uint16)                        {}
func (e *testArrayEncoder) AppendUint8(uint8)                          {}
func (e *testArrayEncoder) AppendUintptr(uintptr)                      {}
func (e *testArrayEncoder) AppendDuration(time.Duration)               {}
func (e *testArrayEncoder) AppendTime(time.Time)                       {}
func (e *testArrayEncoder) AppendArray(zapcore.ArrayMarshaler) error   { return nil }
func (e *testArrayEncoder) AppendObject(zapcore.ObjectMarshaler) error { return nil }
func (e *testArrayEncoder) AppendReflected(interface{}) error          { return nil }

func TestColorize(t *testing.T) {
	tests := []struct {
		name string
		code int
		text string
		want string
	}{
		{"green", 32, "hello", "\x1b[32mhello\x1b[0m"},
		{"yellow", 33, "world", "\x1b[33mworld\x1b[0m"},
		{"empty", 32, "", "\x1b[32m\x1b[0m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := colorize(tt.code, tt.text)
			if got != tt.want {
				t.Errorf("colorize(%d, %q) = %q, want %q", tt.code, tt.text, got, tt.want)
			}
		})
	}
}

func TestGreenTimeEncoder(t *testing.T) {
	enc := &testArrayEncoder{}
	ts := time.Date(2026, 2, 10, 14, 30, 0, 0, time.UTC)

	greenTimeEncoder(ts, enc)

	if len(enc.strings) != 1 {
		t.Fatalf("expected 1 appended string, got %d", len(enc.strings))
	}

	want := "\x1b[32m2026-02-10T14:30:00.000Z\x1b[0m"
	if enc.strings[0] != want {
		t.Errorf("greenTimeEncoder output = %q, want %q", enc.strings[0], want)
	}
}

func TestYellowNameEncoder(t *testing.T) {
	enc := &testArrayEncoder{}

	yellowNameEncoder("setup", enc)

	if len(enc.strings) != 1 {
		t.Fatalf("expected 1 appended string, got %d", len(enc.strings))
	}

	want := "\x1b[33msetup\x1b[0m"
	if enc.strings[0] != want {
		t.Errorf("yellowNameEncoder output = %q, want %q", enc.strings[0], want)
	}
}
