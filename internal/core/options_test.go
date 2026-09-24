package core

import (
	"errors"
	"testing"
	"time"
)

func TestAudioBitrates(t *testing.T) {
	cases := []struct {
		f    AudioFormat
		q    AudioQuality
		want int
	}{{AudioMP3, AudioEconomy, 128}, {AudioM4A, AudioStandard, 128}, {AudioOpus, AudioHigh, 160}}
	for _, c := range cases {
		got, err := AudioBitrate(c.f, c.q)
		if err != nil || got != c.want {
			t.Fatalf("%s/%s=%d,%v", c.f, c.q, got, err)
		}
	}
}
func TestJobTransitions(t *testing.T) {
	j := Job{Status: JobPending}
	now := time.Now()
	if err := j.Transition(JobDownloading, now); err != nil {
		t.Fatal(err)
	}
	if err := j.Transition(JobReady, now); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("expected invalid transition, got %v", err)
	}
	if err := j.Transition(JobCompleted, now); err != nil {
		t.Fatal(err)
	}
	if j.FinishedAt == nil {
		t.Fatal("terminal transition must set FinishedAt")
	}
}
func TestInvalidOptions(t *testing.T) {
	if _, err := AudioBitrate("wav", AudioHigh); !errors.Is(err, ErrInvalidOption) {
		t.Fatal(err)
	}
	if err := OutputMode("wat").Validate(); !errors.Is(err, ErrInvalidOption) {
		t.Fatal(err)
	}
}
