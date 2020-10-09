package main

import (
	"fmt"
	"goredis/internal/redis"
	"goredis/internal/uuid"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

func main() {
	c := make(chan os.Signal)
	signal.Notify(c)

	ln, err := net.Listen("tcp", "127.0.0.1:9999")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	uuidServer := uuid.GetUuidServer(1)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			err = conn.SetReadDeadline(time.Now().Add(time.Duration(30) * time.Second))
			if err != nil {
				fmt.Println("err", err.Error())
			}
			go handleConnection(conn, uuidServer)
		}
	}()
	s := <-c
	ln.Close()
	fmt.Println("退出信号", s)
	os.Exit(0)
}

func handleConnection(conn net.Conn, uuidServer uuid.Server) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recover...:", r)
		}
		err := conn.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	redisPtlServer := redis.GetRedisProtocolServer()

	binaryData := make([]byte, 0, 1024)

	for {

		b := make([]byte, 1024)

		l, err := conn.Read(b)
		if err != nil {
			fmt.Println("error:", err.Error())
			break
		}

		if l > 0 {
			binaryData = append(binaryData, b[0:l]...)

			//判断字节流是否符合协议。是否存在一个字节流存在多条协议，或者协议不全
			match := redisPtlServer.CheckProtocol(binaryData)
			var matchData []byte
			switch match {
			case redis.ErrProtocol:
				binaryData = binaryData[:0]
				respMsg := redisPtlServer.Error("协议异常，解析出错:" + string(binaryData))
				_, writeErr := conn.Write([]byte(respMsg))
				if writeErr != nil {
					fmt.Println(writeErr)
					break
				}
				continue
			case redis.ProtocolLenMiss:
				continue

			case redis.ProtocolLenLegal:
				matchData = binaryData
				binaryData = binaryData[:0]
			default:
				matchData = binaryData[:match]
				binaryData = binaryData[match:]

			}
			keys, err := redisPtlServer.AnalysisProtocol(matchData)
			fmt.Println(keys)
			if err != nil {
				_, err = conn.Write([]byte(redisPtlServer.Error(err.Error())))
				if err != nil {
					fmt.Println(err.Error())
				}
				continue
			}

			if len(keys) < 0 {
				binaryData = binaryData[:0]
				respMsg := redisPtlServer.Error("协议异常，解析出错:" + string(binaryData))
				_, writeErr := conn.Write([]byte(respMsg))
				if writeErr != nil {
					fmt.Println(writeErr)
					break
				}
				continue
			}
			respMsg := ""
			key := keys[0]
			key = strings.ToLower(key)

			switch key {
			case "get":
				if len(keys) == 2 {
					/*clientId, err := strconv.Atoi(keys[1])
					if err != nil {
						fmt.Println(err)
						clientId = 0
					}*/
					clientIdUuid := uuidServer.GetId()
					respMsg = redisPtlServer.BulkReply(strconv.FormatInt(clientIdUuid, 10))
				} else {
					respMsg = redisPtlServer.Error("协议异常，解析出错")
				}
			case "auth":
				respMsg = "+OK\r\n"
			case "ping":
				respMsg = "+PONG\r\n"
			default:
				respMsg = redisPtlServer.Error("unsupport cmd:" + key)
			}
			_, writeErr := conn.Write([]byte(respMsg))
			if writeErr != nil {
				fmt.Println(writeErr)
				break
			}
		}
	}
}
