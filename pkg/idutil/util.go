package idutil

import (
	"github.com/bwmarrin/snowflake"
	"github.com/questx-lab/backend/internal/common"
)

func GetBucketByID(id int64) int64 {
	sID := snowflake.ParseInt64(id)
	return sID.Time() / common.BucketDuration.Milliseconds()
}
