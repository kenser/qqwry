package main

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

var IpData ipData

func (this *ipData) InitIpData() (rs interface{}) {

	// 判断文件是否存在
	_, err := os.Stat(this.FilePath)
	if err != nil && os.IsNotExist(err) {
		rs = errors.New("文件不存在")
		return
	}

	// 打开文件句柄
	this.Path, err = os.OpenFile(this.FilePath, os.O_RDONLY, 0400)
	if err != nil {
		rs = err
		return
	}

	this.Path.Seek(0, 0)
	indexPos := make([]byte, 8)
	if _, err := this.Path.Read(indexPos); err != nil {
		rs = err
		return
	}

	start := binary.LittleEndian.Uint32(indexPos[:4])
	end := binary.LittleEndian.Uint32(indexPos[4:])

	// 索引数量
	indexNum := (end - start) / INDEX_NUM + 1

	// 临时索引,当文件加载完成后将索引给到 ipData 提供给查询使用
	tmpIndex := make([]index, indexNum)
	// 临时文件数据,当文件加载完成后,将数据给到 ipData 提供给查询使用
	tmpData := make(map[uint32]qqwry)

	// index 的缓存数据
	indexBuf := make([]byte, INDEX_NUM)
	// data 的临时缓冲数据
	dataTmpBuf := make([]byte, 1)
	// 从数据区获取 IP 地址
	ipBuf := make([]byte, 4)
	// 数据的偏移量
	dataOffset := uint32(0)

	j := 0
	zeroNum := 0
	// 开始加载索引
	for i := uint32(0); i < indexNum; i++ {

		// 将文件的指针跳转到索引开始的位置
		this.Path.Seek(int64(start+i*INDEX_NUM), 0)
		if _, err := this.Path.Read(indexBuf); err != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		dataOffset = byteToUInt32(indexBuf[4:])

		tmpIndex[j] = index{
			Ip:     binary.LittleEndian.Uint32(indexBuf[:4]),
			Offset: dataOffset,
		}

		j++

		// 开始获取 IP 的地址数据
		this.Path.Seek(int64(dataOffset), 0)

		if _, err = this.Path.Read(ipBuf); err != nil {
			continue
		}

		tmpQQwry := qqwry{
			Ip:      binary.LittleEndian.Uint32(ipBuf),
			Country: make([]byte, 0, 50),
			Area:    make([]byte, 0, 50),
		}

		zeroNum = 0
		for i := 0; i < 1024; i++ {

			if zeroNum > 1 {
				break
			}

			if _, err = this.Path.Read(dataTmpBuf); err != nil {
				continue
			}

			if dataTmpBuf[0] == 0 {
				zeroNum++
				continue
			}

			if zeroNum == 0 {
				tmpQQwry.Country = append(tmpQQwry.Country, dataTmpBuf[0])
			} else if zeroNum == 1 {
				tmpQQwry.Area = append(tmpQQwry.Area, dataTmpBuf[0])
			}
		}

		tmpData[dataOffset] = tmpQQwry

	}

	this.Index = tmpIndex
	this.Data = tmpData

	return
}

// 将 byte 转换为uint32
func byteToUInt32(data []byte) uint32 {
	i := uint32(data[0]) & 0xff
	i |= (uint32(data[1]) << 8) & 0xff00
	i |= (uint32(data[2]) << 16) & 0xff0000
	return i
}