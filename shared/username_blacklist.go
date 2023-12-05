package shared

func IsBlacklisted(name string) bool {
	return GetCache().IsUsernameBlacklisted(name)
}
