package utils

// 默认值则返回 nil
func Ptr[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}

// 如果指针为 nil，则返回类型默认值
func PtrValue[T comparable](v *T) T {
	var zero T
	if v == nil {
		return zero
	}
	return *v
}
