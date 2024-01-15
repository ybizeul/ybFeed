package feed

import (
	"fmt"
	"io"

	"github.com/Appboy/webpush-go"
	"github.com/ybizeul/ybfeed/pkg/yblog"
	"golang.org/x/exp/slog"
)

var pnL = yblog.NewYBLogger("push", []string{"DEBUG", "DEBUG_NOTIFICATIONS"})

func (f *Feed) sendPushNotification() error {
	// Send push notifications
	if f.NotificationSettings == nil {
		pnL.Logger.Debug("Feed has no notifications settings")
		return nil
	}
	for _, subscription := range f.Config.Subscriptions {
		pnL.Logger.Debug("Sending push notification", slog.String("endpoint", subscription.Endpoint))
		resp, _ := webpush.SendNotification([]byte(fmt.Sprintf("New item posted to feed %s", f.Name())), &subscription, &webpush.Options{
			Subscriber:      "ybfeed@tynsoe.org", // Do not include "mailto:"
			VAPIDPublicKey:  f.NotificationSettings.VAPIDPublicKey,
			VAPIDPrivateKey: f.NotificationSettings.VAPIDPrivateKey,
			TTL:             30,
		})
		if pnL.Level() == slog.LevelDebug {
			b, _ := io.ReadAll(resp.Body)
			pnL.Logger.Debug("Response", slog.String("resp", string(b)), slog.String("status", resp.Status))
		}
		defer resp.Body.Close()
	}
	return nil
}
