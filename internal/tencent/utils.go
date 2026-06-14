package tencent

func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func strSlicePtr(ss []string) []*string {
	result := make([]*string, len(ss))
	for i, s := range ss {
		result[i] = stringPtr(s)
	}
	return result
}
