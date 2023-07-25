package directive

// REGISTER COMMUNITY
type EngineRegisterCommunityDirective struct {
	CommunityID string `json:"community_id"`
}

func NewRegisterCommunityDirective(communityID string) *ClientDirective {
	return &ClientDirective{
		Op:   EngineRegisterCommunityDirectiveOp,
		Data: EngineRegisterCommunityDirective{CommunityID: communityID},
	}
}

// UNREGISTER COMMUNITY
type EngineUnregisterCommunityDirective struct {
	CommunityID string `json:"community_id"`
}

func NewUnregisterCommunityDirective(communityID string) *ClientDirective {
	return &ClientDirective{
		Op:   EngineUnregisterCommunityDirectiveOp,
		Data: EngineUnregisterCommunityDirective{CommunityID: communityID},
	}
}

// REGISTER USER
type EngineRegisterUserDirective struct {
	UserID string `json:"user_id"`
}

func NewRegisterUserDirective(userID string) *ClientDirective {
	return &ClientDirective{
		Op:   EngineRegisterUserDirectiveOp,
		Data: EngineRegisterUserDirective{UserID: userID},
	}
}

// UNREGISTER USER
type EngineUnregisterUserDirective struct {
	UserID string `json:"user_id"`
}

func NewUnregisterUserDirective(userID string) *ClientDirective {
	return &ClientDirective{
		Op:   EngineUnregisterUserDirectiveOp,
		Data: EngineUnregisterUserDirective{UserID: userID},
	}
}
