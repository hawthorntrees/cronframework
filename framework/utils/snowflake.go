package utils

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

// Snowflake 雪花算法生成器
type Snowflake struct {
	mu        sync.Mutex
	timestamp int64 // 毫秒级时间戳
	workerID  int64 // 工作节点ID（0-1023）
	sequence  int64 // 序列号（0-4095）
}

// 雪花算法参数（64位ID结构）
const (
	workerIDBits  = 10                          // 工作节点ID位数
	sequenceBits  = 12                          // 序列号位数
	maxWorkerID   = -1 ^ (-1 << workerIDBits)   // 最大工作节点ID：1023
	maxSequence   = -1 ^ (-1 << sequenceBits)   // 最大序列号：4095
	timeShift     = workerIDBits + sequenceBits // 时间戳左移位数：22
	workerIDShift = sequenceBits                // 工作节点ID左移位数：12
	epoch         = 1710000000000               // 起始时间戳（2024-03-08 12:00:00）
)

var (
	snowflake *Snowflake
	once      sync.Once
)

// InitSnowflake 初始化雪花算法（全局唯一）
func InitSnowflake(workerID int64) {
	once.Do(func() {
		if workerID < 1 || workerID > maxWorkerID {
			panic(fmt.Errorf("workerID必须在1-1023之间"))
		}
		snowflake = &Snowflake{
			workerID: workerID,
		}
	})
}

// GenerateTraceID 生成全局唯一追踪号（字符串格式）
func GenerateTraceID() (string, error) {
	id, err := GenerateID()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(id, 10), nil
}

// GenerateID 生成雪花算法ID（int64）
func GenerateID() (int64, error) {
	if snowflake == nil {
		return 0, errors.New("雪花算法未初始化，请先调用InitSnowflake")
	}
	return snowflake.nextID()
}

// nextID 生成下一个唯一ID
func (s *Snowflake) nextID() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()

	// 处理时钟回退（严重错误，直接返回）
	if now < s.timestamp {
		return 0, errors.New("系统时钟回退，无法生成唯一ID")
	}

	// 同一时间戳：序列号自增
	if now == s.timestamp {
		s.sequence = (s.sequence + 1) & maxSequence
		// 序列号用尽：等待下一毫秒
		if s.sequence == 0 {
			for now <= s.timestamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		// 新时间戳：序列号重置为0
		s.sequence = 0
	}

	s.timestamp = now

	// 组合ID：时间戳差 <<22 | workerID <<12 | sequence
	id := (now-epoch)<<timeShift | (s.workerID << workerIDShift) | s.sequence
	return id, nil
}
