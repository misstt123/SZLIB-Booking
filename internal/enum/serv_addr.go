package enum

// ServAddr 楼层区域
type ServAddr string

const (
	Floor2Dong ServAddr = "STSeat-2E"
	Floor2XI   ServAddr = "STSeat-2W"
	Floor3Dong ServAddr = "STSeat-3E"
	Floor3XI   ServAddr = "STSeat-3W"
)

func (a ServAddr) String() string {
	return string(a)
}
