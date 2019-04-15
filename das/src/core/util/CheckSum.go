package util

func CheckSum(data []byte) uint16 {
	var (
		sum    uint16
		length int = len(data)
		index  int
	)
	sum = 0

	// 以每16位为单位进行求和，直到所有的字节全部求完或者只剩下一个8位字节（如果剩余一个8位字节说明字节数为奇数个）
	for length > 0 {
		sum += uint16(data[index])//<<8 //+ uint32(data[index+1])
		index += 1
		length -= 1
	}
	// 如果字节数为奇数个，要加上最后剩下的那个8位字节
	/*if length > 0 {
		sum += uint16(data[index])
	}*/

	// 加上高16位进位的部分
	// sum += (sum >> 16)

	// 别忘了返回的时候先求反
	return sum
}
