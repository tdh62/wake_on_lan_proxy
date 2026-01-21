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
            max-width: 600px;
            width: 100%;
        }
        h1 {
            color: #333;
            text-align: center;
            margin-bottom: 30px;
            font-size: 28px;
        }
        h2 {
            color: #555;
            font-size: 18px;
            margin-bottom: 15px;
            margin-top: 30px;
        }
        .form-group {
            margin-bottom: 20px;
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
        .history-section {
            margin-top: 30px;
            padding-top: 30px;
            border-top: 2px solid #e0e0e0;
        }
        .history-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 15px;
        }
        .clear-all-btn {
            padding: 6px 12px;
            font-size: 12px;
            background: #dc3545;
            width: auto;
        }
        .clear-all-btn:hover {
            background: #c82333;
        }
        .history-list {
            max-height: 300px;
            overflow-y: auto;
        }
        .history-item {
            background: #f8f9fa;
            border: 1px solid #e0e0e0;
            border-radius: 8px;
            padding: 12px;
            margin-bottom: 10px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            cursor: pointer;
            transition: all 0.2s;
        }
        .history-item:hover {
            background: #e9ecef;
            border-color: #667eea;
            transform: translateX(5px);
        }
        .history-info {
            flex: 1;
        }
        .history-name {
            font-weight: 600;
            color: #333;
            margin-bottom: 4px;
        }
        .history-details {
            font-size: 12px;
            color: #666;
        }
        .history-actions {
            display: flex;
            gap: 8px;
        }
        .delete-btn {
            padding: 6px 12px;
            font-size: 12px;
            background: #dc3545;
            width: auto;
        }
        .delete-btn:hover {
            background: #c82333;
        }
        .empty-history {
            text-align: center;
            color: #999;
            padding: 20px;
            font-size: 14px;
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
        <form action="/wake" method="POST" id="wakeForm" onsubmit="saveToHistory(event)">
            <div class="form-group">
                <label for="deviceName">设备名称（可选）</label>
                <input type="text" id="deviceName" name="deviceName" placeholder="例如: 我的电脑">
                <div class="hint">为设备设置一个易记的名称</div>
            </div>
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

        <div class="history-section">
            <div class="history-header">
                <h2>📋 历史记录</h2>
                <button class="clear-all-btn" onclick="clearAllHistory()">清空全部</button>
            </div>
            <div class="history-list" id="historyList">
                <div class="empty-history">暂无历史记录</div>
            </div>
        </div>
    </div>

    <script>
        const MAX_HISTORY = 10;

        // 页面加载时显示历史记录
        window.onload = function() {
            displayHistory();
        };

        // 保存到历史记录
        function saveToHistory(event) {
            const deviceName = document.getElementById('deviceName').value.trim();
            const mac = document.getElementById('mac').value.trim();
            const ip = document.getElementById('ip').value.trim();

            if (!mac) return;

            const record = {
                deviceName: deviceName || mac,
                mac: mac,
                ip: ip,
                timestamp: new Date().toISOString()
            };

            let history = getHistory();

            // 检查是否已存在相同MAC地址的记录，如果存在则更新
            const existingIndex = history.findIndex(item => item.mac.toLowerCase() === mac.toLowerCase());
            if (existingIndex !== -1) {
                history.splice(existingIndex, 1);
            }

            // 添加到开头
            history.unshift(record);

            // 限制历史记录数量
            if (history.length > MAX_HISTORY) {
                history = history.slice(0, MAX_HISTORY);
            }

            localStorage.setItem('wolHistory', JSON.stringify(history));
        }

        // 获取历史记录
        function getHistory() {
            const history = localStorage.getItem('wolHistory');
            return history ? JSON.parse(history) : [];
        }

        // 显示历史记录
        function displayHistory() {
            const history = getHistory();
            const historyList = document.getElementById('historyList');

            if (history.length === 0) {
                historyList.innerHTML = '<div class="empty-history">暂无历史记录</div>';
                return;
            }

            historyList.innerHTML = history.map((record, index) => {
                const date = new Date(record.timestamp);
                const dateStr = date.toLocaleString('zh-CN', {
                    month: '2-digit',
                    day: '2-digit',
                    hour: '2-digit',
                    minute: '2-digit'
                });

                return ` + "`" + `
                    <div class="history-item" onclick="loadFromHistory(${index})">
                        <div class="history-info">
                            <div class="history-name">${escapeHtml(record.deviceName)}</div>
                            <div class="history-details">MAC: ${escapeHtml(record.mac)} | IP: ${escapeHtml(record.ip)} | ${dateStr}</div>
                        </div>
                        <div class="history-actions">
                            <button class="delete-btn" onclick="deleteHistory(event, ${index})">删除</button>
                        </div>
                    </div>
                ` + "`" + `;
            }).join('');
        }

        // 从历史记录加载
        function loadFromHistory(index) {
            const history = getHistory();
            if (index >= 0 && index < history.length) {
                const record = history[index];
                document.getElementById('deviceName').value = record.deviceName;
                document.getElementById('mac').value = record.mac;
                document.getElementById('ip').value = record.ip;

                // 滚动到表单顶部
                window.scrollTo({ top: 0, behavior: 'smooth' });
            }
        }

        // 删除单个历史记录
        function deleteHistory(event, index) {
            event.stopPropagation();

            if (confirm('确定要删除这条记录吗？')) {
                let history = getHistory();
                history.splice(index, 1);
                localStorage.setItem('wolHistory', JSON.stringify(history));
                displayHistory();
            }
        }

        // 清空所有历史记录
        function clearAllHistory() {
            if (confirm('确定要清空所有历史记录吗？')) {
                localStorage.removeItem('wolHistory');
                displayHistory();
            }
        }

        // HTML转义函数
        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }
    </script>
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
            max-width: 600px;
            width: 100%;
        }
        h1 {
            color: #333;
            text-align: center;
            margin-bottom: 30px;
            font-size: 28px;
        }
        h2 {
            color: #555;
            font-size: 18px;
            margin-bottom: 15px;
            margin-top: 30px;
        }
        .form-group {
            margin-bottom: 20px;
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
        .history-section {
            margin-top: 30px;
            padding-top: 30px;
            border-top: 2px solid #e0e0e0;
        }
        .history-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 15px;
        }
        .clear-all-btn {
            padding: 6px 12px;
            font-size: 12px;
            background: #dc3545;
            width: auto;
        }
        .clear-all-btn:hover {
            background: #c82333;
        }
        .history-list {
            max-height: 300px;
            overflow-y: auto;
        }
        .history-item {
            background: #f8f9fa;
            border: 1px solid #e0e0e0;
            border-radius: 8px;
            padding: 12px;
            margin-bottom: 10px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            cursor: pointer;
            transition: all 0.2s;
        }
        .history-item:hover {
            background: #e9ecef;
            border-color: #667eea;
            transform: translateX(5px);
        }
        .history-info {
            flex: 1;
        }
        .history-name {
            font-weight: 600;
            color: #333;
            margin-bottom: 4px;
        }
        .history-details {
            font-size: 12px;
            color: #666;
        }
        .history-actions {
            display: flex;
            gap: 8px;
        }
        .delete-btn {
            padding: 6px 12px;
            font-size: 12px;
            background: #dc3545;
            width: auto;
        }
        .delete-btn:hover {
            background: #c82333;
        }
        .empty-history {
            text-align: center;
            color: #999;
            padding: 20px;
            font-size: 14px;
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
        <form action="/wake" method="POST" id="wakeForm" onsubmit="saveToHistory(event)">
            <div class="form-group">
                <label for="deviceName">设备名称（可选）</label>
                <input type="text" id="deviceName" name="deviceName" placeholder="例如: 我的电脑">
                <div class="hint">为设备设置一个易记的名称</div>
            </div>
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

        <div class="history-section">
            <div class="history-header">
                <h2>📋 历史记录</h2>
                <button class="clear-all-btn" onclick="clearAllHistory()">清空全部</button>
            </div>
            <div class="history-list" id="historyList">
                <div class="empty-history">暂无历史记录</div>
            </div>
        </div>
    </div>

    <script>
        const MAX_HISTORY = 10;

        // 页面加载时显示历史记录
        window.onload = function() {
            displayHistory();
        };

        // 保存到历史记录
        function saveToHistory(event) {
            const deviceName = document.getElementById('deviceName').value.trim();
            const mac = document.getElementById('mac').value.trim();
            const ip = document.getElementById('ip').value.trim();

            if (!mac) return;

            const record = {
                deviceName: deviceName || mac,
                mac: mac,
                ip: ip,
                timestamp: new Date().toISOString()
            };

            let history = getHistory();

            // 检查是否已存在相同MAC地址的记录，如果存在则更新
            const existingIndex = history.findIndex(item => item.mac.toLowerCase() === mac.toLowerCase());
            if (existingIndex !== -1) {
                history.splice(existingIndex, 1);
            }

            // 添加到开头
            history.unshift(record);

            // 限制历史记录数量
            if (history.length > MAX_HISTORY) {
                history = history.slice(0, MAX_HISTORY);
            }

            localStorage.setItem('wolHistory', JSON.stringify(history));
        }

        // 获取历史记录
        function getHistory() {
            const history = localStorage.getItem('wolHistory');
            return history ? JSON.parse(history) : [];
        }

        // 显示历史记录
        function displayHistory() {
            const history = getHistory();
            const historyList = document.getElementById('historyList');

            if (history.length === 0) {
                historyList.innerHTML = '<div class="empty-history">暂无历史记录</div>';
                return;
            }

            historyList.innerHTML = history.map((record, index) => {
                const date = new Date(record.timestamp);
                const dateStr = date.toLocaleString('zh-CN', {
                    month: '2-digit',
                    day: '2-digit',
                    hour: '2-digit',
                    minute: '2-digit'
                });

                return ` + "`" + `
                    <div class="history-item" onclick="loadFromHistory(${index})">
                        <div class="history-info">
                            <div class="history-name">${escapeHtml(record.deviceName)}</div>
                            <div class="history-details">MAC: ${escapeHtml(record.mac)} | IP: ${escapeHtml(record.ip)} | ${dateStr}</div>
                        </div>
                        <div class="history-actions">
                            <button class="delete-btn" onclick="deleteHistory(event, ${index})">删除</button>
                        </div>
                    </div>
                ` + "`" + `;
            }).join('');
        }

        // 从历史记录加载
        function loadFromHistory(index) {
            const history = getHistory();
            if (index >= 0 && index < history.length) {
                const record = history[index];
                document.getElementById('deviceName').value = record.deviceName;
                document.getElementById('mac').value = record.mac;
                document.getElementById('ip').value = record.ip;

                // 滚动到表单顶部
                window.scrollTo({ top: 0, behavior: 'smooth' });
            }
        }

        // 删除单个历史记录
        function deleteHistory(event, index) {
            event.stopPropagation();

            if (confirm('确定要删除这条记录吗？')) {
                let history = getHistory();
                history.splice(index, 1);
                localStorage.setItem('wolHistory', JSON.stringify(history));
                displayHistory();
            }
        }

        // 清空所有历史记录
        function clearAllHistory() {
            if (confirm('确定要清空所有历史记录吗？')) {
                localStorage.removeItem('wolHistory');
                displayHistory();
            }
        }

        // HTML转义函数
        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }
    </script>
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

	// 解析广播地址
	broadcastAddr, err := net.ResolveUDPAddr("udp", broadcastIP+":9")
	if err != nil {
		return fmt.Errorf("无法解析广播地址: %v", err)
	}

	// 创建UDP连接，监听所有接口
	localAddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return fmt.Errorf("无法解析本地地址: %v", err)
	}

	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return fmt.Errorf("无法创建UDP连接: %v", err)
	}
	defer conn.Close()

	// 发送魔术包到广播地址
	n, err := conn.WriteToUDP(magicPacket, broadcastAddr)
	if err != nil {
		return fmt.Errorf("发送数据包失败: %v", err)
	}

	log.Printf("已发送唤醒包到 MAC: %s, 广播地址: %s, 发送字节数: %d", macAddr, broadcastIP, n)
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
