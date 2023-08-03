package common

import (
	"fmt"
	"strings"
)

func RedisKeyUserStatus(userID string) string {
	return fmt.Sprintf("userstatus:%s", userID)
}

func FromRedisKeyUserStatus(key string) string {
	return strings.Split(key, ":")[1]
}

func RedisKeyCommunityOnline(communityID string) string {
	return fmt.Sprintf("communityonline:%s", communityID)
}
