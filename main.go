package main

import (
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
)

type PageData struct {
	Message string
	Success bool
}

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/wake", handleWake)

	fmt.Println("Wake-on-LAN服务已启动，监听端口: 24000")
	fmt.Println("访问 http://localhost:24000 使用服务")
	log.Fatal(http.ListenAndServe(":24000", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>局域网唤醒服务</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            padding: 40px;
            max-width: 500px;
            width: 100%;
        }
        h1 {
            color: #333;
            text-align: center;
            margin-bottom: 30px;
            font-size: 28px;
        }
        .form-group {
            margin-bottom: 25px;
        }
        label {
            display: block;
            color: #555;
            font-weight: 600;
            margin-bottom: 8px;
            font-size: 14px;
        }
        input[type="text"] {
            width: 100%;
            padding: 12px 15px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 16px;
            transition: border-color 0.3s;
        }
        input[type="text"]:focus {
            outline: none;
            border-color: #667eea;
        }
        .hint {
            font-size: 12px;
            color: #888;
            margin-top: 5px;
        }
        button {
            width: 100%;
            padding: 14px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        button:hover {
            transform: translateY(-2px);
            box-shadow: 0 10px 20px rgba(102, 126, 234, 0.4);
        }
        button:active {
            transform: translateY(0);
        }
        .message {
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 20px;
            font-size: 14px;
        }
        .success {
            background-color: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        .error {
            background-color: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🌐 局域网唤醒服务</h1>
        {{if .Message}}
        <div class="message {{if .Success}}success{{else}}error{{end}}">
            {{.Message}}
        </div>
        {{end}}
        <form action="/wake" method="POST">
            <div class="form-group">
                <label for="mac">目标设备MAC地址</label>
                <input type="text" id="mac" name="mac" placeholder="例如: AA:BB:CC:DD:EE:FF" required>
                <div class="hint">支持格式: AA:BB:CC:DD:EE:FF 或 AA-BB-CC-DD-EE-FF</div>
            </div>
            <div class="form-group">
                <label for="ip">广播地址（可选）</label>
                <input type="text" id="ip" name="ip" placeholder="例如: 192.168.1.255" value="255.255.255.255">
                <div class="hint">默认使用全局广播地址 255.255.255.255</div>
            </div>
            <button type="submit">发送唤醒包</button>
        </form>
    </div>
</body>
</html>`

	t, err := template.New("index").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageData{
		Message: "",
		Success: false,
	}

	t.Execute(w, data)
}

func handleWake(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	macAddr := r.FormValue("mac")
	broadcastIP := r.FormValue("ip")

	if broadcastIP == "" {
		broadcastIP = "255.255.255.255"
	}

	err := sendWakeOnLAN(macAddr, broadcastIP)

	tmpl := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>局域网唤醒服务</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            padding: 40px;
            max-width: 500px;
            width: 100%;
        }
        h1 {
            color: #333;
            text-align: center;
            margin-bottom: 30px;
            font-size: 28px;
        }
        .form-group {
            margin-bottom: 25px;
        }
        label {
            display: block;
            color: #555;
            font-weight: 600;
            margin-bottom: 8px;
            font-size: 14px;
        }
        input[type="text"] {
            width: 100%;
            padding: 12px 15px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 16px;
            transition: border-color 0.3s;
        }
        input[type="text"]:focus {
            outline: none;
            border-color: #667eea;
        }
        .hint {
            font-size: 12px;
            color: #888;
            margin-top: 5px;
        }
        button {
            width: 100%;
            padding: 14px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        button:hover {
            transform: translateY(-2px);
            box-shadow: 0 10px 20px rgba(102, 126, 234, 0.4);
        }
        button:active {
            transform: translateY(0);
        }
        .message {
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 20px;
            font-size: 14px;
        }
        .success {
            background-color: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        .error {
            background-color: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🌐 局域网唤醒服务</h1>
        {{if .Message}}
        <div class="message {{if .Success}}success{{else}}error{{end}}">
            {{.Message}}
        </div>
        {{end}}
        <form action="/wake" method="POST">
            <div class="form-group">
                <label for="mac">目标设备MAC地址</label>
                <input type="text" id="mac" name="mac" placeholder="例如: AA:BB:CC:DD:EE:FF" required>
                <div class="hint">支持格式: AA:BB:CC:DD:EE:FF 或 AA-BB-CC-DD-EE-FF</div>
            </div>
            <div class="form-group">
                <label for="ip">广播地址（可选）</label>
                <input type="text" id="ip" name="ip" placeholder="例如: 192.168.1.255" value="255.255.255.255">
                <div class="hint">默认使用全局广播地址 255.255.255.255</div>
            </div>
            <button type="submit">发送唤醒包</button>
        </form>
    </div>
</body>
</html>`

	t, _ := template.New("index").Parse(tmpl)

	data := PageData{}
	if err != nil {
		data.Message = fmt.Sprintf("发送失败: %v", err)
		data.Success = false
	} else {
		data.Message = fmt.Sprintf("唤醒包已成功发送到 %s (广播地址: %s)", macAddr, broadcastIP)
		data.Success = true
	}

	t.Execute(w, data)
}

func sendWakeOnLAN(macAddr string, broadcastIP string) error {
	// 解析MAC地址
	mac, err := parseMACAddress(macAddr)
	if err != nil {
		return fmt.Errorf("无效的MAC地址: %v", err)
	}

	// 创建魔术包
	magicPacket := createMagicPacket(mac)

	// 发送UDP广播包
	addr, err := net.ResolveUDPAddr("udp", broadcastIP+":9")
	if err != nil {
		return fmt.Errorf("无法解析广播地址: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return fmt.Errorf("无法创建UDP连接: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write(magicPacket)
	if err != nil {
		return fmt.Errorf("发送数据包失败: %v", err)
	}

	log.Printf("已发送唤醒包到 MAC: %s, 广播地址: %s", macAddr, broadcastIP)
	return nil
}

func parseMACAddress(macAddr string) ([]byte, error) {
	// 移除常见的分隔符
	macAddr = strings.ReplaceAll(macAddr, ":", "")
	macAddr = strings.ReplaceAll(macAddr, "-", "")
	macAddr = strings.ReplaceAll(macAddr, " ", "")

	// 验证格式
	matched, _ := regexp.MatchString("^[0-9A-Fa-f]{12}$", macAddr)
	if !matched {
		return nil, fmt.Errorf("MAC地址格式不正确，应为12位十六进制字符")
	}

	// 转换为字节数组
	mac, err := hex.DecodeString(macAddr)
	if err != nil {
		return nil, err
	}

	return mac, nil
}

func createMagicPacket(mac []byte) []byte {
	// 魔术包格式: 6个0xFF字节 + 16次重复的MAC地址
	packet := make([]byte, 102)

	// 前6个字节为0xFF
	for i := 0; i < 6; i++ {
		packet[i] = 0xFF
	}

	// 后面重复16次MAC地址
	for i := 0; i < 16; i++ {
		copy(packet[6+i*6:], mac)
	}

	return packet
}
