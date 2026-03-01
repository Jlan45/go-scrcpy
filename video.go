package scrcpy

import (
	"io"
	"sync"
)

// VideoFrame is a self-contained encoded video frame ready for a hardware or
// software decoder.
//
// When Config is non-nil the codec configuration changed since the previous
// frame (new SPS/PPS for H.264, new VPS/SPS/PPS for H.265, new sequence
// headers for AV1).  The decoder must be reinitialized with Config before
// feeding Data to it.
type VideoFrame struct {
	// PTS is the presentation timestamp in microseconds.
	PTS int64
	// IsKeyframe is true for random-access (I-frame) packets.
	IsKeyframe bool
	// Config is non-nil when the codec configuration changed since the last
	// delivered frame (e.g. on stream start or screen rotation).
	// Feed it to your decoder before decoding Data.
	Config []byte
	// Data is the encoded frame payload (H.264/H.265 NAL units or AV1 OBUs).
	Data []byte
}

// FrameDemuxer reads from a scrcpy video socket, groups codec-configuration
// packets with the next data frame, and delivers [VideoFrame] values through
// a bounded channel for asynchronous consumption.
//
// Typical usage:
//
//	d := scrcpy.NewFrameDemuxer(videoConn, 30)
//	defer d.Close()
//
//	for frame := range d.Frames() {
//	    if frame.Config != nil {
//	        decoder.Init(frame.Config)
//	    }
//	    decoder.Decode(frame.PTS, frame.Data)
//	}
//	if err := d.Err(); err != nil {
//	    log.Println("demuxer stopped:", err)
//	}
type FrameDemuxer struct {
	out  chan *VideoFrame
	stop chan struct{}
	once sync.Once
	err  error
}

// NewFrameDemuxer starts a background goroutine that reads video packets from
// r and queues [VideoFrame] values.
//
// bufSize is the frame-buffer capacity (number of frames).  The producer
// goroutine blocks when the buffer is full, providing natural backpressure to
// the network read loop.  A bufSize of 0 creates an unbuffered channel where
// every frame must be consumed before the next one is read.
//
// Call [FrameDemuxer.Close] when done.  To interrupt an in-progress read,
// close the underlying net.Conn; Close alone is not sufficient because
// ReadPacket blocks on the network.
func NewFrameDemuxer(r io.Reader, bufSize int) *FrameDemuxer {
	d := &FrameDemuxer{
		out:  make(chan *VideoFrame, bufSize),
		stop: make(chan struct{}),
	}
	go d.run(r)
	return d
}

func (d *FrameDemuxer) run(r io.Reader) {
	defer close(d.out)

	var (
		config        []byte // most recently seen codec config payload
		configPending bool   // true until config is delivered with a data frame
	)

	for {
		// Check for shutdown before blocking on the network read.
		select {
		case <-d.stop:
			return
		default:
		}

		pkt, err := ReadPacket(r)
		if err != nil {
			d.err = err
			return
		}

		if pkt.IsConfig {
			// Codec configuration packet (H.264 SPS/PPS, H.265 VPS/SPS/PPS,
			// AV1 sequence header, …).  Hold it until the next data packet so
			// we can deliver them together as a single VideoFrame.
			config = pkt.Data
			configPending = true
			continue
		}

		frame := &VideoFrame{
			PTS:        pkt.PTS,
			IsKeyframe: pkt.IsKeyframe,
			Data:       pkt.Data,
		}
		if configPending {
			frame.Config = config
			configPending = false
		}

		select {
		case d.out <- frame:
		case <-d.stop:
			return
		}
	}
}

// Frames returns the read-only channel of assembled video frames.
//
// The channel is closed when the underlying stream ends or an unrecoverable
// read error occurs.  Use [FrameDemuxer.Err] after the channel closes to
// inspect the error.
func (d *FrameDemuxer) Frames() <-chan *VideoFrame {
	return d.out
}

// Err returns the error that caused the demuxer to stop.
//
// It is only meaningful after the channel returned by [FrameDemuxer.Frames]
// has been closed (i.e. after the range loop exits or a receive returns the
// zero value).  Returns nil when the demuxer stopped cleanly via Close.
func (d *FrameDemuxer) Err() error {
	return d.err
}

// Len returns the number of frames currently queued in the buffer.
func (d *FrameDemuxer) Len() int {
	return len(d.out)
}

// Cap returns the buffer capacity passed to NewFrameDemuxer.
func (d *FrameDemuxer) Cap() int {
	return cap(d.out)
}

// Close signals the demuxer goroutine to stop after the current read
// completes.  It does not close the underlying reader; to interrupt an
// in-progress ReadPacket call, close the net.Conn directly.
func (d *FrameDemuxer) Close() {
	d.once.Do(func() { close(d.stop) })
}
