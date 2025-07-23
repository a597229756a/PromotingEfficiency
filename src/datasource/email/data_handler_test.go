package email

import (
	"fmt"
	"os"
	"testing"
)

func TestReadXlsx(t *testing.T) {
	fileBytes, err := ByteFromXlse()
	if err != nil {
		fmt.Println(err)
	}
	dfw, err := ReadXlsx(fileBytes, "进离港航班")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(dfw.qf.Select("ID","COBT"))
}

func ByteFromXlse() ([]byte, error) {
	filePath := "/home/ubuntu-cz/go/src/PromotingEfficiency/data/兴效能.xlsx"

	// 方法1：直接打开文件
	// xlFile, err := xlsx.OpenFile(filePath)
	// if err != nil {
	// 	fmt.Printf("打开文件失败: %v\n", err)
	// 	return
	// }

	// 方法2：如果你想使用OpenBinary，需要先读取文件内容

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return nil, err
	}
	return fileBytes, nil
}
