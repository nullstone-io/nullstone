package nullstone

type Apps struct {
	client *Client
}

func (a Apps) Get(name string) (*Application, error) {
	panic("not implemented")
}
