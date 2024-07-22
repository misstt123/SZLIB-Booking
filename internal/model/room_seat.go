package model

import (
	"SZLIB-Booking/internal/enum"
	"fmt"
)

type Res struct {
	Room []Room `json:"room"`
	CommonRes
}

type CommonRes struct {
	Result  string `json:"result"`
	Message string `json:"message"`
}

// Room 房间信息
type Room struct {
	Seat     []Seat   `json:"seat"`
	Roompara RoomPara `json:"roompara"`
	ServAddr string   `json:"servaddr"`
}

// Seat 座位信息
type Seat struct {
	Id             string           `json:"id"`
	Frontier       string           `json:"frontier"`
	Servaddr       enum.ServAddr    `json:"servaddr"`
	Status         enum.SeatStatus  `json:"status"`
	Alarmflag      int              `json:"alarmflag"`
	Limittime      int              `json:"limittime"`
	Alarmtime      int              `json:"alarmtime"`
	Centerlocktime int              `json:"centerlocktime"`
	Selflocktime   int              `json:"selflocktime"`
	Enterdate      int              `json:"enterdate"`
	Entertime      int              `json:"entertime"`
	Starttime      int              `json:"starttime"`
	Leavetime      int              `json:"leavetime"`
	Pausetime      int              `json:"pausetime"`
	Reader         *Reader          `json:"reader,omitempty"`
	TomorrowQue    []TomorrowReader `json:"tomorrowque,omitempty"`
	NextReader     []NextReader     `json:"nextreader,omitempty"`
	Reducetime     string           `json:"reducetime"`
}

// IsAvailableTimeRange 可预约时段,只会预约09:00-17:00
func (s Seat) IsAvailableTimeRange(bookingDate string) bool {
	sTime, etime := 900, 1700
	if s.Reader != nil {
		if s.Reader.Bookingdate == bookingDate && (s.Reader.Starttime < etime || s.Reader.Endtime > sTime) {
			return false
		}
	}

	if len(s.NextReader) > 0 {
		for _, reader := range s.NextReader {
			if reader.Bookingdate == bookingDate && (reader.Starttime < etime || reader.Endtime > sTime) {
				return false
			}
		}
	}

	if len(s.TomorrowQue) > 0 {
		for _, reader := range s.TomorrowQue {
			if fmt.Sprintf("%d", reader.Bookingdate) == bookingDate && (reader.Starttime < etime || reader.Endtime > sTime) {
				return false
			}
		}
	}
	return true
}

type Reader struct {
	Nickname    string `json:"nickname"`
	Starttime   int    `json:"starttime"`
	Bookingdate string `json:"bookingdate"`
	Bookingid   string `json:"bookingid"`
	Openid      string `json:"openid"`
	Gender      string `json:"gender"`
	Headimgurl  string `json:"headimgurl"`
	Endtime     int    `json:"endtime"`
}

// NextReader 读者信息
type NextReader struct {
	Bookingid   string `json:"bookingid"`
	Openid      string `json:"openid"`
	Nickname    string `json:"nickname"`
	Gender      string `json:"gender"`
	Headimgurl  string `json:"headimgurl"`
	Bookingdate string `json:"bookingdate"`
	Starttime   int    `json:"starttime"`
	Endtime     int    `json:"endtime"`
}

// TomorrowReader 读者信息
type TomorrowReader struct {
	Bookingid   string `json:"bookingid"`
	Openid      string `json:"openid"`
	Nickname    string `json:"nickname"`
	Gender      string `json:"gender"`
	Headimgurl  string `json:"headimgurl"`
	Bookingdate string `json:"bookingdate"`
	Starttime   int    `json:"starttime"`
	Endtime     int    `json:"endtime"`
}

// RoomPara 房间参数
type RoomPara struct {
	Servaddr         string `json:"servaddr"`
	Serviceid        string `json:"serviceid"`
	Frontier         string `json:"frontier"`
	Seatnum          int    `json:"seatnum"`
	Limittime        int    `json:"limittime"`
	Centerlocktime   int    `json:"centerlocktime"`
	Selflocktime     int    `json:"selflocktime"`
	Alarmtime        int    `json:"alarmtime"`
	Sysmsg           string `json:"sysmsg"`
	Statusafterleave string `json:"statusafterleave"`
	Datingtime       string `json:"datingtime"`
	Week             []struct {
		Id        string `json:"id"`
		Notes     string `json:"notes"`
		Status    string `json:"status"`
		Opentime  string `json:"opentime"`
		Closetime string `json:"closetime"`
	} `json:"week"`
	Print         bool   `json:"print"`
	Position      string `json:"position"`
	Servaddrnotes string `json:"servaddrnotes"`
	Addrnotes     string `json:"addrnotes"`
	Library       string `json:"library"`
	Libnotes      string `json:"libnotes"`
}
