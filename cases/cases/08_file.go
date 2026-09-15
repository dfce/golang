package cases

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func FileCase() {
	// copyD2D()
	// readWriteFile()
	// readAllLine()
	// readBuf()
	// readScanner()
	streamCase()
}

// 文件复制
var (
	// cwd, _ = os.Getwd()
	source = filepath.Join("cases", "filecase", "source")
	dest   = filepath.Join("cases", "filecase", "dest")
)

// 拷贝文件
func copyD2D() {
	list := getFiles(source)
	fmt.Println("source:", source)

	for _, f := range list {
		_, name := path.Split(f)
		destFileName := dest + "/copy-" + name
		// copy
		fmt.Println("copyFile d:", f, destFileName)
		res, _ := copyFile(f, destFileName)
		fmt.Println("copy res:", res)
	}
}

// 读写文件
func readWriteFile() {
	list := getFiles(source)
	for _, f := range list {
		bytes, err := os.ReadFile(f)

		if err != nil {
			fmt.Println("os.ReadFile error:", err)
			return
		}
		_, name := path.Split(f)
		destName := dest + "/normal-new-" + name
		err = os.WriteFile(destName, bytes, 0644)
		if err != nil {
			fmt.Println("os.WriteFile error:", err)
			return
		}
	}
}

// 一次性读取文件， 按行拆分 并打印，只时候小文件
func readAllLine() {
	var fpath = filepath.Join("cases", "filecase", "source", "src.txt")
	fileHandler := openFile(fpath)
	defer fileHandler.Close()

	bytes, err := io.ReadAll(fileHandler)
	if err != nil {
		fmt.Println("读取文件失败")
		return
	}

	list := strings.Split(string(bytes), "\n")
	for _, l := range list {
		fmt.Println(l)
	}
}

// 通过bufio按行读取
// bufio 通过对io模块的封装，提供了缓冲区，能一定程度上减少大数据块读写的开销
// 当发起读写操作时，会尝试从缓冲区读取数据，缓冲区没有数据后才会从数据源获取
// 缓冲区大小默认4K
func readBuf() {
	var fpath = filepath.Join("cases", "filecase", "source", "src.txt")
	fileHandler := openFile(fpath)
	defer fileHandler.Close()

	reader := bufio.NewReader(fileHandler)

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		fmt.Println(line)
	}
}

// 通过scanner 按行读取， 单行默认大小64K
func readScanner() {
	var fpath = filepath.Join("cases", "filecase", "source", "src.txt")
	fileHandler := openFile(fpath)
	defer fileHandler.Close()

	scanner := bufio.NewScanner(fileHandler)
	if scanner.Err() != nil {
		fmt.Println("readScanner error", scanner.Err().Error())
		return
	}
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
	}
}

// stream 流读取写入
func streamCase() {
	var srcName = filepath.Join("cases", "filecase", "source", "image.png")
	src := openFile(srcName)
	defer src.Close()

	destName := filepath.Join("cases", "filecase", "dest", "image-rbuf.png")
	dest, err := os.OpenFile(destName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("打开目标文件失败", err)
		return
	}
	defer dest.Close()

	buf := make([]byte, 1024)
	for {
		n, err := src.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println("按buf读取文件失败", err)
		}
		if n == 0 {
			break
		}
		len, _ := dest.Write(buf[:n])
		fmt.Println("按buf写入文件长度：", len)
	}
}

func openFile(fpath string) *os.File {
	fmt.Println("fpath:", fpath)
	fileHandler, err := os.OpenFile(fpath, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println("打开文件失败", err)
		return nil
	}
	return fileHandler
}
func getFiles(dir string) []string {
	fs, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("目录%s读取文件失败， empty\n", dir)
		return nil
	}

	list := make([]string, 0)
	for _, f := range fs {
		fullName := strings.Trim(dir, "/") + "/" + f.Name()
		if f.IsDir() {
			l := getFiles(fullName)
			list = append(list, l...)
			continue
		}
		list = append(list, fullName)
	}
	return list
}

func copyFile(srcName, destName string) (int64, error) {
	src, err := os.Open(srcName)
	if err != nil {
		fmt.Println("打开源文件目录失败")
		return 0, err
	}
	defer src.Close()

	dest, err := os.OpenFile(destName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("打开目标目录失败")
		return 0, err
	}
	defer dest.Close()

	return io.Copy(dest, src)
}
