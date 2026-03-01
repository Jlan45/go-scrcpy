package scrcpy

// CodecID identifies a video or audio codec.
// The value is a 4-byte ASCII code packed into a uint32 (big-endian).
type CodecID uint32

// Video codec IDs.
const (
	CodecH264 CodecID = 0x68323634 // "h264"
	CodecH265 CodecID = 0x68323635 // "h265"
	CodecAV1  CodecID = 0x00617631 // "\x00av1"
)

// Audio codec IDs.
const (
	CodecOpus CodecID = 0x6f707573 // "opus"
	CodecAAC  CodecID = 0x00616163 // "\x00aac"
	CodecFLAC CodecID = 0x666c6163 // "flac"
	CodecRAW  CodecID = 0x00726177 // "\x00raw"
)

// Special codec sentinel values sent by the server.
const (
	// CodecDisabled is sent when the server has the stream disabled.
	CodecDisabled CodecID = 0x00000000
	// CodecError is sent when the server could not negotiate a codec.
	CodecError CodecID = 0x00000001
)

// Packet flag bits embedded in the upper two bits of the PTS field.
const (
	// PacketFlagConfig marks a codec configuration (SPS/PPS) packet.
	PacketFlagConfig uint64 = 1 << 63
	// PacketFlagKeyframe marks an intra (keyframe) packet.
	PacketFlagKeyframe uint64 = 1 << 62
	// PacketPTSMask masks out the flag bits, leaving only the PTS value.
	PacketPTSMask uint64 = ^(PacketFlagConfig | PacketFlagKeyframe)
)
