package syncservice

type syncServiceMessage struct {
	id      string
	msgType string
	version string
	payload []byte
}
