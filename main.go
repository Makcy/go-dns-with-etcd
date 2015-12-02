package main

import (
	"flag"
	"log"
	"net"
)

//启动命令 sudo ./mydns -dns=114.114.114.114:53
var (
	dns     = flag.String("dns", "114.114.114.114:53", "DNS地址（本地查不到时向该服务器查询）")
	isDebug = flag.Bool("debug", true, "是否启用调试模式")
)

func main() {
	flag.Parse()

	sock, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 53,
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	defer sock.Close()

	log.Println("启动服务并监听 0.0.0.0:53 端口...")

	for {
		// 读取UDP数据包
		buf := make([]byte, 1024)
		n, addr, _ := sock.ReadFromUDP(buf)
		ip := ""
		go func() {
			// DNS报文解析
			msg := UnpackMsg(buf[:n])
			oriKey := msg.GetQuestion(0).Name
			realKey, isMatch := matchRegex(oriKey)
			if isMatch {
				//Case 1:以.docker结尾的匹配格式 进行查询
				ip, _ = getValueInEtcd(realKey, "http://127.0.0.1:2379")
			} else {
				//Case 2:防止非docker服务，但符合key的情况发生
				ip = ""
			}

			if ip != "" {
				msg.SetResponse()
				msg.AddAnswer(NewA(msg.GetQuestion(0).Name, ip))
				ret := PackMsg(msg)
				sock.WriteToUDP(ret, addr)
				debug("[L]解析: ", msg.GetQuestion(0).Name)
			} else {
				ret, err := query(buf[:n])
				check_error(err)
				if err == nil {
					sock.WriteToUDP(ret, addr)
				}
				debug("[R]解析: ", msg.GetQuestion(0).Name)
			}
		}()

	}

}

func query(msg []byte) ([]byte, error) {
	raddr, err := net.ResolveUDPAddr("udp", *dns)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	defer conn.Close()

	if err != nil {
		return nil, err
	}

	_, err = conn.Write(msg)
	if err != nil {
		return nil, err
	}

	ret := make([]byte, 4096)
	n, _, err := conn.ReadFromUDP(ret)
	if err != nil {
		return nil, err
	}

	return ret[0:n], nil
}

func check_error(err error) {
	if err != nil {
		log.Println(err)
	}
}

func debug(fmt ...interface{}) {
	if *isDebug {
		log.Println(fmt...)
	}
}
