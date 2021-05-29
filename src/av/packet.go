package av

import (
	"fmt"
)

type Packet struct {
	avType          uint8
	streamID        uint32
	timestamp       uint32
	data            []byte
	audioTagHandler AudioTagI
	videoTagHandler VideoTagI
}

func (p *Packet) ToCsvHeader() string {
	return fmt.Sprintf("avType,streamID,timestamp,dataLength\n")
}

func (p *Packet) ToCsvLine() string {
	return fmt.Sprintf("%d,%d,%d,%d\n", p.avType, p.streamID, p.timestamp, len(p.data))
}

type MessageI interface {
	GetAvType() (uint8, error)
	GetMessageStreamID() uint32
	GetTimestamp() uint32
	GetData() []byte
}

func NewPacket(csi MessageI, di DemuxerI) (*Packet, error) {
	var err error
	avType, err := csi.GetAvType()
	if err != nil {
		return nil, err
	}

	var audioTagHandler AudioTagI
	var videoTagHandler VideoTagI
	switch avType {
	case TypeAudio:
		if audioTagHandler, err = di.ParseAudioTag(csi.GetData()); err != nil {
			return nil, err
		}
	case TypeVideo:
		if videoTagHandler, err = di.ParseVideoTag(csi.GetData()); err != nil {
			return nil, err
		}
	case TypeMetadata:
		//todo
	}

	return &Packet{
		avType:          avType,
		streamID:        csi.GetMessageStreamID(),
		timestamp:       csi.GetTimestamp(),
		data:            csi.GetData(),
		audioTagHandler: audioTagHandler,
		videoTagHandler: videoTagHandler,
	}, nil
}

func (p *Packet) String() string {
	return fmt.Sprintf("packet info, avTypeName: %s, streamID: %d, timestamp: %d, data length: %d",
		p.getAvTypeName(), p.streamID, p.timestamp, len(p.data))
}

func (p *Packet) IsVideo() bool {
	return p.avType == TypeVideo
}

func (p *Packet) IsAudio() bool {
	return p.avType == TypeAudio
}

func (p *Packet) IsMetadata() bool {
	return p.avType == TypeMetadata
}

func (p *Packet) GetAvType() uint8 {
	return p.avType
}

func (p *Packet) getAvTypeName() string {
	if p.avType == TypeVideo {
		return "video"
	} else if p.avType == TypeAudio {
		return "audio"
	} else if p.avType == TypeMetadata {
		return "metadata"
	} else {
		return "unknown"
	}
}

func (p *Packet) GetData() []byte {
	return p.data
}

func (p *Packet) GetDataLength() uint32 {
	return uint32(len(p.data))
}

func (p *Packet) GetStreamID() uint32 {
	return p.streamID
}

func (p *Packet) GetTimestamp() uint32 {
	return p.timestamp
}

func (p *Packet) GetAudioTagHandler() AudioTagI {
	return p.audioTagHandler
}

func (p *Packet) GetVideoTagHandler() VideoTagI {
	return p.videoTagHandler
}

type AudioTagI interface {
	SoundFormat() uint8
	AACPacketType() uint8
}

type VideoTagI interface {
	FrameType() uint8
	AVCPacketType() uint8
}

type DemuxerI interface {
	ParseAudioTag(b []byte) (AudioTagI, error)
	ParseVideoTag(b []byte) (VideoTagI, error)
}
