package types

type DeviceLog struct {
	Serverid  string
	Tagid     string
	Address   string
	Timestamp string
	Value     string
	Quality   string
}

func NewDeviceLog(tagid, address, timestamp, value, quality string) DeviceLog {
	return DeviceLog{
		Tagid:     tagid,
		Address:   address,
		Timestamp: timestamp,
		Value:     value,
		Quality:   quality,
	}

}
