package uuid

import (
	"fmt"
	"sync"
	"time"
)

type snowflakeData struct {
	sequenceId   int64
	workId       int64
	curTimestamp int64
}
type snowflake struct {
	m    *sync.Mutex
	data *snowflakeData
}

const sequenceMask = 4095
const maxWordId = 1023

func getSnowflakeData(workId int) *snowflakeData {
	obj := new(snowflakeData)
	obj.sequenceId = 0
	obj.workId = int64(workId & maxWordId)
	obj.curTimestamp = time.Now().UnixNano() / 1e6
	return obj
}

func (s snowflake) GetId() int64 {
	defer s.m.Unlock()
	s.m.Lock()
	var workId, sequenceId, timestamp int64
	workId = s.data.workId << 12
	curTime := getTimestamp()
	if s.data.curTimestamp == curTime {
		sequenceId = (s.data.sequenceId + 1) & sequenceMask
		if sequenceId == 0 {
			sequenceId = 1
			time.Sleep(1 * time.Millisecond)
			curTime = getTimestamp()
			fmt.Println("等待下一毫秒")
		}
	} else {
		sequenceId = 1
	}
	s.data.sequenceId = sequenceId
	s.data.curTimestamp = curTime
	timestamp = curTime << 22
	id := timestamp | workId | sequenceId
	return id
}

func getTimestamp() int64 {
	timestamp := time.Now().UnixNano() / 1e6
	return timestamp
}
