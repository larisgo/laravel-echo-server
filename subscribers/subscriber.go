package subscribers

import (
	"github.com/larisgo/laravel-echo-server/types"
)

type Broadcast func(string, types.Data)

type Subscriber interface {
	/**
	 * Subscribe to incoming events.
	 *
	 * @param  {Function} callback
	 * @return {void}
	 */
	Subscribe(Broadcast)
}
