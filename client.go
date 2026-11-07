package nacos

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

type configAPI interface {
	GetConfig(param vo.ConfigParam) (string, error)
	ListenConfig(param vo.ConfigParam) error
	CancelListenConfig(param vo.ConfigParam) error
}

// ErrClosed is returned when a closed Client is used to start a watch.
var ErrClosed = errors.New("nacos: client is closed")

// retryDelay is the pause between Download attempts.
const retryDelay = time.Second

// watch is one active ListenConfig registration.
type watch struct {
	param   vo.ConfigParam
	stopped atomic.Bool
	done    chan struct{}
}

// Client downloads and watches Nacos configuration content.
type Client struct {
	cfg Config
	api configAPI

	mu      sync.Mutex
	closed  bool
	watches map[*watch]struct{}
}

// NewClient creates a Nacos config client.
func NewClient(cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	api, err := newConfigAPI(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{cfg: cfg, api: api, watches: make(map[*watch]struct{})}, nil
}

// NewClientWithAPI injects the config API for tests.
func NewClientWithAPI(cfg Config, api configAPI) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if api == nil {
		return nil, errors.New("nacos: api is nil")
	}
	return &Client{cfg: cfg, api: api, watches: make(map[*watch]struct{})}, nil
}

// Download fetches remote config with one retry.
func (c *Client) Download(ctx context.Context) (string, error) {
	if c == nil || c.api == nil {
		return "", errors.New("nacos: client is not initialized")
	}
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		if attempt > 0 {
			if err := sleepContext(ctx, retryDelay); err != nil {
				return "", err
			}
		}
		content, err := c.api.GetConfig(vo.ConfigParam{
			DataId: c.cfg.DataID,
			Group:  c.cfg.Group,
		})
		if err != nil {
			lastErr = err
			continue
		}
		if content == "" {
			lastErr = errors.New("nacos: empty config content")
			continue
		}
		return content, nil
	}
	if lastErr == nil {
		lastErr = errors.New("nacos: download failed")
	}
	return "", lastErr
}

// Watch registers a change listener. onChange receives raw config bytes.
// The listener is cancelled (CancelListenConfig) when ctx is done or Close is
// called; onChange is not invoked after cancellation.
func (c *Client) Watch(ctx context.Context, onChange func(content []byte) error) error {
	if c == nil || c.api == nil {
		return errors.New("nacos: client is not initialized")
	}
	if onChange == nil {
		return errors.New("nacos: onChange is nil")
	}
	if ctx == nil {
		return errors.New("nacos: ctx is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	w := &watch{done: make(chan struct{})}
	w.param = vo.ConfigParam{
		DataId: c.cfg.DataID,
		Group:  c.cfg.Group,
		OnChange: func(namespace, group, dataID, data string) {
			if w.stopped.Load() {
				return
			}
			_ = onChange([]byte(data))
		},
	}

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return ErrClosed
	}
	if c.watches == nil {
		c.watches = make(map[*watch]struct{})
	}
	if err := c.api.ListenConfig(w.param); err != nil {
		c.mu.Unlock()
		return fmt.Errorf("nacos: listen config: %w", err)
	}
	c.watches[w] = struct{}{}
	c.mu.Unlock()

	go func() {
		select {
		case <-ctx.Done():
			_ = c.stopWatch(w)
		case <-w.done:
		}
	}()
	return nil
}

// stopWatch cancels one registration; it is idempotent.
func (c *Client) stopWatch(w *watch) error {
	c.mu.Lock()
	if _, ok := c.watches[w]; !ok {
		c.mu.Unlock()
		return nil
	}
	delete(c.watches, w)
	c.mu.Unlock()

	w.stopped.Store(true)
	close(w.done)
	if err := c.api.CancelListenConfig(vo.ConfigParam{DataId: w.param.DataId, Group: w.param.Group}); err != nil {
		return fmt.Errorf("nacos: cancel listen config: %w", err)
	}
	return nil
}

// Close cancels every active watch via CancelListenConfig. It is safe to call
// more than once; later Watch calls return ErrClosed.
func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	c.closed = true
	active := make([]*watch, 0, len(c.watches))
	for w := range c.watches {
		active = append(active, w)
	}
	c.mu.Unlock()

	var errs []error
	for _, w := range active {
		if err := c.stopWatch(w); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// sleepContext waits for d or until ctx is done.
func sleepContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func newConfigAPI(cfg Config) (config_client.IConfigClient, error) {
	serverConfigs := []constant.ServerConfig{
		*constant.NewServerConfig(
			cfg.Address,
			cfg.Port,
			constant.WithContextPath("/nacos"),
			constant.WithScheme(cfg.Scheme),
		),
	}
	clientConfig := constant.ClientConfig{
		NamespaceId:         cfg.NamespaceID,
		Username:            cfg.Username,
		Password:            cfg.Password,
		TimeoutMs:           uint64(cfg.Timeout.Milliseconds()),
		BeatInterval:        int64(cfg.PollTime.Milliseconds()),
		NotLoadCacheAtStart: true,
		LogLevel:            "info",
	}
	api, err := clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig:  &clientConfig,
		ServerConfigs: serverConfigs,
	})
	if err != nil {
		return nil, fmt.Errorf("nacos: create client: %w", err)
	}
	return api, nil
}
