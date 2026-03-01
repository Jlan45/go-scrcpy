package scrcpy

import "encoding/binary"

// ControlMsgType is the first byte of every control message.
type ControlMsgType uint8

const (
	ControlMsgInjectKeycode               ControlMsgType = 0
	ControlMsgInjectText                  ControlMsgType = 1
	ControlMsgInjectTouchEvent            ControlMsgType = 2
	ControlMsgInjectScrollEvent           ControlMsgType = 3
	ControlMsgBackOrScreenOn              ControlMsgType = 4
	ControlMsgExpandNotificationPanel     ControlMsgType = 5
	ControlMsgExpandSettingsPanel         ControlMsgType = 6
	ControlMsgCollapsePanels              ControlMsgType = 7
	ControlMsgGetClipboard                ControlMsgType = 8
	ControlMsgSetClipboard                ControlMsgType = 9
	ControlMsgSetDisplayPower             ControlMsgType = 10
	ControlMsgRotateDevice                ControlMsgType = 11
	ControlMsgUHIDCreate                  ControlMsgType = 12
	ControlMsgUHIDInput                   ControlMsgType = 13
	ControlMsgUHIDDestroy                 ControlMsgType = 14
	ControlMsgOpenHardKeyboardSettings    ControlMsgType = 15
	ControlMsgStartApp                    ControlMsgType = 16
	ControlMsgResetVideo                  ControlMsgType = 17
)

// ---------------------------------------------------------------------------
// INJECT_KEYCODE  –  14 bytes total
//
//	[type:1][action:1][keycode:4][repeat:4][metastate:4]
// ---------------------------------------------------------------------------

// BuildInjectKeycode builds an INJECT_KEYCODE control message.
//
//   - action: KeyEventActionDown or KeyEventActionUp
//   - keycode: one of the Keycode* constants
//   - repeat: repeat count (0 for a normal key press)
//   - metaState: OR of Meta* flag constants (0 for no modifier)
func BuildInjectKeycode(action KeyEventAction, keycode KeyCode, repeat uint32, metaState MetaState) []byte {
	buf := make([]byte, 14)
	buf[0] = byte(ControlMsgInjectKeycode)
	buf[1] = byte(action)
	binary.BigEndian.PutUint32(buf[2:6], uint32(keycode))
	binary.BigEndian.PutUint32(buf[6:10], repeat)
	binary.BigEndian.PutUint32(buf[10:14], uint32(metaState))
	return buf
}

// ---------------------------------------------------------------------------
// INJECT_TEXT  –  5 + len(text) bytes
//
//	[type:1][length:4][text:length]
// ---------------------------------------------------------------------------

// BuildInjectText builds an INJECT_TEXT control message.
// The text is UTF-8 encoded; the server limits it to 8191 bytes.
func BuildInjectText(text string) []byte {
	encoded := []byte(text)
	buf := make([]byte, 5+len(encoded))
	buf[0] = byte(ControlMsgInjectText)
	binary.BigEndian.PutUint32(buf[1:5], uint32(len(encoded)))
	copy(buf[5:], encoded)
	return buf
}

// ---------------------------------------------------------------------------
// INJECT_TOUCH_EVENT  –  32 bytes total
//
//	[type:1][action:1][pointerID:8][x:4][y:4][screenW:2][screenH:2]
//	[pressure:2][actionButton:4][buttons:4]
// ---------------------------------------------------------------------------

// BuildInjectTouchEvent builds an INJECT_TOUCH_EVENT control message.
//
//   - action: one of the MotionAction* constants
//   - pointerID: unique finger/pointer identifier; use PointerIDMouse for mouse
//   - pos: touch position in screen pixels
//   - screenSize: current screen dimensions (needed for server-side scaling)
//   - pressure: contact pressure in [0.0, 1.0]
//   - actionButton: button that triggered the action (0 for touch)
//   - buttons: currently pressed button mask
func BuildInjectTouchEvent(
	action MotionEventAction,
	pointerID int64,
	pos Point,
	screenSize Size,
	pressure float32,
	actionButton, buttons MotionEventButtons,
) []byte {
	buf := make([]byte, 32)
	buf[0] = byte(ControlMsgInjectTouchEvent)
	buf[1] = byte(action)
	binary.BigEndian.PutUint64(buf[2:10], uint64(pointerID))
	binary.BigEndian.PutUint32(buf[10:14], uint32(pos.X))
	binary.BigEndian.PutUint32(buf[14:18], uint32(pos.Y))
	binary.BigEndian.PutUint16(buf[18:20], screenSize.Width)
	binary.BigEndian.PutUint16(buf[20:22], screenSize.Height)
	binary.BigEndian.PutUint16(buf[22:24], floatToU16FP(pressure))
	binary.BigEndian.PutUint32(buf[24:28], uint32(actionButton))
	binary.BigEndian.PutUint32(buf[28:32], uint32(buttons))
	return buf
}

