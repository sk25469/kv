package codec

import (
	codec_model "github.com/sk25469/kv/internal/codec/model"
	network_model "github.com/sk25469/kv/internal/network/model"
)

type ICommCodec interface {
	Encode(codec_model.CommandType, *network_model.NodeConfig, *network_model.NodeConfig) *codec_model.CommunicationModel
	Decode() (string, error)
}

type CommCodecLayer struct {
	CommunicationModel *codec_model.CommunicationModel
}

func NewCommCodecLayer() *CommCodecLayer {
	return &CommCodecLayer{
		CommunicationModel: &codec_model.CommunicationModel{},
	}
}

func (c *CommCodecLayer) Encode(cmdType codec_model.CommandType, sendTo, sentFrom *network_model.NodeConfig) (*codec_model.CommunicationModel, error) {
	return codec_model.NewCommunicationModel(cmdType, sendTo, sentFrom), nil
}

func (c *CommCodecLayer) Decode() (string, error) {
	res, err := c.CommunicationModel.Decode()
	if err != nil {
		return "", err
	}
	return string(res), nil
}
