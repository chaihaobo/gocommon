package encoder

import (
	"github.com/fatih/color"
	"github.com/hokaccha/go-prettyjson"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

type jsonColorEncoder struct {
	zapcore.Encoder
	formatter *prettyjson.Formatter
}

func NewJSONColorEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	formatter := prettyjson.NewFormatter()
	formatter.Newline = ""
	formatter.Indent = 0
	color.NoColor = false
	return &jsonColorEncoder{
		zapcore.NewJSONEncoder(cfg),
		formatter,
	}
}

func (j jsonColorEncoder) EncodeEntry(entry zapcore.Entry, field []zapcore.Field) (*buffer.Buffer, error) {
	encodeEntry, err := j.Encoder.EncodeEntry(entry, field)
	if err != nil {
		return nil, err
	}
	colorfulEncodedEntry, err := j.formatter.Format(encodeEntry.Bytes())
	if err != nil {
		return nil, err
	}

	encodeEntry.Reset()
	encodeEntry.AppendBytes(colorfulEncodedEntry)
	encodeEntry.AppendBytes([]byte("\n"))
	return encodeEntry, nil
}
