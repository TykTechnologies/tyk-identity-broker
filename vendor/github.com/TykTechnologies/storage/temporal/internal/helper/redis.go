package helper

import (
	"strconv"

	"github.com/TykTechnologies/storage/temporal/model"
)

// GetRedisAddrs returns a list of redis addresses from the types.RedisOptions.
func GetRedisAddrs(opts *model.RedisOptions) (addrs []string) {
	if len(opts.Addrs) != 0 {
		addrs = opts.Addrs
	} else {
		for h, p := range opts.Hosts {
			addr := h + ":" + p
			addrs = append(addrs, addr)
		}
	}

	if len(addrs) == 0 && opts.Port != 0 {
		addr := opts.Host + ":" + strconv.Itoa(opts.Port)
		addrs = append(addrs, addr)
	}

	return addrs
}
