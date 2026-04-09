package twsp

import (
	"fmt"
	"time"

	natslib "github.com/nats-io/nats.go"

	"github.com/include2md/eventsdk/sdk"
	transportnats "github.com/include2md/eventsdk/sdk/internal/transport/nats"
)

type Options struct {
	NATSURL  string
	Timeout  time.Duration
	Username string
	Password string
	Token    string
}

type Client struct {
	*sdk.SDKClient
	conn *natslib.Conn
	js   natslib.JetStreamContext
}

func NewClient(opts ...Options) (*Client, error) {
	if len(opts) > 1 {
		return nil, fmt.Errorf("at most one options argument is allowed")
	}

	var resolved Options
	if len(opts) == 1 {
		resolved = opts[0]
	}

	nc, js, timeout, err := connect(resolved)
	if err != nil {
		return nil, err
	}

	tr, err := transportnats.New(nc, js)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("create nats transport: %w", err)
	}

	return &Client{
		SDKClient: sdk.NewClient(tr, timeout),
		conn:      nc,
		js:        js,
	}, nil
}

func (c *Client) Close() {
	if c != nil && c.conn != nil {
		c.conn.Close()
	}
}

func (c *Client) EnsureStream(streamName string, subjects ...string) error {
	if c == nil || c.js == nil {
		return fmt.Errorf("client is not initialized")
	}
	if streamName == "" {
		return fmt.Errorf("stream name is required")
	}

	info, err := c.js.StreamInfo(streamName)
	if err == nil {
		cfg := info.Config
		existing := make(map[string]struct{}, len(cfg.Subjects))
		for _, s := range cfg.Subjects {
			existing[s] = struct{}{}
		}
		changed := false
		for _, s := range subjects {
			if s == "" {
				continue
			}
			if _, ok := existing[s]; !ok {
				cfg.Subjects = append(cfg.Subjects, s)
				changed = true
			}
		}
		if !changed {
			return nil
		}
		_, err = c.js.UpdateStream(&cfg)
		return err
	}

	filtered := make([]string, 0, len(subjects))
	for _, s := range subjects {
		if s != "" {
			filtered = append(filtered, s)
		}
	}
	_, err = c.js.AddStream(&natslib.StreamConfig{
		Name:     streamName,
		Subjects: filtered,
	})
	return err
}

func connect(opts Options) (*natslib.Conn, natslib.JetStreamContext, time.Duration, error) {
	natsURL := opts.NATSURL
	if natsURL == "" {
		natsURL = "nats://127.0.0.1:4222"
	}
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	nc, err := natslib.Connect(natsURL, natsConnectOptions(opts)...)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("connect nats: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, nil, 0, fmt.Errorf("create jetstream: %w", err)
	}

	return nc, js, timeout, nil
}

func natsConnectOptions(opts Options) []natslib.Option {
	if opts.Token != "" {
		return []natslib.Option{natslib.Token(opts.Token)}
	}
	if opts.Username != "" || opts.Password != "" {
		return []natslib.Option{natslib.UserInfo(opts.Username, opts.Password)}
	}
	return nil
}
