package cbor

import (
	"encoding/binary"
	"errors"
	"io"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// Maximum message size.
const (
	maxMessageSize = 104857600 // 100 MiB

	MetricCodecSize     = "oasis_codec_size"
	MetricCodecSizeHelp = "CBOR codec message size (bytes)."
)

var (
	errMessageTooLarge  = errors.New("codec: message too large")
	errMessageMalformed = errors.New("codec: message is malformed")

	codecValueSize = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: MetricCodecSize,
			Help: MetricCodecSizeHelp,
		},
		[]string{"call"},
	)

	codecCollectors = []prometheus.Collector{
		codecValueSize,
	}

	labelRead  = prometheus.Labels{"call": "read"}
	labelWrite = prometheus.Labels{"call": "write"}

	metricsOnce sync.Once
)

// MessageReader is a reader wrapper that decodes CBOR-encoded Message structures.
type MessageReader struct {
	reader io.Reader
}

// Read deserializes a single CBOR-encoded Message from the underlying reader.
func (c *MessageReader) Read(msg interface{}) error {
	// Read 32-bit length prefix.
	rawLength := make([]byte, 4)
	if _, err := io.ReadAtLeast(c.reader, rawLength, 4); err != nil {
		return err
	}

	length := binary.BigEndian.Uint32(rawLength)
	codecValueSize.With(labelRead).Observe(float64(length))
	if length > maxMessageSize {
		return errMessageTooLarge
	}

	// Decode message bytes.
	r := io.LimitReader(c.reader, int64(length))
	dec := NewDecoder(r)
	if err := dec.Decode(msg); err != nil {
		return err
	}
	if r.(*io.LimitedReader).N > 0 {
		return errMessageMalformed
	}

	return nil
}

// MessageWriter is a writer wrapper that encodes Messages structures to CBOR.
type MessageWriter struct {
	writer io.Writer
}

// Write serializes a single Message to CBOR and writes it to the underlying writer.
func (c *MessageWriter) Write(msg interface{}) error {
	// Encode into CBOR.
	data := Marshal(msg)
	length := len(data)
	codecValueSize.With(labelWrite).Observe(float64(length))
	if length > maxMessageSize {
		return errMessageTooLarge
	}

	// Write 32-bit length prefix and encoded data.
	rawLength := make([]byte, 4)
	binary.BigEndian.PutUint32(rawLength, uint32(length))
	if _, err := c.writer.Write(rawLength); err != nil {
		return err
	}
	if _, err := c.writer.Write(data); err != nil {
		return err
	}

	return nil
}

// MessageCodec is a length-prefixed Message encoder/decoder.
type MessageCodec struct {
	MessageReader
	MessageWriter
}

// NewMessageCodec constructs a new Message encoder/decoder.
func NewMessageCodec(rw io.ReadWriter) *MessageCodec {
	metricsOnce.Do(func() {
		prometheus.MustRegister(codecCollectors...)
	})

	return &MessageCodec{
		MessageReader: MessageReader{reader: rw},
		MessageWriter: MessageWriter{writer: rw},
	}
}
