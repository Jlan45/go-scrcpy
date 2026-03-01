package scrcpy

import (
	"encoding/binary"
	"fmt"
	"io"
)

// deviceNameLen is the fixed-size field the server uses for the device name.
const deviceNameLen = 64

// ReadDeviceInfo reads the 64-byte device-name field that the server sends as
// the very first thing on every connection (video, audio, and control).
func ReadDeviceInfo(r io.Reader) (*DeviceInfo, error) {
	buf := make([]byte, deviceNameLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("scrcpy: read device name: %w", err)
	}
	return &DeviceInfo{DeviceName: cstringLen(buf)}, nil
}

// ReadVideoInfo reads the video-stream handshake that follows the device name
// on the video socket:
//
//	codec  (4 bytes, big-endian)
//	width  (4 bytes, big-endian)
//	height (4 bytes, big-endian)
func ReadVideoInfo(r io.Reader) (*VideoInfo, error) {
	var buf [12]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return nil, fmt.Errorf("scrcpy: read video info: %w", err)
	}
	return &VideoInfo{
		Codec:  CodecID(binary.BigEndian.Uint32(buf[0:4])),
		Width:  binary.BigEndian.Uint32(buf[4:8]),
		Height: binary.BigEndian.Uint32(buf[8:12]),
	}, nil
}

// ReadAudioInfo reads the audio-stream handshake that follows the device name
// on the audio socket:
//
//	codec (4 bytes, big-endian)
func ReadAudioInfo(r io.Reader) (*AudioInfo, error) {
	var buf [4]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return nil, fmt.Errorf("scrcpy: read audio info: %w", err)
	}
	return &AudioInfo{
		Codec: CodecID(binary.BigEndian.Uint32(buf[:])),
	}, nil
}

// ReadPacket reads one video or audio packet from a stream socket.
//
// Packet wire format (12-byte header followed by payload):
//
//	pts_with_flags  uint64  (bit 63 = config, bit 62 = keyframe, bits 0-61 = PTS µs)
//	data_size       uint32
//	data            [data_size]byte
func ReadPacket(r io.Reader) (*Packet, error) {
	var hdr [12]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return nil, fmt.Errorf("scrcpy: read packet header: %w", err)
	}

	ptsFlags := binary.BigEndian.Uint64(hdr[0:8])
	dataLen := binary.BigEndian.Uint32(hdr[8:12])

	isConfig := (ptsFlags & PacketFlagConfig) != 0
	isKeyframe := (ptsFlags & PacketFlagKeyframe) != 0
	pts := int64(ptsFlags & PacketPTSMask)

	data := make([]byte, dataLen)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, fmt.Errorf("scrcpy: read packet data: %w", err)
	}

	return &Packet{
		PTS:        pts,
		IsConfig:   isConfig,
		IsKeyframe: isKeyframe,
		Data:       data,
	}, nil
}

// cstringLen converts a null-padded byte slice to a Go string.
func cstringLen(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}
