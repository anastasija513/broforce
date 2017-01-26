package bus

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/InnovaCo/broforce/logger"
)

func TestSimple(t *testing.T) {
	logger.New(os.Stderr, "panic")

	t.Run("SafeHandler", func(t *testing.T) {
		Retry := uint32(0)

		f := func(e Event) error {
			Retry++
			return fmt.Errorf("Error number %d", Retry)
		}

		SafeHandler(f, SafeParams{Retry: 3, Delay: 1})(Event{})

		if Retry != 4 {
			t.Errorf("Retry %d != 4", Retry)
			t.Fail()
		}
	})

	t.Run("PubSub", func(t *testing.T) {
		got := false
		a := &simpleAdapter{}
		a.Run()

		handler := func(e Event) error {
			got = true
			return nil
		}

		a.Subscribe(UnknownEvent, handler)
		a.Publish(Event{Subject: UnknownEvent, Data: []byte(""), Coding: JsonCoding})

		time.Sleep(1 * time.Second)

		if !got {
			t.Error("Not got")
			t.Fail()
		}
	})
}
