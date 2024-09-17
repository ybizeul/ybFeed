package feed

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/Appboy/webpush-go"
	"golang.org/x/exp/slog"
)

var FeedConfigErrorCantWrite = errors.New("can't write feed configuration")
var FeedConfigErrorNotFound = errors.New("feed configuration not found")
var FeedConfigErrorInvalid = errors.New("feed configuration invalid")
var FeedConfigErrorPinExpired = errors.New("feed pin expired")
var FeedConfigErrorPinIncorrect = errors.New("feed pin incorrect")
var FeedConfigErrorPinIncorrectLength = errors.New("feed pin length is not 4")

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
		if err := result.migratev1v2(); err != nil {
			return nil, err
		}

	}
	b, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", FeedConfigErrorNotFound, configPath)
		}
	}
	err = json.Unmarshal(b, result)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", FeedConfigErrorInvalid, configPath)
	}
	return result, nil
}

func (config *FeedConfig) Write() error {
	configPath := path.Join(config.feed.Path, "config.json")

	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("%w: %s", FeedConfigErrorCantWrite, configPath)
	}

	e := json.NewEncoder(f)
	e.SetIndent("", "  ")
	err = e.Encode(config)

	if err != nil {
		return fmt.Errorf("%w: %s", FeedConfigErrorCantWrite, configPath)
	}
	return nil
}

func (config *FeedConfig) SetPIN(s string) error {
	if len(s) != 4 {
		return FeedConfigErrorPinIncorrectLength
	}
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
		slog.Warn("PIN expired")
		return FeedConfigErrorPinExpired
	}
	if s != p.PIN {
		return FeedConfigErrorPinIncorrect
	}

	return nil
}
