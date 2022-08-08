package subscribers

import (
	"github.com/larisgo/laravel-echo-server/types"
)

type Broadcast func(string, *types.Data)

type Subscriber interface {
	// Subscribe to incoming events.
	Subscribe(Broadcast)

	// Unsubscribe from events to broadcast.
	UnSubscribe()
}
