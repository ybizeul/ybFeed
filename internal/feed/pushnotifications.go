package feed

import (
	"fmt"
	"os"

	"github.com/Appboy/webpush-go"
	"golang.org/x/exp/slog"
)

var pnLogLevel = new(slog.LevelVar)
var pnLogger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: pnLogLevel})).WithGroup("pushNotification")

func init() {
	if os.Getenv("DEBUG") != "" || os.Getenv("DEBUG_NOTIFICATIONS") != "" {
		pnLogLevel.Set(slog.LevelDebug)
	}
}

func (f *Feed) sendPushNotification() error {
	// Send push notifications
	if f.NotificationSettings == nil {
		pnLogger.Debug("Feed has no notifications settings")
		return nil
	}
	for _, subscription := range f.Config.Subscriptions {
		pnLogger.Debug("Sending push notification", slog.String("endpoint", subscription.Endpoint))
		resp, _ := webpush.SendNotification([]byte(fmt.Sprintf("New item posted to feed %s", f.Name())), &subscription, &webpush.Options{
			Subscriber:      "example@example.com", // Do not include "mailto:"
			VAPIDPublicKey:  f.NotificationSettings.VAPIDPublicKey,
			VAPIDPrivateKey: f.NotificationSettings.VAPIDPrivateKey,
			TTL:             30,
		})
		pnLogger.Debug("Response", slog.Any("resp", resp))
		defer resp.Body.Close()
	}
	return nil
}
