# DAS_go

// 转换8进制utf-8字符串到中文
// eg: `\346\200\241` -> 怡
func convertOctonaryUtf8(in string) string {
	s := []byte(in)
	reg := regexp.MustCompile(`\\[0-7]{3}`)

	out := reg.ReplaceAllFunc(s,
		func(b []byte) []byte {
			i, _ := strconv.ParseInt(string(b[1:]), 8, 0)
			return []byte{byte(i)}
		})
	return string(out)
}