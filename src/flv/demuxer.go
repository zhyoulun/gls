package flv

import (
	"github.com/pkg/errors"
	"github.com/zhyoulun/gls/src/av"
	"github.com/zhyoulun/gls/src/core"
)

type Demuxer struct {
}

func NewDemuxer() *Demuxer {
	return &Demuxer{}
}

func (d *Demuxer) ParseAudioTag(b []byte) (av.AudioTagI, error) {
	tag := &AudioTag{}
	if len(b) < 1 {
		return nil, errors.Errorf("invalid audiodata length, want>=1, got: %d", len(b))
	}
	flags := b[0]
	tag.soundFormat = flags >> 4
	tag.soundRate = (flags >> 2) & 0x3
	tag.soundSize = (flags >> 1) & 0x1
	tag.soundType = flags & 0x1

	//validate
	if tag.soundFormat == SoundFormatAAC && tag.soundRate != SoundRate3 {
		return nil, errors.Wrapf(core.ErrorInvalidData, "for aac, SoundRate want: %d, got: %d", SoundRate3, tag.soundRate)
	}
	if tag.soundFormat == SoundFormatNellymoser && tag.soundType != SoundType0 {
		return nil, errors.Wrapf(core.ErrorInvalidData, "for nellymoser, SoundType want: %d, got: %d", SoundType0, tag.soundType)
	}
	if tag.soundFormat == SoundFormatAAC && tag.soundType != SoundType1 {
		return nil, errors.Wrapf(core.ErrorInvalidData, "for aac, SoundType want: %d, got: %d", SoundType1, tag.soundType)
	}

	if len(b) < 2 {
		return nil, errors.Errorf("invalid audiodata length, want>=2, got: %d", len(b))
	}
	if tag.soundFormat == SoundFormatAAC {
		tag.aacPacketType = b[1]
		if tag.aacPacketType != AACPacketTypeAACSequenceHeader && tag.aacPacketType != AACPacketTypeAACRaw {
			return nil, errors.Wrapf(core.ErrorInvalidData, "AACPacketType should be %d or %d, got: %d",
				AACPacketTypeAACSequenceHeader, AACPacketTypeAACRaw, tag.aacPacketType)
		}
	}
	return tag, nil
}

func (d *Demuxer) ParseVideoTag(b []byte) (av.VideoTagI, error) {
	tag := &VideoTag{}
	if len(b) < 5 {
		return nil, errors.Errorf("invalid videodata length, want>=5, got: %d", len(b))
	}
	flags := b[0]
	tag.frameType = flags >> 4
	tag.codecID = flags & 0xf
	if tag.frameType != FrameTypeInterFrame && tag.frameType != FrameTypeKeyFrame {
		return nil, errors.Errorf("invalid videodata frameType, want: %d, %d, got: %d",
			FrameTypeInterFrame, FrameTypeKeyFrame, tag.frameType)
	}
	if tag.codecID != codeIDAvc {
		return nil, errors.Errorf("invalid videodata codecID, want: %d, got: %d", codeIDAvc, tag.codecID)
	}

	tag.avcPacketType = b[1]
	if tag.avcPacketType == AVCPacketTypeAVCNALU {
		for i := 2; i < 5; i++ {
			tag.compositionTime = tag.compositionTime<<8 + int32(b[i])
		}
	} else {
		tag.compositionTime = 0
	}

	return tag, nil
}
