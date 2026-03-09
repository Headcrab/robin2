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
