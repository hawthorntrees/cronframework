package utils

import (
	"github.com/sony/sonyflake"
	"strconv"
	"time"
)

var _sf *sonyflake.Sonyflake
var _settings sonyflake.Settings

func InitSnowflake(workerID uint16) {
	_settings := sonyflake.Settings{
		StartTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		MachineID: func() (uint16, error) {
			return workerID, nil // 机器ID（范围 0-65535，默认16位）
		},
	}
	_sf = sonyflake.NewSonyflake(_settings)
	if _sf == nil {
		panic("创建 sonyflake 实例失败")
	}
}

func GenerateTraceID() (string, error) {
	id, err := _sf.NextID()
	return strconv.FormatInt(int64(id), 10), err
}

func GetTimestampAndMachineID(id uint64) (uint64, uint64) {
	timestamp := (id >> 22) + uint64(_settings.StartTime.UnixMilli())
	machineID := (id >> 6) & 0xffff

	return timestamp, machineID
}
