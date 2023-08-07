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

func RedisKeyFollower(communityID string) string {
	return fmt.Sprintf("followers:%s", communityID)
}

func RedisValueFollower(username, userID string) string {
	return fmt.Sprintf("%s***%s", username, userID)
}

func FromRedisValueFollower(value string) (string, string) {
	parts := strings.Split(value, "***")
	return parts[0], parts[1]
}