// ---------------------------------------------------------------------------
// INJECT_SCROLL_EVENT  –  21 bytes total
//
//	[type:1][x:4][y:4][screenW:2][screenH:2][hScroll:2][vScroll:2][buttons:4]
// ---------------------------------------------------------------------------

// BuildInjectScrollEvent builds an INJECT_SCROLL_EVENT control message.
//
//   - pos: pointer position in screen pixels
//   - screenSize: current screen dimensions
//   - hScroll: horizontal scroll delta, clamped to [-1.0, 1.0]
//   - vScroll: vertical scroll delta, clamped to [-1.0, 1.0]
//   - buttons: currently pressed button mask
func BuildInjectScrollEvent(
	pos Point,
	screenSize Size,
	hScroll, vScroll float32,
	buttons MotionEventButtons,
) []byte {
	buf := make([]byte, 21)
	buf[0] = byte(ControlMsgInjectScrollEvent)
	binary.BigEndian.PutUint32(buf[1:5], uint32(pos.X))
	binary.BigEndian.PutUint32(buf[5:9], uint32(pos.Y))
	binary.BigEndian.PutUint16(buf[9:11], screenSize.Width)
	binary.BigEndian.PutUint16(buf[11:13], screenSize.Height)
	binary.BigEndian.PutUint16(buf[13:15], floatToI16FP(hScroll))
	binary.BigEndian.PutUint16(buf[15:17], floatToI16FP(vScroll))
	binary.BigEndian.PutUint32(buf[17:21], uint32(buttons))
	return buf
}

// ---------------------------------------------------------------------------
// BACK_OR_SCREEN_ON  –  2 bytes total
//
//	[type:1][action:1]
// ---------------------------------------------------------------------------

// BuildBackOrScreenOn builds a BACK_OR_SCREEN_ON control message.
// When the screen is off, the action wakes it; when on, it simulates Back.
func BuildBackOrScreenOn(action KeyEventAction) []byte {
	return []byte{byte(ControlMsgBackOrScreenOn), byte(action)}
}

// ---------------------------------------------------------------------------
// One-byte panel / device commands
// ---------------------------------------------------------------------------

// BuildExpandNotificationPanel builds an EXPAND_NOTIFICATION_PANEL message.
func BuildExpandNotificationPanel() []byte {
	return []byte{byte(ControlMsgExpandNotificationPanel)}
}

// BuildExpandSettingsPanel builds an EXPAND_SETTINGS_PANEL message.
func BuildExpandSettingsPanel() []byte {
	return []byte{byte(ControlMsgExpandSettingsPanel)}
}

// BuildCollapsePanels builds a COLLAPSE_PANELS message.
func BuildCollapsePanels() []byte {
	return []byte{byte(ControlMsgCollapsePanels)}
}

// BuildRotateDevice builds a ROTATE_DEVICE message.
func BuildRotateDevice() []byte {
	return []byte{byte(ControlMsgRotateDevice)}
}

// BuildOpenHardKeyboardSettings builds an OPEN_HARD_KEYBOARD_SETTINGS message.
func BuildOpenHardKeyboardSettings() []byte {
	return []byte{byte(ControlMsgOpenHardKeyboardSettings)}
}

// BuildResetVideo builds a RESET_VIDEO message, requesting the server to
// restart the video stream (e.g. after a resolution change).
func BuildResetVideo() []byte {
	return []byte{byte(ControlMsgResetVideo)}
}

// ---------------------------------------------------------------------------
// GET_CLIPBOARD  –  2 bytes total
//
//	[type:1][copyKey:1]
// ---------------------------------------------------------------------------

// BuildGetClipboard builds a GET_CLIPBOARD message.
// copyKey controls whether the server triggers a copy/cut before reading.
func BuildGetClipboard(copyKey GetClipboardCopyKey) []byte {
	return []byte{byte(ControlMsgGetClipboard), byte(copyKey)}
}

// ---------------------------------------------------------------------------
// SET_CLIPBOARD  –  14 + len(text) bytes
//
//	[type:1][sequence:8][paste:1][length:4][text:length]
// ---------------------------------------------------------------------------

// BuildSetClipboard builds a SET_CLIPBOARD message.
//
//   - sequence: monotonically increasing number; the server echoes it in
//     ACK_CLIPBOARD so the caller can match acknowledgements.
//   - paste: if true, the server will paste the new text immediately.
//   - text: UTF-8 clipboard content (max 262 135 bytes).
func BuildSetClipboard(sequence uint64, paste bool, text string) []byte {
	encoded := []byte(text)
	buf := make([]byte, 14+len(encoded))
	buf[0] = byte(ControlMsgSetClipboard)
	binary.BigEndian.PutUint64(buf[1:9], sequence)
	if paste {
		buf[9] = 1
	}
	binary.BigEndian.PutUint32(buf[10:14], uint32(len(encoded)))
	copy(buf[14:], encoded)
	return buf
}

// ---------------------------------------------------------------------------
// SET_DISPLAY_POWER  –  2 bytes total
//
//	[type:1][on:1]
// ---------------------------------------------------------------------------

