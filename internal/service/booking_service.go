package service

import (
	"SZLIB-Booking/internal/enum"
	"SZLIB-Booking/internal/model"
	"SZLIB-Booking/internal/pkg"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type BookingService struct {
	ReadKey      string
	PreferFloor  enum.ServAddr
	PreferSeatID string
}

// BookingRun 预约逻辑
func (b *BookingService) BookingRun() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(10*60))
	defer cancel()
	// 1.先通过喜好位置来筛选
	date := time.Now().Add(time.Hour * time.Duration(24)).Format("20060102")
	var err error
	defer func() {
		err = b.sendServerChan(ctx, date)
		if err != nil {
			log.Err(err).Str("date", date).Msg("send serverChan err")
		}
	}()

	if b.PreferSeatID != "" && b.PreferFloor != "" {
		err = b.bookingSeats(ctx, date, b.PreferFloor, b.PreferSeatID)
		if err != nil {
			log.Err(err).Str("date", date).Str("floor", b.PreferFloor.String()).Str("seatID", b.PreferSeatID)
		}
	}
	if err == nil {
		return nil
	}

	// 2. 找不到从最佳喜好楼层中找
	if b.PreferFloor == "" {
		return nil
	}
	seats, err := b.getLibrarySeat(ctx, b.PreferFloor, "")
	if err != nil {
		log.Err(err).Str("date", date).Str("floor", b.PreferFloor.String()).Msg("get library seats err")
		return err
	}
	var freeSeats []model.Seat
	for _, seat := range seats {
		if !seat.IsAvailableTimeRange(date) {
			continue
		}
		seatID, err := strconv.ParseInt(seat.Id, 10, 64)
		if err != nil {
			log.Err(err).Str("seatID", seat.Id).Str("floor", seat.Servaddr.String()).Msg("parse SeariD str err")
			continue
		}

		if (seatID-1)%4 == 0 {
			freeSeats = append(freeSeats, seat)
		}
	}

	// 3.随机选一个座位
	rand.Seed(time.Now().UnixNano())
	//isSuccess := false
	// 重试个五次吧
	for i := 0; i < 5; i++ {
		idx := rand.Intn(len(freeSeats) - 1)
		err = b.bookingSeats(ctx, date, b.PreferFloor, freeSeats[idx].Id)
		if err != nil {
			log.Err(err).Str("date", date).Int("count", i+1).Str("floor", b.PreferFloor.String()).Str("seatID", b.PreferSeatID)
		}
		if err == nil {
			//isSuccess = true
			break
		}
	}

	return nil
}

// getLibrarySeat 获取图书馆座位
func (b *BookingService) getLibrarySeat(ctx context.Context, floor enum.ServAddr, seatID string) ([]model.Seat, error) {
	response, err := pkg.GetClient().
		R().
		SetContext(ctx).
		SetResult(&model.Res{}).
		SetFormData(map[string]string{
			"servaddr": floor.String(),
			"seatid":   seatID,
			"smpsw":    b.smpsw(),
		}).
		Post(enum.GetLibrarySeatsURL)
	if err != nil {
		log.Err(err).Str("url", enum.GetLibrarySeatsURL)
		return nil, err
	}
	err = b.responseErrorHandle(ctx, response, err)
	if err != nil {
		return nil, err
	}
	result := response.Result()
	var res *model.Res
	var ok bool
	if res, ok = result.(*model.Res); !ok || res == nil {
		return nil, errors.New("invalid response")
	}
	if len(res.Room) == 0 || len(res.Room[0].Seat) == 0 {
		return nil, errors.New("invalid response")
	}

	return res.Room[0].Seat, nil
}

func (b *BookingService) smpsw() string {
	originStr := "Seat" + time.Now().Format("2006") + "Manage" + time.Now().Format("0102")
	//hash := md5.Sum([]byte(originStr))
	h := md5.New()
	h.Write([]byte(originStr))
	return hex.EncodeToString(h.Sum(nil))
}

