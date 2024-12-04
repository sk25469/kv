package codec

import (
	"errors"
	"strings"

	codec_model "github.com/sk25469/kv/internal/codec/model"
)

type ICodec interface {
	Encode(data string, sendTo, sentFrom interface{}) (interface{}, error)
	Decode(data interface{}) ([]byte, error)
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
	data = data[:len(data)-1]
	if strings.Contains(data, "COMM:") {
		// remove the last character from data
		commModel, err := codec_model.CommunicationModelToJSON([]byte(data))
		if err != nil {
			return nil, err
		}
		return c.commCodecLayerService.Encode(commModel.Command, commModel.SendTo, commModel.SentFrom)
	}
	return c.commandCodecLayerService.Encode(data), nil
}

func (c *CodecLayerService) Decode(data interface{}) ([]byte, error) {
	switch v := data.(type) {
	case *codec_model.Command:
		return v.Decode(), nil
	case *codec_model.CommunicationModel:
		res, err := v.Decode()
		return res, err
	default:
		return []byte{}, errors.New("unknown message type")
	}
}
