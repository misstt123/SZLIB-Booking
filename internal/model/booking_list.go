package model

import "SZLIB-Booking/internal/enum"

// Booking 预约记录
type Booking struct {
	Id            int           `json:"id"`
	Openid        string        `json:"openid"`
	Nickname      string        `json:"nickname"`
	Sex           int           `json:"sex"`
	Headimgurl    string        `json:"headimgurl"`
	Bookingdate   int           `json:"bookingdate"`
	Starttime     int           `json:"starttime"`
	Endtime       int           `json:"endtime"`
	Flag          string        `json:"flag"`
	Servaddr      enum.ServAddr `json:"servaddr"`
	Servaddrid    string        `json:"servaddrid"`
	Seatid        string        `json:"seatid"`
	Frontier      string        `json:"frontier"`
	Address       string        `json:"address"`
	Library       string        `json:"library"`
	Libnotes      string        `json:"libnotes"`
	Eventdatetime int           `json:"eventdatetime"`
}

type BookingRes struct {
	Message  string    `json:"message"`
	Rscount  int       `json:"rscount"`
	Bookings []Booking `json:"data"`
}
