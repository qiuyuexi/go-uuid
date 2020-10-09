package redis

import (
	"errors"
	"fmt"
	"strconv"
)

type redisPtServer struct {
}

const clrf = "\r\n"
const firstCharacterByte = 42
const keyFirstCharacterByte = 36

const ErrProtocol = -2
const ProtocolLenMiss = -1
const ProtocolLenLegal = 0

/**
http://redisdoc.com/topic/protocol.html
*/

//获取命令个数及下标
func (s redisPtServer) getCmdInfo(binaryData []byte) (cmdNum int, cmdEndIndex int, errCode int) {

	binaryDataLen := len(binaryData)

	//至少4个字节 *1\r\n
	if binaryDataLen < 4 {
		return 0, 0, ProtocolLenMiss
	}

	//第一个字节必定是*
	if binaryData[0] != firstCharacterByte {
		return 0, 0, ErrProtocol
	}

	commandEndIndex, err := s.parse(binaryData)

	if err != nil {
		fmt.Println(err)
		return 0, 0, ProtocolLenMiss
	}

	//获取命令的个数
	commandNum, err := strconv.Atoi(string(binaryData[1 : commandEndIndex-1]))
	if err != nil {
		fmt.Println(err)
		return 0, 0, ErrProtocol
	}

	return commandNum, commandEndIndex, ProtocolLenLegal
}

func (s redisPtServer) getKeyLenInfo(keyStart int, binaryData []byte) (keyLenEndIndex int, keyLength int, errCode int) {

	binaryDataLen := len(binaryData)

	//偏移量是否大于字节数组长度
	if keyStart > binaryDataLen {
		return 0, 0, ProtocolLenMiss
	}

	//获取新命令字节数组
	newData := binaryData[keyStart:]
	if newData[0] != keyFirstCharacterByte {
		return 0, 0, ErrProtocol
	}

	//命令长度下标 \r\n \n的位置
	keyLenIndex, err := s.parse(newData)
	if err != nil {
		fmt.Println(err)
		return 0, 0, ProtocolLenMiss
	}

	if keyLenIndex == 0 {
		return 0, 0, ProtocolLenMiss
	}

	//命令长度
	keyLen, err := strconv.Atoi(string(newData[1 : keyLenIndex-1]))
	if err != nil {
		fmt.Println(err)
		return 0, 0, ErrProtocol
	}

	return keyStart + keyLenIndex, keyLen, ProtocolLenLegal
}

/**
-2 协议异常
-1 不够
0 刚好
> 0 长度过长，需要截断。
*/
func (s redisPtServer) CheckProtocol(binaryData []byte) int {
	binaryDataLen := len(binaryData)

	//获取key个数 及偏移量
	commandNum, offset, errCode := s.getCmdInfo(binaryData)

	if errCode != ProtocolLenLegal {
		return errCode
	}

	offset = offset + 1
	fmt.Println(string(binaryData))
	for i := 0; i < commandNum; i++ {

		//获取key的信息，key的长度
		keyLenEndIndex, keyLen, errCode := s.getKeyLenInfo(offset, binaryData)
		if errCode != ProtocolLenLegal {
			return errCode
		}

		//value的起始位置
		keyStartIndex := keyLenEndIndex + 1

		// slice[start,end]  最后一个数的下标是end-1.所以valueEndIndex是最后一个数的下一个下标
		keyEndIndex := keyStartIndex + keyLen + 2

		//value占用的字节范围，当前字节流是否符合。不符合，则表示协议不全。
		if keyEndIndex > binaryDataLen {
			fmt.Println(keyEndIndex, binaryDataLen)
			return ProtocolLenMiss
		}
		newData := binaryData[keyStartIndex:keyEndIndex]
		fmt.Println(binaryData, offset, keyLenEndIndex, keyStartIndex, keyEndIndex)

		//判断value是否合法 。最后俩个字节是\r\n
		keyLegal := s.parseValue(newData)
		if !keyLegal {
			return ErrProtocol
		}

		//下一行key的位置 ,因为valueEndIndex 就是下一个数的下标，不需要减一
		offset = keyEndIndex
	}

	//遍历结束，如果offset = binaryDataLen 则表示协议正常，无多余流
	if offset > binaryDataLen {
		return ProtocolLenMiss
	}

	if offset < binaryDataLen {
		return offset
	}

	return ProtocolLenLegal
}

/**
分析协议 获取key
*/
func (s redisPtServer) AnalysisProtocol(binaryData []byte) ([]string, error) {

	commandEnd, _ := s.parse(binaryData)

	//新的数据从commandEnd+1 开始
	commandNum, _ := strconv.Atoi(string(binaryData[1 : commandEnd-1]))
	offset := commandEnd

	keys := make([]string, 0, commandNum)
	for i := 0; i < commandNum; i++ {
		offset = offset + 1

		newData := binaryData[offset:]

		//命令长度
		kenLenIndex, _ := s.parse(newData)

		kenLen, _ := strconv.Atoi(string(newData[1 : kenLenIndex-1]))

		//具体内容 kenLenIndex + 1
		valueByte := newData[kenLenIndex+1 : kenLenIndex+1+kenLen]
		valueIndex := kenLen + 2

		//偏移量 +2 \r\n
		offset = offset + kenLenIndex + valueIndex

		keys = append(keys, string(valueByte))
	}
	return keys, nil
}

func (s redisPtServer) parse(binaryData []byte) (int, error) {
	length := len(binaryData)
	if length < 3 {
		return 0, errors.New("字节数组长度异常")
	}

	//从第2个字节开始查找\r\n
	i := 1
	for {
		if i+1 >= length {
			return 0, errors.New("字节数组长度异常")
		}
		if binaryData[i] == 13 && binaryData[i+1] == 10 {
			return i + 1, nil
		} else {
			i++
		}
	}
}

func (s redisPtServer) parseValue(binaryData []byte) bool {
	length := len(binaryData)
	if length < 3 {
		return false
	}
	//后往前扫 value 可以包含clrf
	i := length - 1
	if binaryData[i-1] == 13 && binaryData[i] == 10 {
		return true
	} else {
		return false
	}
}

func (s redisPtServer) Error(msg string) string {
	errMsg := "-ERR " + msg + clrf

	return errMsg
}

func (s redisPtServer) BulkReply(msg string) string {
	byteLength := len([]byte(msg))

	str := fmt.Sprintf("$%d%s%s%s", byteLength, clrf, msg, clrf)
	return str
}
