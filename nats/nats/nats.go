package nats

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"github.com/siangyeh8818/commonTools/errors"

	traceRequestID "github.com/siangyeh8818/commonTools/trace/requestID"
	traceTime "github.com/siangyeh8818/commonTools/trace/time"
)

var (
	waitGroup sync.WaitGroup
)

type Topic struct {
	Topic string
}

// Client 訂閱方

type Client struct {
	natsConn *nats.Conn
	cfg      *Config

	Channels []Channel
}

// NewClient ...
func NewClient(cfg *Config) (*Client, error) {
	nc, err := NewNatsConn(cfg)
	if err != nil {
		return nil, err
	}

	client := Client{
		natsConn: nc,
		cfg:      cfg,
	}

	//client.natsConn.SetReconnectHandler(client.reconnect)

	return &client, nil
}

// NewNatsConn ...
func NewNatsConn(c *Config) (*nats.Conn, error) {

	waitGroup.Add(1)

	var natsConn *nats.Conn

	natsConn, err := nats.Connect(strings.Join(c.Address, ","),
		nats.Name(c.Name),
		//nats.MaxReconnects(-1),
		//nats.ReconnectWait(5*time.Second),
		nats.UserInfo(c.Account, c.Password),
		nats.ReconnectHandler(func(conn *nats.Conn) {
			log.Info().Msg(" nats reconnect event triggered")
		}),
		nats.DisconnectErrHandler(func(conn *nats.Conn, cErr error) {
			if cErr != nil {
				log.Error().Msgf(" nats disconnectErr event triggered cErr=%v", cErr)
				return
			}
			log.Info().Msg(" nats disconnectErr event triggered")
		}),
		nats.ClosedHandler(func(conn *nats.Conn) {
			log.Info().Msg(" nats closed event triggered")
			waitGroup.Done()
		}),
		nats.ErrorHandler(func(conn *nats.Conn, sub *nats.Subscription, sErr error) {
			log.Error().Msgf("XLB: nats error event triggered sub=%s sErr=%v", sub.Subject, sErr)
		}),
	)

	if err != nil {
		log.Error().Msgf("connect to nats server error %s", err.Error())
	}

	if c.ClientID == "" {
		c.ClientID = uuid.New().String()
		if c.AppID != "" {
			c.ClientID = fmt.Sprintf("%s_%s", c.AppID, c.ClientID)
		}
	}
	return natsConn, nil
}

// 連線流放

func (c *Client) Drain() error {
	err := c.natsConn.Drain()
	if err != nil {
		return err
	}
	waitGroup.Wait()

	return nil
}


// Config for stan client
type Config struct {
	Name        string   `mapstructure:"name" yaml:"name"`
	Address     []string `mapstructure:"address" yaml:"address"`
	ClientID    string   `mapstructure:"client_id" yaml:"client_id"`
	DurableName string   `mapstructure:"durable_name" yaml:"durable_name"` // 如果有設定所有channel 會使用同一個 durableName. 提供 stan 去紀錄, 上一次讀到哪裡
	AppID       string   `mapstructure:"app_id" yaml:"app_id"`
	Account     string   `mapstructure:"account" yaml:"account"`
	Password    string   `mapstructure:"password" yaml:"password"`
}



// Pub 推送
func (c *Client) Pub(ctx context.Context, subject string, header map[string][]string, data []byte) error {

	var h = nats.Header{
		"request_id": []string{traceRequestID.FromContext(ctx)},
		"time":       []string{strconv.FormatInt(traceTime.GetFromContext(ctx), 10)},
	}
	for k, v := range header {
		h[k] = v
	}
	msg := &nats.Msg{
		Subject: subject,
		Header:  h,
		Data:    data,
	}
	if err := c.natsConn.PublishMsg(msg); err != nil {
		return errors.Wrapf(errors.ErrInternal, "fail to publish to nats, err: %s", err.Error())
	}

	return nil
}

// Channel ...
type Channel struct {
	ChannelName string
	GroupName   string
	Handler     func(ctx context.Context, msg *nats.Msg) error
}

//  Sub...
func (c *Client) Sub(topic string, handler func(ctx context.Context, msg *nats.Msg) error) error {

	_, err := c.natsConn.Subscribe(topic, func(msg *nats.Msg) {
		defer recoverLog()
		internalCtx, _ := context.WithTimeout(context.Background(), 30*time.Second)
		var requestID string
		var t int64
		if len(msg.Header["request_id"]) > 0 {
			requestID = msg.Header["request_id"][0]
		}
		if len(msg.Header["time"]) > 0 {
			t, _ = strconv.ParseInt(msg.Header["time"][0], 10, 64)
		}
		internalCtx = traceTime.ContextWithTime(traceRequestID.ContextWithXRequestID(internalCtx, requestID), t)
		logger := log.With().Int64("time", t).Str("request_id", requestID).Str("endpoint", topic).Logger()
		logger.Info().Msgf("%+v", msg)
		internalCtx = logger.WithContext(internalCtx)

		err := handler(internalCtx, msg)
		if err != nil {
			logger.Error().Msgf("channel: %s, error: %+v", topic, err)
		}
	})
	if err != nil {
		return err
	}

	return nil
}

// RegisterChannel ...
func (c *Client) RegisterChannel(channels []Channel) error {
	c.Channels = channels
	for i := range channels {
		log.Info().Msgf("Register channel: %s", channels[i].ChannelName)
		name, group, handler := channels[i].ChannelName, channels[i].GroupName, channels[i].Handler

		_, err := c.natsConn.QueueSubscribe(name, group, func(msg *nats.Msg) {
			defer recoverLog()
			internalCtx, _ := context.WithTimeout(context.Background(), 30*time.Second)
			var requestID string
			var t int64
			if len(msg.Header["request_id"]) > 0 {
				requestID = msg.Header["request_id"][0]
			}
			if len(msg.Header["time"]) > 0 {
				t, _ = strconv.ParseInt(msg.Header["time"][0], 10, 64)
			}
			internalCtx = traceTime.ContextWithTime(traceRequestID.ContextWithXRequestID(internalCtx, requestID), t)
			logger := log.With().Int64("time", t).Str("request_id", requestID).Str("endpoint", name).Logger()
			logger.Info().Msgf("%+v", msg)
			internalCtx = logger.WithContext(internalCtx)

			err := handler(internalCtx, msg)
			if err != nil {
				logger.Error().Msgf("channel: %s, error: %+v", name, err)
			}

		})
		if err != nil {
			return err
		}
	}
	return nil
}

func recoverLog() {
	if r := recover(); r != nil {
		var msg string
		for i := 2; ; i++ {
			_, file, line, ok := runtime.Caller(i)
			if !ok {
				break
			}
			msg += fmt.Sprintf("%s:%d\n", file, line)
		}
		log.Error().Msgf("%s\n↧↧↧↧↧↧ PANIC ↧↧↧↧↧↧\n%s↥↥↥↥↥↥ PANIC ↥↥↥↥↥↥", r, msg)
	}
}
