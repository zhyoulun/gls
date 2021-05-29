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
	typeUserControl = 4

	typeAudio = 8
	typeVideo = 9

	typeDataAMF3 = 15
	typeDataAMF0 = 18

	typeSharedObjectAMF3 = 16
	typeSharedObjectAMF0 = 19

	typeCommandAMF3 = 17
	typeCommandAMF0 = 20

	//todo metadata?
	typeAggregate = 22
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
	commandConnect      = "connect"
	commandCall         = "call"
	commandCreateStream = "createStream"

	//NetStream client->server
	commandPlay         = "play"
	commandPlay2        = "play2"
	commandDeleteStream = "deleteStream"
	commandCloseStream  = "closeStream"
	commandReceiveAudio = "receiveAudio"
	commandReceiveVideo = "receiveVideo"
	commandPublish      = "publish"
	commandSeek         = "seek"
	commandPause        = "pause"

	//NetStream server->client
	commandOnStatus = "onStatus"
)

//flag values for the audioCodecs property
const (
	supportSndNone    = 0x0001
	supportSndAdpcm   = 0x0002
	supportSndMp3     = 0x0004
	supportSndIntel   = 0x0008
	supportSndUnused  = 0x0010
	supportSndNelly8  = 0x0020
	supportSndNelly   = 0x0040
	supportSndG711a   = 0x0080
	supportSndG711u   = 0x0100
	supportSndNelly16 = 0x0200
	supportSndAac     = 0x0400
	supportSndSpeex   = 0x0800
	supportSndAll     = 0x0FFF
)

//flag values for the videoCodecs property
const (
	supportVidUnused    = 0x0001
	supportVidJpeg      = 0x0002
	supportVidSorenson  = 0x0004
	supportVidHomebrew  = 0x0008
	supportVidVp6       = 0x0010
	supportVidVp6alpha  = 0x0020
	supportVidHomeBrewv = 0x0040
	supportVidH264      = 0x0080
	supportVidAll       = 0x00ff
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
