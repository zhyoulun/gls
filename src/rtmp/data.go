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
	Type           string //todo 协议上没有，livego和ffmpeg有
}
