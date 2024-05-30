package ordinals

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func checkFileDir(dirPath string) {

	// 检查文件目录是否存在
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// 如果文件夹不存在，则创建文件夹
		err := os.MkdirAll(dirPath, 0777)
		if err != nil {
			fmt.Println("创建文件夹失败:", err)
			return
		}
		fmt.Println("文件夹创建成功:", dirPath)
	} else {
		// 如果文件夹存在，则删除文件夹下的内容
		err := removeAllContents(dirPath)
		if err != nil {
			fmt.Println("删除文件夹内容失败:", err)
			return
		}
		fmt.Println("文件夹内容已删除:", dirPath)
	}
}

// removeAllContents 删除文件夹下的所有内容
func removeAllContents(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())
		if entry.IsDir() {
			// 递归删除子文件夹
			err := removeAllContents(fullPath)
			if err != nil {
				return err
			}
		} else {
			// 删除文件
			err := os.Remove(fullPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func TestGenesisCard(t *testing.T) {
	checkFileDir("./inscribefile")
	for i := 1; i <= 2100; i++ {
		GenesisCard(32)
		fileName := fmt.Sprintf("./inscribefile/%d.svg", i)
		err := ioutil.WriteFile(fileName, []byte(GenesisCard(int32(i))), 0644)
		if err != nil {
			fmt.Println("file save fail:", err)
			return
		}

		fmt.Println("file save success:", fileName)
	}

}
