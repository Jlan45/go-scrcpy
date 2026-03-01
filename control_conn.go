package scrcpy

import (
	"io"
	"sync"
)

// ControlConn wraps a scrcpy control socket and provides:
//
//   - Concurrency-safe control message sending (Send and all method wrappers).
//   - A background goroutine that continuously reads device messages and
//     delivers them through a buffered channel ([ControlConn.DeviceMsgs]).
//
// Typical usage:
//
//	ctrl := scrcpy.NewControlConn(conn, 16)
//	defer ctrl.Close()
//
//	// Send from any goroutine
//	ctrl.InjectTouchEvent(...)
//	ctrl.SetClipboard(1, true, "hello")
//
//	// Receive device messages in a goroutine
//	go func() {
//	    for msg := range ctrl.DeviceMsgs() {
//	        switch msg.Type {
//	        case scrcpy.DeviceMsgClipboard:
//	            fmt.Println("clipboard:", msg.ClipboardText)
//	        case scrcpy.DeviceMsgAckClipboard:
//	            fmt.Println("ack seq:", msg.AckSequence)
//	        }
//	    }
//	}()
type ControlConn struct {
	rw   io.ReadWriter
	mu   sync.Mutex    // serialises concurrent writes
	msgs chan *DeviceMsg
	stop chan struct{}
	once sync.Once
	err  error
}

// NewControlConn wraps rw (typically a net.Conn after the handshake) as a
// [ControlConn].
//
// bufSize is the capacity of the device-message channel returned by
// [ControlConn.DeviceMsgs].  0 creates an unbuffered channel where each
// incoming device message blocks the reader goroutine until it is consumed.
//
// Call [ControlConn.Close] when done.  To interrupt an in-progress read,
// close the underlying net.Conn directly.
func NewControlConn(rw io.ReadWriter, bufSize int) *ControlConn {
	c := &ControlConn{
		rw:   rw,
		msgs: make(chan *DeviceMsg, bufSize),
		stop: make(chan struct{}),
	}
	go c.readLoop()
	return c
}

func (c *ControlConn) readLoop() {
	defer close(c.msgs)
	for {
		// Check for shutdown before blocking on the network read.
		select {
		case <-c.stop:
			return
		default:
		}

		msg, err := ReadDeviceMsg(c.rw)
		if err != nil {
			c.err = err
			return
		}

		select {
		case c.msgs <- msg:
		case <-c.stop:
			return
		}
	}
}

// Send writes a pre-serialised control message to the socket.
// It is safe to call from multiple goroutines concurrently.
func (c *ControlConn) Send(msg []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.rw.Write(msg)
	return err
}

// DeviceMsgs returns the read-only channel of device messages pushed by the
// server.  The channel is closed when the underlying connection ends or an
// error occurs.  Use [ControlConn.Err] after the channel closes to inspect
// the cause.
func (c *ControlConn) DeviceMsgs() <-chan *DeviceMsg {
	return c.msgs
}

// Err returns the error that caused the device-message reader to stop.
// It is only meaningful after the channel returned by [ControlConn.DeviceMsgs]
// has been closed.  Returns nil when the reader stopped via Close.
func (c *ControlConn) Err() error {
	return c.err
}

// Close signals the background reader goroutine to stop.
// It does not close the underlying connection.
func (c *ControlConn) Close() {
	c.once.Do(func() { close(c.stop) })
}

// ---------------------------------------------------------------------------
// Convenience methods – thin wrappers around the Build* functions.
// Each method calls Send internally and returns any write error.
// ---------------------------------------------------------------------------

// InjectKeycode sends an INJECT_KEYCODE message.
func (c *ControlConn) InjectKeycode(action KeyEventAction, keycode KeyCode, repeat uint32, metaState MetaState) error {
	return c.Send(BuildInjectKeycode(action, keycode, repeat, metaState))
}

// InjectText sends an INJECT_TEXT message.
func (c *ControlConn) InjectText(text string) error {
	return c.Send(BuildInjectText(text))
}

// InjectTouchEvent sends an INJECT_TOUCH_EVENT message.
func (c *ControlConn) InjectTouchEvent(action MotionEventAction, pointerID int64, pos Point, screenSize Size, pressure float32, actionButton, buttons MotionEventButtons) error {
	return c.Send(BuildInjectTouchEvent(action, pointerID, pos, screenSize, pressure, actionButton, buttons))
}

// InjectScrollEvent sends an INJECT_SCROLL_EVENT message.
func (c *ControlConn) InjectScrollEvent(pos Point, screenSize Size, hScroll, vScroll float32, buttons MotionEventButtons) error {
	return c.Send(BuildInjectScrollEvent(pos, screenSize, hScroll, vScroll, buttons))
}

// BackOrScreenOn sends a BACK_OR_SCREEN_ON message.
func (c *ControlConn) BackOrScreenOn(action KeyEventAction) error {
	return c.Send(BuildBackOrScreenOn(action))
}

// ExpandNotificationPanel sends an EXPAND_NOTIFICATION_PANEL message.
func (c *ControlConn) ExpandNotificationPanel() error {
	return c.Send(BuildExpandNotificationPanel())
}

// ExpandSettingsPanel sends an EXPAND_SETTINGS_PANEL message.
func (c *ControlConn) ExpandSettingsPanel() error {
	return c.Send(BuildExpandSettingsPanel())
}

// CollapsePanels sends a COLLAPSE_PANELS message.
func (c *ControlConn) CollapsePanels() error {
	return c.Send(BuildCollapsePanels())
}

// GetClipboard sends a GET_CLIPBOARD message.
func (c *ControlConn) GetClipboard(copyKey GetClipboardCopyKey) error {
	return c.Send(BuildGetClipboard(copyKey))
}

// SetClipboard sends a SET_CLIPBOARD message.
func (c *ControlConn) SetClipboard(sequence uint64, paste bool, text string) error {
	return c.Send(BuildSetClipboard(sequence, paste, text))
}

// SetDisplayPower sends a SET_DISPLAY_POWER message.
func (c *ControlConn) SetDisplayPower(on bool) error {
	return c.Send(BuildSetDisplayPower(on))
}

// RotateDevice sends a ROTATE_DEVICE message.
func (c *ControlConn) RotateDevice() error {
	return c.Send(BuildRotateDevice())
}

// UHIDCreate sends a UHID_CREATE message.
func (c *ControlConn) UHIDCreate(id, vendorID, productID uint16, name string, reportDesc []byte) error {
	return c.Send(BuildUHIDCreate(id, vendorID, productID, name, reportDesc))
}

// UHIDInput sends a UHID_INPUT message.
func (c *ControlConn) UHIDInput(id uint16, data []byte) error {
	return c.Send(BuildUHIDInput(id, data))
}

// UHIDDestroy sends a UHID_DESTROY message.
func (c *ControlConn) UHIDDestroy(id uint16) error {
	return c.Send(BuildUHIDDestroy(id))
}

// OpenHardKeyboardSettings sends an OPEN_HARD_KEYBOARD_SETTINGS message.
func (c *ControlConn) OpenHardKeyboardSettings() error {
	return c.Send(BuildOpenHardKeyboardSettings())
}

// StartApp sends a START_APP message.
func (c *ControlConn) StartApp(name string) error {
	return c.Send(BuildStartApp(name))
}

// ResetVideo sends a RESET_VIDEO message.
func (c *ControlConn) ResetVideo() error {
	return c.Send(BuildResetVideo())
}
