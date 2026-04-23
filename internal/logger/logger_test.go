package logger

import "testing"

func TestFatalPanics(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected panic from Fatal")
		}

		fatalErr, ok := recovered.(FatalError)
		if !ok {
			t.Fatalf("expected FatalError panic, got %T", recovered)
		}
		if fatalErr.Msg != "boom" {
			t.Fatalf("unexpected fatal message: %q", fatalErr.Msg)
		}
	}()

	Fatal("boom")
}

func TestParseLogLineNormalizesFatalLevel(t *testing.T) {
	item := parseLogLine(`{"time":"2026-04-23T15:43:42.058388+05:00","level":"ERROR+1","msg":"boom"}`)

	if item.Level != "FATAL" {
		t.Fatalf("expected FATAL level, got %q", item.Level)
	}
	if item.Msg != "boom" {
		t.Fatalf("unexpected message: %q", item.Msg)
	}
}
