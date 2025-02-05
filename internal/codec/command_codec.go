package codec

import (
	codec_model "github.com/sk25469/kv/internal/codec/model"
)

type ICommandCodec interface {
	Encode(rawCommand string) *codec_model.Command
	Decode() []byte
}

type CommandCodecLayer struct {
	CommandModel *codec_model.Command
}

func NewCommandCodecLayer() *CommandCodecLayer {
	return &CommandCodecLayer{
		CommandModel: &codec_model.Command{},
	}
}

func (c *CommandCodecLayer) Encode(rawCommand string) *codec_model.Command {
	return c.CommandModel.Encode(rawCommand)
}

func (c *CommandCodecLayer) Decode() []byte {
	return c.CommandModel.Decode()
}
