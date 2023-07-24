package numberutil

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

const BucketDuration int64 = 1000 * 60 * 60 * 24 * 10 // 10 days

func BucketFrom(id int64) int64 {
	if id != 0 {
		return ((id >> 22) + snowflake.Epoch) / BucketDuration
	}

	return time.Now().UnixMilli() / BucketDuration
}
