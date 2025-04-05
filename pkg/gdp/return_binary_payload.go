package gdp

import "github.com/randomizedcoder/gdp/pkg/gdp_config"

func (g *GDP) ReturnBinaryPayload(proto any, mc *gdp_config.MarshalConfig) (binaryProto *[]byte) {
	return g.marshal(proto, mc)
}
