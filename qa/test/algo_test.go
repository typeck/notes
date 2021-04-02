package test


//sunday算法, 字符匹配
//https://cloud.tencent.com/developer/article/1683667?from=information.detail.sunday%E7%AE%97%E6%B3%95
func strStr(haystack, needle string) int {
	if len(needle) == 0 {
		return 0
	}
	m := make(map[byte]int)
	//记录needle中每个字符出现的位置
	for i := range needle {
		m[needle[i]] = i 
	}
	for i := 0; i <= len(haystack); {
		tmpi := i
		j := 0
		for ;j < len(needle) && tmpi < len(haystack); j++ {
			if haystack[tmpi] != needle[j] {
				break
			}
			tmpi++
		}
		//匹配成功
		if j == len(needle) {
			return i
		}
		if i + len(needle) >= len(haystack) {
			return -1
		}
		hit, exists := m[haystack[i+len(needle)]]
		if exists {
			i += len(needle) - hit
		}else {
			i +=len(needle) + 1
		}
	}
	return -1
}