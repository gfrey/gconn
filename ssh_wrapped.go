package gconn

type wrappedSSHClient struct {
	Client
	wrapper string
}

// NewWrappedClient will use the given client and wrap all sessions with a
// command. This can be used to wrap commands in a `sudo` session for example.
func NewWrappedClient(client Client, wrapper string) Client {
	return &wrappedSSHClient{Client: client, wrapper: wrapper}
}

func (wc *wrappedSSHClient) NewSession(cmd string, args ...string) (Session, error) {
	return wc.Client.NewSession(wc.wrapper, append([]string{cmd}, args...)...)
}
