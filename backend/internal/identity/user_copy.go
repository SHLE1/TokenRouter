package identity

// CopyUser 浅复制用户结构体。
// 关联切片、映射和指针共享引用。跨请求缓存通过各自的入口深复制。
func CopyUser(user *User) *User {
	if user == nil {
		return nil
	}
	copy := *user
	return &copy
}
