package scrcpy

import (
	"fmt"
	"io"
)

// SessionOptions configures the buffer sizes used by the background workers
// started inside [NewSession].  Zero values fall back to the listed defaults.
type SessionOptions struct {
	// VideoBufSize is the [FrameDemuxer] channel capacity (default: 30).
	VideoBufSize int
	// AudioBufSize is the [PacketBuffer] channel capacity (default: 60).
	AudioBufSize int
	// CtrlBufSize is the [ControlConn] device-message channel capacity
	// (default: 16).
	CtrlBufSize int
}

func (o *SessionOptions) applyDefaults() {
	if o.VideoBufSize <= 0 {
		o.VideoBufSize = 30
	}
	if o.AudioBufSize <= 0 {
		o.AudioBufSize = 60
	}
	if o.CtrlBufSize <= 0 {
		o.CtrlBufSize = 16
	}
}

// Session holds the fully-initialised high-level objects for all three scrcpy
// stream connections.  Create one with [NewSession].
//
// Fields for disabled streams are nil (e.g. Frames is nil when the server was
// started with --no-video, Control is nil when started with --no-control).
type Session struct {
	// Device is the Android device name, read from the first successful
	// handshake.
	Device DeviceInfo

	// Video holds the codec and initial resolution reported during the video
	// handshake.  Valid only when Frames != nil.
	Video VideoInfo

	// Audio holds the audio codec reported during the audio handshake.
	// Valid only when AudioPkts != nil.
	Audio AudioInfo

	// Frames delivers decoded video frames.  nil if videoConn was nil.
	Frames *FrameDemuxer

	// AudioPkts delivers raw audio packets.  nil if audioConn was nil.
	AudioPkts *PacketBuffer

	// Control manages the control socket.  nil if ctrlConn was nil.
	Control *ControlConn
}

// NewSession performs the scrcpy handshake on each provided connection and
// starts the background worker goroutines.
//
// Pass nil for connections that are not in use:
//
//	// video + control only (--no-audio)
//	sess, err := scrcpy.NewSession(videoConn, nil, ctrlConn, scrcpy.SessionOptions{})
//
// At least one connection must be non-nil.  Each non-nil connection must
// already be a raw TCP (or any io.ReadWriter) connection — port-forwarding
// and ADB setup are outside the scope of this library.
//
// If any handshake step fails, all already-started workers are stopped before
// the error is returned.
func NewSession(videoConn, audioConn, ctrlConn io.ReadWriter, opts SessionOptions) (*Session, error) {
	opts.applyDefaults()

	if videoConn == nil && audioConn == nil && ctrlConn == nil {
		return nil, fmt.Errorf("scrcpy: NewSession: at least one connection must be non-nil")
	}

	s := &Session{}

	// ── video ────────────────────────────────────────────────────────────────
	if videoConn != nil {
		info, err := ReadDeviceInfo(videoConn)
		if err != nil {
			return nil, fmt.Errorf("scrcpy: video handshake (device info): %w", err)
		}
		s.Device = *info

		vi, err := ReadVideoInfo(videoConn)
		if err != nil {
			return nil, fmt.Errorf("scrcpy: video handshake (video info): %w", err)
		}
		s.Video = *vi
		s.Frames = NewFrameDemuxer(videoConn, opts.VideoBufSize)
	}

	// ── audio ────────────────────────────────────────────────────────────────
	if audioConn != nil {
		info, err := ReadDeviceInfo(audioConn)
		if err != nil {
			s.close()
			return nil, fmt.Errorf("scrcpy: audio handshake (device info): %w", err)
		}
		if s.Device.DeviceName == "" {
			s.Device = *info
		}

		ai, err := ReadAudioInfo(audioConn)
		if err != nil {
			s.close()
			return nil, fmt.Errorf("scrcpy: audio handshake (audio info): %w", err)
		}
		s.Audio = *ai
		s.AudioPkts = NewPacketBuffer(audioConn, opts.AudioBufSize)
	}

	// ── control ──────────────────────────────────────────────────────────────
	if ctrlConn != nil {
		info, err := ReadDeviceInfo(ctrlConn)
		if err != nil {
			s.close()
			return nil, fmt.Errorf("scrcpy: control handshake (device info): %w", err)
		}
		if s.Device.DeviceName == "" {
			s.Device = *info
		}
		s.Control = NewControlConn(ctrlConn, opts.CtrlBufSize)
	}

	return s, nil
}

// Close stops all background worker goroutines (FrameDemuxer, PacketBuffer,
// ControlConn).  It does not close the underlying network connections; call
// Close on each net.Conn separately to interrupt in-progress reads.
func (s *Session) Close() {
	s.close()
}

func (s *Session) close() {
	if s.Frames != nil {
		s.Frames.Close()
	}
	if s.AudioPkts != nil {
		s.AudioPkts.Close()
	}
	if s.Control != nil {
		s.Control.Close()
	}
}
