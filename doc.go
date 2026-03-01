// Package scrcpy implements the scrcpy client-side protocol for Android
// screen mirroring. It provides low-level functions to:
//
//   - Read the initial device/video/audio handshake from the server
//   - Parse video and audio stream packets
//   - Build control messages (client → server)
//   - Parse device messages (server → client)
//
// # Protocol Overview
//
// scrcpy v3.3.4 uses up to three separate TCP connections:
//
//   - Video stream: device name → video codec + initial size → packet loop
//   - Audio stream: device name → audio codec → packet loop
//   - Control stream: bidirectional control/device messages
//
// All multi-byte integers are encoded in big-endian (network) byte order.
//
// # Quick start
//
//	// Read handshake from video connection
//	info, _ := scrcpy.ReadDeviceInfo(videoConn)
//	video, _ := scrcpy.ReadVideoInfo(videoConn)
//
//	// Read video packets in a loop
//	for {
//	    pkt, _ := scrcpy.ReadPacket(videoConn)
//	    // pkt.Data contains raw H.264/H.265/AV1 NAL units
//	}
//
//	// Send a touch event on the control connection
//	msg := scrcpy.BuildInjectTouchEvent(
//	    scrcpy.MotionActionDown, scrcpy.PointerIDMouse,
//	    scrcpy.Point{X: 540, Y: 960},
//	    scrcpy.Size{Width: 1080, Height: 1920},
//	    1.0, 0, scrcpy.ButtonPrimary,
//	)
//	controlConn.Write(msg)
package scrcpy
