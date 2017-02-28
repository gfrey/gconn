package gconn

type wrappedSSHClient struct {
	Client
	wrapper string
}

func NewWrappedClient(client Client, wrapper string) Client {
	return &wrappedSSHClient{Client: client, wrapper: wrapper}
}

func (wc *wrappedSSHClient) NewSession(cmd string, args ...string) (Session, error) {
	return wc.Client.NewSession(wc.wrapper, append([]string{cmd}, args...)...)
}
