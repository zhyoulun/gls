package flv

type audioData struct {
	soundFormat uint8
	soundRate   uint8
	soundSize   uint8
	soundType   uint8
}

type aacAudioData struct {
	aacPacketType uint8
}

type videoData struct {
	frameType uint8
	codecID   uint8
}

type avcVideoPacket struct {
	avcPacketType   uint8
	compositionTime int32
}

type AudioTag struct {
	audioData
	aacAudioData
}

type VideoTag struct {
	videoData
	avcVideoPacket
}

func (at *AudioTag) SoundFormat() uint8 {
	return at.soundFormat
}

func (at *AudioTag) AACPacketType() uint8 {
	return at.aacPacketType
}

func (vt *VideoTag) FrameType() uint8 {
	return vt.frameType
}

func (vt *VideoTag) AVCPacketType() uint8 {
	return vt.avcPacketType
}
