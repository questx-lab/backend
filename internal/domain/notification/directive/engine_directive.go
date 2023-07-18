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
