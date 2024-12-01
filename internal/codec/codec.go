package codec

import codec_model "github.com/sk25469/kv/internal/codec/model"

type ICodec interface {
	Encode(string) (*codec_model.Command, error)
	Decode(*codec_model.Command) ([]byte, error)
}

type CodecLayer struct {
	Command *codec_model.Command
}

func NewCodecLayer() *CodecLayer {
	return &CodecLayer{
		Command: &codec_model.Command{},
	}
}

func (c *CodecLayer) Encode(data string) (*codec_model.Command, error) {
	return c.Command.Encode(data), nil
}

func (c *CodecLayer) Decode(cmd *codec_model.Command) ([]byte, error) {
	return []byte(cmd.Decode()), nil
}
