package codec

import (
	"errors"
	"strings"

	codec_model "github.com/sk25469/kv/internal/codec/model"
	network "github.com/sk25469/kv/internal/network/model"
)

type ICodec interface {
	Encode(data string, sendTo, sentFrom interface{}) (interface{}, error)
	Decode(data interface{}) (interface{}, error)
}

type CodecLayerParams struct {
	CommandCodecLayerService ICommandCodec
	CommCodecLayerService    ICommCodec
}

type CodecLayerService struct {
	commandCodecLayerService *CommandCodecLayer
	commCodecLayerService    *CommCodecLayer
}

func NewCodecLayerService() *CodecLayerService {
	return &CodecLayerService{
		commandCodecLayerService: NewCommandCodecLayer(),
		commCodecLayerService:    NewCommCodecLayer(),
	}
}

func (c *CodecLayerService) Encode(data string, sendTo, sentFrom interface{}) (interface{}, error) {
	if strings.HasPrefix(data, "COMM:") {
		return c.commCodecLayerService.Encode(codec_model.CommandType(data), sendTo.(*network.NodeConfig), sentFrom.(*network.NodeConfig))
	}
	return c.commandCodecLayerService.Encode(data), nil
}

func (c *CodecLayerService) Decode(data interface{}) (string, error) {
	switch v := data.(type) {
	case *codec_model.Command:
		return v.Decode(), nil
	case *codec_model.CommunicationModel:
		res, err := v.Decode()
		return string(res), err
	default:
		return "", errors.New("unknown message type")
	}
}
