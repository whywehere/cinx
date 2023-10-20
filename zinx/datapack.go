package zinx

import (
	"bytes"
	"cinx/zinx/utils/global"
	"cinx/zinx/ziface"
	"encoding/binary"
	"errors"
	"io"
)

const HeaderLen = 8

type DataPack struct{}

// NewDataPack 封包拆包实例初始化方法
func NewDataPack() *DataPack {
	return &DataPack{}
}

// GetHeadLen 获取包头长度方法
func (dp *DataPack) GetHeadLen() uint32 {
	//Id uint32(4字节) +  DataLen uint32(4字节)
	return 8
}

func writeBinary(dataBuffer io.Writer, data interface{}) error {
	if err := binary.Write(dataBuffer, binary.LittleEndian, data); err != nil {
		return err
	}
	return nil
}

// Pack 封包方法(压缩数据)
func (dp *DataPack) Pack(msg ziface.IMessage) ([]byte, error) {
	dataBuffer := bytes.NewBuffer([]byte{})

	if err := writeBinary(dataBuffer, msg.GetDataLen()); err != nil {
		return nil, err
	}

	if err := writeBinary(dataBuffer, msg.GetMsgId()); err != nil {
		return nil, err
	}

	if err := writeBinary(dataBuffer, msg.GetData()); err != nil {
		return nil, err
	}

	return dataBuffer.Bytes(), nil

}

// Unpack 拆包方法(解压数据)
func (dp *DataPack) Unpack(binaryData []byte) (ziface.IMessage, error) {
	//创建一个从输入二进制数据的ioReader
	dataBuff := bytes.NewReader(binaryData)

	//只解压head的信息，得到dataLen和msgID
	msg := &Message{}

	//读dataLen
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.DataLen); err != nil {
		return nil, err
	}

	//读msgID
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.Id); err != nil {
		return nil, err
	}

	//判断dataLen的长度是否超出我们允许的最大包长度
	if global.MaxPacketSize > 0 && msg.DataLen > global.MaxPacketSize {
		return nil, errors.New("too large msg data recieved")
	}

	//这里只需要把head的数据拆包出来就可以了，然后再通过head的长度，再从conn读取一次数据
	return msg, nil
}
