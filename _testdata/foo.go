package testdata

type Server struct{}

type (
	foobarReqType struct{
		A int
		B string
	}

	foobarRespType struct {
		X string
		Y int
	}
)

func (Server) FooBar(req foobarReqType) foobarRespType {
	return foobarRespType{
		X: req.B,
		Y: req.A,
	}
}
