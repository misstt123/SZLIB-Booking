package main

import (
	"SZLIB-Booking/internal/enum"
	"SZLIB-Booking/internal/service"
	"github.com/go-co-op/gocron"
	"github.com/oklog/run"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	timezone, _ := time.LoadLocation("Asia/Shanghai")
	s := gocron.NewScheduler(timezone)

	g := run.Group{}
	g.Add(func() error {
		return runSchedule(s)
	}, func(err error) {
		s.Stop()
	})
	_ = g.Run()

	//// block until you are ready to shut down
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	select {
	case <-sigChan:
		s.Stop()
	}

}

// runSchedule 运行定时任务
func runSchedule(s *gocron.Scheduler) error {
	bookingService := &service.BookingService{
		PreferFloor:  enum.Floor3Dong,
		PreferSeatID: "129",
	}
	job, err := s.Every(1).Days().At("23:30").Do(func() {
		_ = bookingService.BookingRun()
	})
	if err != nil {
		log.Err(err).Str("job", job.GetName()).Msg("run job err")
		return err
	}
	s.StartBlocking()
	return nil
}
