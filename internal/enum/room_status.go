package enum

// SeatStatus 座位状态
type SeatStatus string

const (
	Booked SeatStatus = "C" // 已预约
	UnBook SeatStatus = "U" // 可预约
)

// IsAvailableBooking 是否可预约
func (s SeatStatus) IsAvailableBooking() bool {
	return s == UnBook
}
