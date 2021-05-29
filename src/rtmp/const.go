package rtmp

const (
	rtmpVersion = 3
)

type Fmt byte

const (
	fmt0 Fmt = 0x00
	fmt1 Fmt = 0x01
	fmt2 Fmt = 0x02
	fmt3 Fmt = 0x03
)

const (
	minValidMaximumChunkSize      = 1
	maxValidMaximumChunkSize      = 0x7fffffff
	defaultRemoteMaximumChunkSize = 128
	defaultLocalMaximumChunkSize  = 1024
)

const (
	max3BTimestamp = 0xffffff
)

const (
	messageStreamID0 = 0
)

const (
	chunkStreamID2 = 2
)

//message type id
const (
	//protocol control message
	//these protocol control message must have message stream id 0(messageStreamID0)(known as the control stream)
	//	and be sent in chunk stream id 2(chunkStreamID2)
	typeSetChunkSize              = 1
	typeAbort                     = 2
	typeAcknowledgement           = 3
	typeWindowAcknowledgementSize = 5
	typeSetPeerBandwidth          = 6

	//user control message
	//use control messages should use message stream id 0(messageStreamID0)(known as the control stream)
	//	when sent over rtmp chunk stream, be send on chunk stream id 2(chunkStreamID2)
	typeUserControl = 4

	typeAudio = 8
	typeVideo = 9

	typeDataAMF3 = 15
	typeDataAMF0 = 18

	typeSharedObjectAMF3 = 16
	typeSharedObjectAMF0 = 19

	typeCommandAMF3 = 17
	typeCommandAMF0 = 20

	typeAggregate = 22
)

const (
	transactionID0 = 0
	transactionID1 = 1
)

//user control message events
const (
	EventStreamBegin      = 0
	EventStreamEOF        = 1
	EventStreamDry        = 2
	EventSetBufferLength  = 3
	EventStreamIsRecorded = 4
	EventPingRequest      = 6
	EventPingResponse     = 7
)

//command name list
const (
	//NetConnection
	commandNetConnectionConnect      = "connect"
	commandNetConnectionCall         = "call"
	commandNetConnectionCreateStream = "createStream"

	//NetStream client->server
	commandNetStreamPlay         = "play"
	commandNetStreamPlay2        = "play2"
	commandNetStreamDeleteStream = "deleteStream"
	commandNetStreamCloseStream  = "closeStream"
	commandNetStreamReceiveAudio = "receiveAudio"
	commandNetStreamReceiveVideo = "receiveVideo"
	commandNetStreamPublish      = "publish"
	commandNetStreamSeek         = "seek"
	commandNetStreamPause        = "pause"

	//NetStream server->client
	commandNetStreamOnStatus = "onStatus"
)

const (
	netStreamStatusLevelWarning = "warning"
	netStreamStatusLevelStatus  = "status"
	netStreamStatusLevelError   = "error"
)

//flag values for the audioCodecs property
const (
	audioCodecSupportSndNone    = 0x0001
	audioCodecSupportSndAdpcm   = 0x0002
	audioCodecSupportSndMp3     = 0x0004
	audioCodecSupportSndIntel   = 0x0008
	audioCodecSupportSndUnused  = 0x0010
	audioCodecSupportSndNelly8  = 0x0020
	audioCodecSupportSndNelly   = 0x0040
	audioCodecSupportSndG711a   = 0x0080
	audioCodecSupportSndG711u   = 0x0100
	audioCodecSupportSndNelly16 = 0x0200
	audioCodecSupportSndAac     = 0x0400
	audioCodecSupportSndSpeex   = 0x0800
	audioCodecSupportSndAll     = 0x0FFF
)

//flag values for the videoCodecs property
const (
	videoCodecSupportVidUnused    = 0x0001
	videoCodecSupportVidJpeg      = 0x0002
	videoCodecSupportVidSorenson  = 0x0004
	videoCodecSupportVidHomebrew  = 0x0008
	videoCodecSupportVidVp6       = 0x0010
	videoCodecSupportVidVp6alpha  = 0x0020
	videoCodecSupportVidHomeBrewv = 0x0040
	videoCodecSupportVidH264      = 0x0080
	videoCodecSupportVidAll       = 0x00ff
)

//flag values for the videoFunction property
const (
	supportVidClientSeek = 1
)

const (
	limitTypeHard    = 0
	limitTypeSoft    = 1
	limitTypeDynamic = 2
)
