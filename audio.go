package scrcpy

import (
	"io"
	"sync"
)

// PacketBuffer reads packets from an audio stream socket and delivers them
// through a buffered channel for asynchronous consumption.
//
// Unlike [FrameDemuxer], all packets are delivered as-is — including codec
// configuration packets (Packet.IsConfig == true).  Callers should inspect
// IsConfig to detect codec initialisation packets (e.g. Opus ID header).
//
// Typical usage:
//
//	for pkt := range buf.Packets() {
//	    if pkt.IsConfig {
//	        audioDecoder.Init(pkt.Data)
//	        continue
//	    }
//	    audioDecoder.Decode(pkt.PTS, pkt.Data)
//	}
type PacketBuffer struct {
	out  chan *Packet
	stop chan struct{}
	once sync.Once
	err  error
}

// NewPacketBuffer starts a background goroutine that reads packets from r and
// queues them.  bufSize is the channel capacity; the producer blocks when the
// buffer is full.
func NewPacketBuffer(r io.Reader, bufSize int) *PacketBuffer {
	b := &PacketBuffer{
		out:  make(chan *Packet, bufSize),
		stop: make(chan struct{}),
	}
	go b.run(r)
	return b
}

func (b *PacketBuffer) run(r io.Reader) {
	defer close(b.out)
	for {
		select {
		case <-b.stop:
			return
		default:
		}

		pkt, err := ReadPacket(r)
		if err != nil {
			b.err = err
			return
		}

		select {
		case b.out <- pkt:
		case <-b.stop:
			return
		}
	}
}

// Packets returns the read-only channel of audio packets.
// The channel is closed when the stream ends or an error occurs.
func (b *PacketBuffer) Packets() <-chan *Packet {
	return b.out
}

// Err returns the error that caused the buffer to stop.
// Only meaningful after the Packets() channel has been closed.
func (b *PacketBuffer) Err() error {
	return b.err
}

// Len returns the number of packets currently queued.
func (b *PacketBuffer) Len() int { return len(b.out) }

// Cap returns the buffer capacity passed to NewPacketBuffer.
func (b *PacketBuffer) Cap() int { return cap(b.out) }

// Close signals the background goroutine to stop.
// Does not close the underlying reader.
func (b *PacketBuffer) Close() {
	b.once.Do(func() { close(b.stop) })
}
