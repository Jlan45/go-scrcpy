package scrcpy

// Point is a 2-D position on the device screen in pixels.
type Point struct {
	X, Y int32
}

// Size is the width and height of the device screen in pixels.
type Size struct {
	Width, Height uint16
}

// DeviceInfo is received once per connection as the first thing the server
// sends.  The device name is the Android Build.MODEL string, truncated to 63
// UTF-8 bytes.
type DeviceInfo struct {
	DeviceName string
}

// VideoInfo is received on the video socket immediately after DeviceInfo.
type VideoInfo struct {
	// Codec identifies the video codec in use.
	Codec CodecID
	// Width and Height are the initial encoded dimensions.  They may change
	// if the device is rotated; the new dimensions arrive via a config Packet.
	Width, Height uint32
}

// AudioInfo is received on the audio socket immediately after DeviceInfo.
type AudioInfo struct {
	// Codec identifies the audio codec in use.  CodecDisabled means the
	// server was launched with --no-audio.
	Codec CodecID
}

// Packet is a single video or audio data unit read from a stream socket.
type Packet struct {
	// PTS is the presentation timestamp in microseconds.
	// It is meaningful only when IsConfig is false.
	PTS int64
	// IsConfig is true for codec configuration packets (H.264 SPS/PPS, etc.).
	// These must be passed to the decoder before any data packets.
	IsConfig bool
	// IsKeyframe is true when the packet is an intra (random-access) frame.
	IsKeyframe bool
	// Data is the raw codec payload (NAL units, Opus frames, etc.).
	Data []byte
}
