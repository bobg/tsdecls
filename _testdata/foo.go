package testdata

type Server struct{}

type (
	reqType struct{
		A int
		B string
	}

	respType struct {
		X string
		Y int
	}
)

func (Server) FooBar(req reqType) respType {
	return respType{
		X: req.B,
		Y: req.A,
	}
}
