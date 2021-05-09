package rtmp

type ConnectCommentObject struct {
	App            string
	Flashver       string
	SwfUrl         string
	TcUrl          string
	Fpad           bool
	AudioCodecs    int
	VideoCodecs    int
	VideoFunction  int
	PageUrl        string
	ObjectEncoding float64
	Type           string //todo 不在协议上
}
