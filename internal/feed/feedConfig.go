package feed

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"time"

	"github.com/Appboy/webpush-go"
	"golang.org/x/exp/slog"
)

type FeedConfig struct {
	Secret        string `json:"secret"`
	PIN           *PIN   `json:"pin,omitempty"`
	Subscriptions []webpush.Subscription
	feed          *Feed
}

type PIN struct {
	PIN        string    `json:"pin,omitempty"`
	Expiration time.Time `json:"expiration,omitempty"`
}

func (config *FeedConfig) migratev1v2() error {
	if config.Secret == "" {
		secretPath := path.Join(config.feed.Path, "secret")
		b, err := os.ReadFile(secretPath)
		if err != nil {
			return err
		}
		config.Secret = string(b)
	}

	pinPath := path.Join(config.feed.Path, "pin")
	stat, err := os.Stat(pinPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		pincode, err := os.ReadFile(pinPath)
		if err != nil {
			return err
		}

		pin := &PIN{
			PIN:        string(pincode),
			Expiration: stat.ModTime().Add(time.Minute * 2),
		}
		config.PIN = pin
	}
	err = config.Write()

	if err != nil {
		return err
	}
	os.Remove(path.Join(config.feed.Path, "secret"))
	os.Remove(pinPath)
	return nil
}

func FeedConfigForFeed(f *Feed) (*FeedConfig, error) {
	result := &FeedConfig{feed: f}

	configPath := path.Join(f.Path, "config.json")
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		result.migratev1v2()
	}
	b, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &FeedError{
				Code:    500,
				Message: "Configuration does not exists",
			}
		}
	}
	err = json.Unmarshal(b, result)
	if err != nil {
		return nil, &FeedError{
			Code:    500,
			Message: "Unable to read configuration",
		}
	}
	return result, nil
}

func (config *FeedConfig) Write() error {
	b, err := json.Marshal(config)
	if err != nil {
		return &FeedError{
			Code:    500,
			Message: "Unable to create configuration",
		}
	}

	err = os.WriteFile(path.Join(config.feed.Path, "config.json"), b, 0600)
	if err != nil {
		return &FeedError{
			Code:    500,
			Message: "Unable to write configuration",
		}
	}
	return nil
}

func (config *FeedConfig) SetPIN(s string) error {
	pin := &PIN{
		PIN:        s,
		Expiration: time.Now().Add(2 * time.Minute),
	}
	config.PIN = pin
	err := config.Write()
	if err != nil {
		return err
	}
	return nil
}

func (config *FeedConfig) AddSubscription(s webpush.Subscription) error {
	for _, t := range config.Subscriptions {
		if s.Endpoint == t.Endpoint && s.Keys.Auth == t.Keys.Auth && s.Keys.P256dh == t.Keys.P256dh {
			return nil
		}
	}
	config.Subscriptions = append(config.Subscriptions, s)
	err := config.Write()
	if err != nil {
		return err
	}
	return nil
}

func (config *FeedConfig) DeleteSubscription(s webpush.Subscription) error {
	keepSubscriptions := []webpush.Subscription{}

	for _, t := range config.Subscriptions {
		if s.Endpoint == t.Endpoint && s.Keys.Auth == t.Keys.Auth && s.Keys.P256dh == t.Keys.P256dh {
			continue
		}
		keepSubscriptions = append(keepSubscriptions, t)
	}
	config.Subscriptions = keepSubscriptions
	err := config.Write()
	if err != nil {
		return err
	}
	return nil
}

func (p *PIN) IsValid(s string) error {
	if p.Expiration.Before(time.Now()) {
		code := 401
		slog.Warn("PIN expired", slog.Int("return", code))
		return &FeedError{
			Code:    code,
			Message: "Pin Expired",
		}
	}
	if s != p.PIN {
		code := 401
		return &FeedError{
			Code:    code,
			Message: "PIN Incorrect",
		}
	}

	return nil
}
