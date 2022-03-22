package tarus_io

type ChannelFactory = func(inp, oup string) (Factory, error)

type Router interface {
	MakeIOChannel(iop string) (ChannelFactory, error)
}
