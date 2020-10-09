package redis

import (
	"fmt"
	"testing"
)

func TestRedisSP_Explain(t *testing.T) {
	data := []byte("*2\r\n$3\r\nGET\r\n$1\r\n1\r\n")
	resp := redisPtServer{}
	l, err := resp.parse(data)
	if err != nil {
		t.Error(err)
	}
	if l != 3 {
		t.Error(l)
	}

}
func TestRedisSP_GetKeys(t *testing.T) {
	data := []byte("*2\r\n$3\r\nGET\r\n$1\r\n1\r\n")
	resp := redisPtServer{}
	keys, err := resp.AnalysisProtocol(data)
	if err != nil {
		fmt.Println(err)
	}
	targetKeys := []string{"GET", "1"}
	if len(keys) != len(targetKeys) {
		t.Error("长度不等")
	}
	for k, v := range keys {
		if v != targetKeys[k] {
			t.Error(v, targetKeys[k])
		}
	}
}

func TestRedisPtServer_CheckProtocol(t *testing.T) {
	data := []byte("*2\r\n$3\r\nGET\r\n$1\r\n1\r\n")
	resp := redisPtServer{}
	l := resp.CheckProtocol(data)
	if l != ProtocolLenLegal {
		t.Errorf("期望%d,实际%d", ProtocolLenLegal, l)
	}
	data1 := []byte("*2\r\n$3\r\nGET\r\n$1\r\n1\r")
	l1 := resp.CheckProtocol(data1)
	if l1 != ProtocolLenMiss {
		t.Errorf("期望%d,实际%d", ProtocolLenMiss, l1)
	}
	data2 := []byte("*2\r\n$3\r\nGET\r\n$1\r\n1\r\n*2\r\n$3\r\nGET\r\n$1\r\n1\r\n")
	l2 := resp.CheckProtocol(data2)
	if l2 != 20 {
		t.Errorf("期望%d,实际%d", 20, l2)
	}
	data4 := []byte("*2\r\n$3\r\nDEL\r\n$1\r\n1\r\n")
	l4 := resp.CheckProtocol(data4)
	if l != ProtocolLenLegal {
		t.Errorf("期望%d,实际%d", ProtocolLenLegal, l4)
	}

	data5 := []byte("*2\r\n$3\r\nDEL\r\n$1\r\n122\r\n")
	l5 := resp.CheckProtocol(data5)
	if l5 != ErrProtocol {
		t.Errorf("期望%d,实际%d", ErrProtocol, l5)
	}
}
