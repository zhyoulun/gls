package flv

//Format of SoundData
//Format 7,8,14 and 15 are reserved for internal use
//AAC is supported in Flash Player 9,0,115,0 and higher
//Speex is supported in Flash Player 10 and higher
const (
	SoundFormatLinearPCMPlatformEndian = 0  //Linear PCM, platform endian
	SoundFormatADPCM                   = 1  //ADPCM
	SoundFormatMP3                     = 2  //MP3
	SoundFormatLinearPCMLittleEndian   = 3  //Linear PCM, little endian
	SoundFormatNellymoser16kHzMono     = 4  //Nellymoser 16-kHz mono
	SoundFormatNellymoser8kHzMono      = 5  //Nellymoser 8-kHz mono
	SoundFormatNellymoser              = 6  //Nellymoser
	SoundFormatG711ALawLogarithmicPCM  = 7  //G.711 A-law logarithmic PCM
	SoundFormatG711MuLawLogarithmicPCM = 8  //G.711 mu-law logarithmic PCM
	SoundFormatReserved                = 9  //reserved
	SoundFormatAAC                     = 10 //AAC
	SoundFormatSpeex                   = 11 //Speex
	SoundMP38Khz                       = 14 //MP3 8-kHz
	SoundDeviceSpecificSound           = 15 //Device-specific sound
)

//Sampling rate
//For AAC: always 3
const (
	SoundRate0 = 0 //5.5-kHz
	SoundRate1 = 1 //11-kHz
	SoundRate2 = 2 //22-kHz
	SoundRate3 = 3 //44-kHz
)

//todo 这里的校验还没有做，因为不知道什么叫uncompressed format和compressed format
//Size of each sample
//this parameter only pertains to uncompressed formats.
//compressed format always decode to 16 bits internally.
const (
	SoundSize0 = 0 //snd8Bit
	SoundSize1 = 1 //snd16Bit
)

//Mono or stereo sound
//For Nellymonser: always 0
//For AAC: always 1
const (
	SoundType0 = 0 //sndMono
	SoundType1 = 1 //SndStereo
)

const (
	AACPacketTypeAACSequenceHeader = 0 //AAC sequence header
	AACPacketTypeAACRaw            = 1 //AAC raw
)

const (
	FrameTypeKeyFrame              = 1 //for avc, a seekable frame
	FrameTypeInterFrame            = 2 //for avc, a non-seekable frame
	FrameTypeDisposableInterFrame  = 3 //h.263 only
	FrameTypeGeneratedKeyFrame     = 4 //reserved for server use only
	FrameTypeVideoInfoCommandFrame = 5 //video info/command frame
)

const (
	codeIDJpeg                   = 1 //currently unused
	codeIDSorensonH263           = 2 //sorenson h.263
	codeIDScreenVideo            = 3 //screen video
	codeIDOn2VP6                 = 4 //on2 vp6
	codeIDOn2VP6WithAlphaChannel = 5 //on2 vp6 with alpha channel
	codeIDScreenVideoVersion2    = 6 //screen video version 2
	codeIDAvc                    = 7 //avc
)

const (
	AVCPacketTypeAVCSequenceHeader = 0 //avc sequence header
	AVCPacketTypeAVCNALU           = 1 //avc nalu
	AVCPacketTypeAVCEndOfSequence  = 2 //avc end of sequence(lower level nalu sequence ender is not required or supported)
)
