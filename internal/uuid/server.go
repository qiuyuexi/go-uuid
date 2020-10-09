package uuid

import "sync"

type Server interface {
	GetId() int64
}

func GetUuidServer(workId int) Server {
	mutex := new(sync.Mutex)
	return snowflake{m: mutex, data: getSnowflakeData(workId)}
}
