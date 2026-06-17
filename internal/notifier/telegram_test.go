package notifier

import (
	"testing"
)

func TestSendStateChangeAlert_InvalidToken(t *testing.T) {
	err := SendStateChangeAlert("", "123", "test", "Online", "Offline", "HTTP 500 | timeout")
	if err == nil {
		t.Error("expected error for empty token")
	}
}

func TestSendStateChangeAlert_InvalidChatID(t *testing.T) {
	err := SendStateChangeAlert("invalid:token", "", "test", "Online", "Offline", "details")
	if err == nil {
		t.Error("expected error for empty chat_id")
	}
}

func TestSendStateChangeAlert_BadToken(t *testing.T) {
	err := SendStateChangeAlert("bad_token", "123", "test", "Online", "Offline", "details")
	if err == nil {
		t.Error("expected error for bad token")
	}
}

func TestSendStateChangeAlert_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SendStateChangeAlert panicked: %v", r)
		}
	}()

	_ = SendStateChangeAlert("", "", "", "Online", "Offline", "")
}
