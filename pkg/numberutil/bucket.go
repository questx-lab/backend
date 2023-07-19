package numberutil

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

const BucketDuration int64 = 1000 * 60 * 60 * 24 * 10 // 10 days

func BucketFrom(id int64) int64 {
	if id != 0 {
		sfID := snowflake.ParseInt64(id)
		return sfID.Time() / BucketDuration
	}

	return time.Now().UnixMilli() / BucketDuration
}
