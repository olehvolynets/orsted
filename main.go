package main

import (
	"context"
	"embed"
	"flag"
	"os/signal"
	"syscall"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
)

//go:embed *.mp3
//go:embed *.wav
var f embed.FS

var (
	bpm    = flag.Int("bpm", 120, "Tempo")
	accent = flag.Bool("accent", false, "Add accent beat")
)

func main() {
	flag.Parse()

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	interval := time.Duration(float64(time.Minute) / float64(*bpm))
	// As beep plays the stream concurrently by default, need to adjust the tick duration
	// to have some time for syncronisation to happen and start the next tick on time
	shortInterval := time.Duration(float64(interval) * 0.9)

	baseTick, format, err := decodeFile("tick.mp3")
	if err != nil {
		panic(err)
	}
	defer baseTick.Close()

	accentTick, _, err := decodeFile("tick.mp3")
	if err != nil {
		panic(err)
	}
	defer accentTick.Close()

	if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/1000)); err != nil {
		panic(err)
	}

	var t0 time.Time

	for {
		select {
		case <-ctx.Done():
			return
		default:
			t0 = time.Now()
			_ = baseTick.Seek(0)

			s := beep.Take(format.SampleRate.N(shortInterval), baseTick)
			var dur time.Duration
			cb := beep.Callback(func() { dur = time.Since(t0) })

			speaker.Play(beep.Seq(s, cb))

			time.Sleep(interval - dur)
		}
	}
}

func decodeFile(name string) (beep.StreamSeekCloser, beep.Format, error) {
	file, err := f.Open(name)
	if err != nil {
		return nil, beep.Format{}, err
	}

	return mp3.Decode(file)
}