// BuildSetDisplayPower builds a SET_DISPLAY_POWER message.
func BuildSetDisplayPower(on bool) []byte {
	v := byte(DisplayPowerModeOff)
	if on {
		v = byte(DisplayPowerModeOn)
	}
	return []byte{byte(ControlMsgSetDisplayPower), v}
}

// ---------------------------------------------------------------------------
// UHID_CREATE  –  variable length
//
//	[type:1][id:2][vendorID:2][productID:2][nameLen:1][name:nameLen]
//	[descLen:2][descriptor:descLen]
//
// name is truncated to 127 bytes (write_string_tiny limit in scrcpy v3).
// ---------------------------------------------------------------------------

// BuildUHIDCreate builds a UHID_CREATE message that registers a virtual HID
// device on the Android side.
//
//   - id: caller-chosen device identifier (reused in UHID_INPUT / UHID_DESTROY)
//   - vendorID, productID: USB vendor/product IDs
//   - name: human-readable device name (max 127 bytes UTF-8)
//   - reportDesc: HID report descriptor bytes
func BuildUHIDCreate(id, vendorID, productID uint16, name string, reportDesc []byte) []byte {
	nameBuf := []byte(name)
	if len(nameBuf) > 127 {
		nameBuf = nameBuf[:127]
	}
	// Layout: type(1) + id(2) + vendorID(2) + productID(2) + nameLen(1) + name + descLen(2) + desc
	buf := make([]byte, 10+len(nameBuf)+len(reportDesc))
	buf[0] = byte(ControlMsgUHIDCreate)
	binary.BigEndian.PutUint16(buf[1:3], id)
	binary.BigEndian.PutUint16(buf[3:5], vendorID)
	binary.BigEndian.PutUint16(buf[5:7], productID)
	buf[7] = uint8(len(nameBuf)) // 1-byte length (write_string_tiny)
	copy(buf[8:], nameBuf)
	off := 8 + len(nameBuf)
	binary.BigEndian.PutUint16(buf[off:off+2], uint16(len(reportDesc)))
	copy(buf[off+2:], reportDesc)
	return buf
}

// ---------------------------------------------------------------------------
// UHID_INPUT  –  5 + len(data) bytes
//
//	[type:1][id:2][dataLen:2][data:dataLen]
// ---------------------------------------------------------------------------

// BuildUHIDInput builds a UHID_INPUT message that sends HID input data to a
// previously registered virtual HID device.
func BuildUHIDInput(id uint16, data []byte) []byte {
	buf := make([]byte, 5+len(data))
	buf[0] = byte(ControlMsgUHIDInput)
	binary.BigEndian.PutUint16(buf[1:3], id)
	binary.BigEndian.PutUint16(buf[3:5], uint16(len(data)))
	copy(buf[5:], data)
	return buf
}

// ---------------------------------------------------------------------------
// UHID_DESTROY  –  3 bytes total
//
//	[type:1][id:2]
// ---------------------------------------------------------------------------

// BuildUHIDDestroy builds a UHID_DESTROY message that unregisters a virtual
// HID device previously created with UHID_CREATE.
func BuildUHIDDestroy(id uint16) []byte {
	buf := make([]byte, 3)
	buf[0] = byte(ControlMsgUHIDDestroy)
	binary.BigEndian.PutUint16(buf[1:3], id)
	return buf
}

// ---------------------------------------------------------------------------
// START_APP  –  2 + len(name) bytes
//
//	[type:1][nameLen:1][name:nameLen]
//
// name is truncated to 255 bytes (write_string_tiny limit in scrcpy v3).
// ---------------------------------------------------------------------------

// BuildStartApp builds a START_APP message that asks the server to launch an
// application by package name or activity name (max 255 bytes UTF-8).
func BuildStartApp(name string) []byte {
	encoded := []byte(name)
	if len(encoded) > 255 {
		encoded = encoded[:255]
	}
	buf := make([]byte, 2+len(encoded))
	buf[0] = byte(ControlMsgStartApp)
	buf[1] = uint8(len(encoded)) // 1-byte length (write_string_tiny)
	copy(buf[2:], encoded)
	return buf
}

// ---------------------------------------------------------------------------
// Fixed-point helpers (mirrors scrcpy's sc_float_to_u16fp / sc_float_to_i16fp)
// ---------------------------------------------------------------------------

// floatToU16FP converts a float32 in [0.0, 1.0] to an unsigned 16-bit
// fixed-point value (used for touch pressure).
func floatToU16FP(v float32) uint16 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 0xffff
	}
	return uint16(v * 0xffff)
}

// floatToI16FP converts a float32 in [-1.0, 1.0] to a signed 16-bit
// fixed-point value (used for scroll deltas).
func floatToI16FP(v float32) uint16 {
	if v <= -1 {
		return 0x8000 // two's-complement of int16(-32768)
	}
	if v >= 1 {
		return 0x7fff
	}
	if v < 0 {
		return uint16(int16(v * 0x8000))
	}
	return uint16(int16(v * 0x7fff))
}
