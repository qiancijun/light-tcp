package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/qiancijun/ltcp"
)

// Client 客户端

func main() {
	raddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	laddr, err := net.ResolveUDPAddr("udp", ":")
	if err != nil {
		panic(err)
	}
	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		panic(err)
	}

	// p := ltcp.NewPacket(ltcp.PacketTypeData, []byte("Hello"))
	// data, err := p.Serialize()
	// if err != nil {
	// 	panic(err)
	// }
	// conn.Write(data)
	// return

	ltcpConn := ltcp.NewConn(conn, ltcp.DefaultLtcpConnOptions)
	fmt.Printf("client %s connect to server %s\n", laddr.String(), raddr.String())
	defer ltcpConn.Close()

	go func() {
		for {
			data := make([]byte, 32767)
			n, err := ltcpConn.Read(data)
			if err != nil {
				fmt.Println("read error: ", err)
				break
			}
			fmt.Println("client receive data: ", string(data[:n]))
		}
	}()

	// datas := [][]byte{
	// 	[]byte("Hello"),
	// 	[]byte(","),
	// 	[]byte("World"),
	// }

	// for _, data := range datas {
	// 	fmt.Println("ready to send data to server, data len: ", len(data))
	// 	_, err := ltcpConn.Write(data)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		break
	// 	}
	// }

	file, err := os.OpenFile("examples/files/saki2.jpg", os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	data := make([]byte, 8192)
	for {
		n, err := file.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		ltcpConn.Write(data[:n])
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(1 * time.Second)

	info1, _ := file.Stat()
	file2, _ := os.OpenFile("examples/files/copy.jpg", os.O_RDONLY, 0644)
	info2, _ := file2.Stat()
	fmt.Println(info1.Size())
	fmt.Println(info2.Size())

	compareFiles("examples/files/copy.jpg", "examples/files/saki2.jpg")
}

// 改进的 hexdump 函数
func hexdump(data []byte, label string) {
	if len(data) == 0 {
		fmt.Printf("%s: [空数据]\n", label)
		return
	}

	fmt.Printf("=== %s (长度: %d 字节) ===\n", label, len(data))

	for i := 0; i < len(data); i += 16 {
		// 打印偏移量
		fmt.Printf("%08x: ", i)

		// 打印十六进制
		for j := 0; j < 16; j++ {
			if i+j < len(data) {
				fmt.Printf("%02x ", data[i+j])
			} else {
				fmt.Print("   ") // 对齐填充
			}
			if j == 7 { // 8字节分隔
				fmt.Print(" ")
			}
		}

		// 打印ASCII表示
		fmt.Print(" |")
		for j := 0; j < 16; j++ {
			if i+j >= len(data) {
				break
			}
			b := data[i+j]
			if b >= 32 && b <= 126 { // 可打印ASCII字符
				fmt.Printf("%c", b)
			} else {
				fmt.Print(".")
			}
		}
		fmt.Println("|")
	}
}

// 改进的 compareFiles 函数
func compareFiles(original, received string) error {
	orig, err := ioutil.ReadFile(original)
	if err != nil {
		return fmt.Errorf("读取原始文件失败: %v", err)
	}

	recv, err := ioutil.ReadFile(received)
	if err != nil {
		return fmt.Errorf("读取接收文件失败: %v", err)
	}

	// 检查文件大小
	if len(orig) != len(recv) {
		fmt.Printf("文件大小不同 - 原始: %d 字节, 接收: %d 字节\n", len(orig), len(recv))
	}

	// 找出第一个差异点
	minLen := min(len(orig), len(recv))
	for i := 0; i < minLen; i++ {
		if orig[i] != recv[i] {
			fmt.Printf("\n发现首个差异位置: 偏移量 %d (0x%x)\n", i, i)

			// 显示差异周围的上下文
			start := max(0, i-16)
			end := min(minLen, i+16)

			hexdump(orig[start:end], "原始文件数据")
			hexdump(recv[start:end], "接收文件数据")

			// 显示具体差异字节
			fmt.Println("\n差异字节详情:")
			for j := start; j < end; j++ {
				if orig[j] != recv[j] {
					fmt.Printf("位置 %d (0x%x): 原始=0x%02x 接收=0x%02x\n",
						j, j, orig[j], recv[j])
				}
			}
			return nil
		}
	}

	if len(orig) != len(recv) {
		fmt.Printf("文件内容前%d字节相同，但大小不同\n", minLen)
	} else {
		fmt.Println("两个文件内容完全相同")
	}
	return nil
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
