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
		ReadKey:      "oNDDO1234445jFMWDYglkfdjhoiglc0",
		PreferFloor:  enum.Floor3Dong,
		PreferSeatID: "129",
	}
	//job, err := s.Every(5).Seconds().Do(func() {
	//	_ = bookingService.BookingRun()
	//})

	// 1.添加预约任务
	job, err := s.Every(1).Days().At("10:00").Do(func() {
		time.Sleep(time.Duration(5) * time.Second)
		// 重试个3次吧
		for i := 0; i < 3; i++ {
			err := bookingService.BookingRun()
			if err == nil {
				break
			}
			time.Sleep(time.Duration(20) * time.Second)
		}
	})
	if err != nil {
		log.Err(err).Str("job", job.GetName()).Msg("run job err")
		return err
	}

	// 2.添加自动签到任务
	// todo ...

	s.StartBlocking()
	return nil
}
