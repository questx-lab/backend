package numberutil

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

const BucketDuration int64 = 1000 * 60 * 60 * 24 * 10 // 10 days

func BucketFrom(id int64) int64 {
	if id != 0 {
		sfid := snowflake.ParseInt64(id)
		return sfid.Time() / BucketDuration
	}

	return time.Now().UnixMilli() / BucketDuration
}