// bookingSeats 预约位置
func (b *BookingService) bookingSeats(ctx context.Context, bookDate string, floor enum.ServAddr, seatId string) error {
	response, err := pkg.GetClient().
		R().
		SetContext(ctx).
		SetResult(&model.CommonRes{}).
		SetFormData(map[string]string{
			"servaddr":   floor.String(),
			"seatid":     seatId,
			"bookdate":   bookDate,
			"bookdata":   `[{"starttime":"09:00","endtime":"17:00"}]`,
			"readerkey":  b.ReadKey,
			"nickname":   "用户已经注销",
			"sex":        "0",
			"headimgurl": `https://thirdwx.qlogo.cn/mmopen/vi_32/PiajxSqBRaELSFYibJFEiaxK4ykiaibiclu7Ey6vOw3D7VVCbs1U5dma6nbFjIPJsZiaamH7ta7bsL8DEG4EMo01y9T3PgY2BE5ibAQh3y1ujzXOGx4u7gaHb9KTpA/132`,
			"smpsw":      b.smpsw(),
		}).
		Post(enum.BookingURL)
	if err != nil {
		log.Err(err).Str("url", enum.GetLibrarySeatsURL)
		return err
	}
	err = b.responseErrorHandle(ctx, response, err)
	if err != nil {
		return err
	}
	result := response.Result()
	var res *model.CommonRes
	var ok bool
	if res, ok = result.(*model.CommonRes); !ok || res == nil {
		return errors.New("invalid response")
	}
	if res.Result != "OK" {
		log.Error().Str("message", res.Message).Str("result", res.Result).Msg("预约失败!")
		return errors.New("预约失败")
	}
	log.Info().Str("message", res.Message).Str("result", res.Result).Msg("预约成功!")
	return nil
}

const (
	serverChanKey = "oNDDO1234445jFMWDYglkfdjhoiglc0ss"
	text          = `
楼层：%s  
座位号：%s`
)

// sendServerChan 微信通知
func (b *BookingService) sendServerChan(ctx context.Context, bookingDate string) error {
	//format := time.Now().Add(24 * time.Hour).Format("20060102")
	//format := "20240721"
	response, err := pkg.GetClient().
		R().
		SetContext(ctx).
		SetResult(&model.BookingRes{}).
		SetQueryParams(map[string]string{
			"bookingdate": bookingDate,
			"readerkey":   b.ReadKey,
			"type":        "",
			"pagesize":    "0",
			"pageindex":   "0",
			"smpsw":       b.smpsw(),
		}).Get(enum.GetBookingListURL)
	if err != nil {
		log.Err(err).Str("url", enum.GetBookingListURL)
		return err
	}
	err = b.responseErrorHandle(ctx, response, err)
	if err != nil {
		return err
	}
	result := response.Result()
	var res *model.BookingRes
	var ok bool
	if res, ok = result.(*model.BookingRes); !ok || res == nil {
		return errors.New("invalid response")
	}
	if res.Message != "OK" {
		return errors.New("请求失败")
	}

	bookings := res.Bookings
	//if len(bookings) == 0 {
	//	return errors.New("无记录")
	//}

	isSuccess := false
	var tmpBooking model.Booking
	for _, booking := range bookings {
		if fmt.Sprintf("%d", booking.Bookingdate) == bookingDate && booking.Flag == "Booking" {
			isSuccess = true
			tmpBooking = booking
			break
		}
	}

	data := url.Values{}

	dateStr := time.Now().Format("01-02")
	if isSuccess {
		data.Set("text", dateStr+"预约成功")
		data.Set("desp", fmt.Sprintf(text, tmpBooking.Servaddr, tmpBooking.Seatid))
	} else {
		data.Set("text", dateStr+"预约失败")
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("https://sctapi.ftqq.com/%s.send", serverChanKey), strings.NewReader(data.Encode()))
	if err != nil {
		log.Err(err).Msg("serverchan send fail")
		return nil
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		log.Err(err).Msg("serverchan send fail")
		return nil
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Err(err).Msg("serverchan read body err")
		return nil
	}

	if !isSuccess {
		return errors.New("预约失败")
	}

	return nil
}

func (b *BookingService) responseErrorHandle(ctx context.Context, rsp *resty.Response, err error) error {
	if err != nil || rsp == nil {
		return errors.New("http response error, err: " + err.Error())
	}
	code := rsp.StatusCode()
	if code != http.StatusOK && code != http.StatusCreated && code != http.StatusAccepted {
		//logger.Error(ctx, "fpx_call_http_code_error", log.Any("code", code), log.Any("rsp", rsp))
		log.Error().Int("code", code).Any("rsp", rsp).Msg("http code invalid")
		return errors.New("http status code invalid: " + strconv.Itoa(code))
	}
	if len(rsp.Body()) == 0 {
		log.Error().Msg("http response empty body")
		return errors.New("http response empty body")
	}
	return nil
}
