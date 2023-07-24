package entity

type ChatChannelBucket struct {
	ChannelID int64
	Bucket    int64
	Quantity  int64
}

func (b *ChatChannelBucket) TableName() string {
	return "channel_buckets"
}
