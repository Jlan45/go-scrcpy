package scrcpy

import (
	"encoding/binary"
	"fmt"
	"io"
)

// DeviceMsgType is the first byte of every device message (server → client).
type DeviceMsgType uint8

const (
	DeviceMsgClipboard    DeviceMsgType = 0
	DeviceMsgAckClipboard DeviceMsgType = 1
	DeviceMsgUHIDOutput   DeviceMsgType = 2
)

// DeviceMsg is a message received from the scrcpy server on the control
// socket.  Inspect Type to determine which fields are populated.
type DeviceMsg struct {
	Type DeviceMsgType

	// DeviceMsgClipboard – clipboard text pushed by the device.
	ClipboardText string

	// DeviceMsgAckClipboard – acknowledgement for a previous SET_CLIPBOARD.
	// The sequence number matches the one sent in BuildSetClipboard.
	AckSequence uint64

	// DeviceMsgUHIDOutput – HID output report from a virtual HID device.
	UHID_ID   uint16
	UHID_Data []byte
}

// ReadDeviceMsg reads one device message from the control socket.
//
// Wire formats:
//
//	CLIPBOARD    [type:1][length:4][text:length]
//	ACK_CLIPBOARD [type:1][sequence:8]
//	UHID_OUTPUT  [type:1][id:2][dataLen:2][data:dataLen]
func ReadDeviceMsg(r io.Reader) (*DeviceMsg, error) {
	var typeBuf [1]byte
	if _, err := io.ReadFull(r, typeBuf[:]); err != nil {
		return nil, fmt.Errorf("scrcpy: read device message type: %w", err)
	}

	msg := &DeviceMsg{Type: DeviceMsgType(typeBuf[0])}

	switch msg.Type {
	case DeviceMsgClipboard:
		var lenBuf [4]byte
		if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
			return nil, fmt.Errorf("scrcpy: read clipboard length: %w", err)
		}
		length := binary.BigEndian.Uint32(lenBuf[:])
		text := make([]byte, length)
		if _, err := io.ReadFull(r, text); err != nil {
			return nil, fmt.Errorf("scrcpy: read clipboard text: %w", err)
		}
		msg.ClipboardText = string(text)

	case DeviceMsgAckClipboard:
		var seqBuf [8]byte
		if _, err := io.ReadFull(r, seqBuf[:]); err != nil {
			return nil, fmt.Errorf("scrcpy: read ack clipboard sequence: %w", err)
		}
		msg.AckSequence = binary.BigEndian.Uint64(seqBuf[:])

	case DeviceMsgUHIDOutput:
		var hdr [4]byte
		if _, err := io.ReadFull(r, hdr[:]); err != nil {
			return nil, fmt.Errorf("scrcpy: read uhid output header: %w", err)
		}
		msg.UHID_ID = binary.BigEndian.Uint16(hdr[0:2])
		dataLen := binary.BigEndian.Uint16(hdr[2:4])
		data := make([]byte, dataLen)
		if _, err := io.ReadFull(r, data); err != nil {
			return nil, fmt.Errorf("scrcpy: read uhid output data: %w", err)
		}
		msg.UHID_Data = data

	default:
		return nil, fmt.Errorf("scrcpy: unknown device message type: %d", msg.Type)
	}

	return msg, nil
}
