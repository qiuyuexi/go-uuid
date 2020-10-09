package redis

type PtlServer interface {
	AnalysisProtocol(binaryData []byte) ([]string, error)
	Error(msg string) string
	BulkReply(msg string) string
	CheckProtocol(binaryData []byte) int
}

func GetRedisProtocolServer() PtlServer {
	return redisPtServer{}
}
