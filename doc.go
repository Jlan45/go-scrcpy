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
// scrcpy uses a reverse tunnel by default: the Android server connects back
// to the desktop, so the desktop must listen first.
//
//	// adb reverse localabstract:scrcpy tcp:27183
//	ln, _ := net.Listen("tcp", "127.0.0.1:27183")
//	videoConn, _ := ln.Accept() // video (first)
//	audioConn, _ := ln.Accept() // audio (second)
//	ctrlConn,  _ := ln.Accept() // control (third)
//
//	sess, _ := scrcpy.NewSession(videoConn, audioConn, ctrlConn, scrcpy.SessionOptions{})
//	defer sess.Close()
//
//	for frame := range sess.Frames.Frames() {
//	    // frame.Config != nil → reinit decoder
//	    // frame.Data = encoded NAL units / AV1 OBUs
//	}
//
//	sess.Control.InjectTouchEvent(
//	    scrcpy.MotionActionDown, scrcpy.PointerIDMouse,
//	    scrcpy.Point{X: 540, Y: 960},
//	    scrcpy.Size{Width: 1080, Height: 1920},
//	    1.0, 0, scrcpy.ButtonPrimary,
//	)
package scrcpy
