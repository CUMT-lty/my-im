package connect

import (
	"github.com/lty/my-go-chat/proto"
	"sync"
	"sync/atomic"
)

// TODO: 桶机制，学了一些 go map 对 bmap 的管理
type Bucket struct {
	cLock         sync.RWMutex     // protect the channels for chs
	chs           map[int]*Channel // map sub key to a channel
	bucketOptions BucketOptions
	rooms         map[int]*Room // bucket room channels
	routines      []chan *proto.PushRoomMsgRequest
	routinesNum   uint64
	//broadcast     chan []byte
}

// 桶配置选项
type BucketOptions struct {
	ChannelSize   int
	RoomSize      int
	RoutineAmount uint64
	RoutineSize   int
}

// 创建新桶
func NewBucket(bucketOptions BucketOptions) (b *Bucket) {
	b = new(Bucket)
	b.bucketOptions = bucketOptions
	b.chs = make(map[int]*Channel, bucketOptions.ChannelSize)
	b.routines = make([]chan *proto.PushRoomMsgRequest, bucketOptions.RoutineAmount)
	b.rooms = make(map[int]*Room, bucketOptions.RoomSize)
	for i := uint64(0); i < b.bucketOptions.RoutineAmount; i++ {
		c := make(chan *proto.PushRoomMsgRequest, bucketOptions.RoutineSize)
		b.routines[i] = c
		go b.PushRoom(c)
	}
	return
}

// 不断读取向房间内投送消息的请求，并投送到对应房间
func (b *Bucket) PushRoom(ch chan *proto.PushRoomMsgRequest) {
	for {
		var (
			arg  *proto.PushRoomMsgRequest
			room *Room
		)
		arg = <-ch
		if room = b.Room(arg.RoomId); room != nil {
			room.Push(&arg.Msg)
		}
	}
}

// 通过房间 roomId 获取 Room
func (b *Bucket) Room(rId int) (room *Room) {
	b.cLock.RLock()
	room, _ = b.rooms[rId]
	b.cLock.RUnlock()
	return
}

func (b *Bucket) Put(userId int, roomId int, ch *Channel) (err error) {
	var (
		room *Room
		has  bool
	)
	b.cLock.Lock()        // 获取锁
	if roomId != NoRoom { // NoRoom = -1
		if room, has = b.rooms[roomId]; !has { // 如果房间不存在
			room = NewRoom(roomId) // 创建新房间
			b.rooms[roomId] = room
		}
		ch.Room = room
	}
	ch.userId = userId
	b.chs[userId] = ch
	b.cLock.Unlock() // 释放锁

	if room != nil {
		err = room.Put(ch) // 将会话加入其所在房间的会话链表
	}
	return
}

func (b *Bucket) DeleteChannel(ch *Channel) {
	var (
		has  bool
		room *Room
	)
	b.cLock.RLock() // 加锁
	if ch, has = b.chs[ch.userId]; has {
		room = b.chs[ch.userId].Room
		//delete from bucket
		delete(b.chs, ch.userId) // 从 map 中删除元素，可以用原生的 delete 方法
	}
	if room != nil && room.DeleteChannel(ch) {
		// if room empty delete,will mark room.drop is true
		if room.drop == true { // 如果房间内没有会话了
			delete(b.rooms, room.Id) // 删除房间
		}
	}
	b.cLock.RUnlock() // 释放锁
}

// 通过用户 userId 获取用户会话对象 Channel
func (b *Bucket) Channel(userId int) (ch *Channel) {
	b.cLock.RLock() // 加锁
	ch = b.chs[userId]
	b.cLock.RUnlock() // 释放锁
	return
}

func (b *Bucket) BroadcastRoom(pushRoomMsgReq *proto.PushRoomMsgRequest) {
	num := atomic.AddUint64(&b.routinesNum, 1) % b.bucketOptions.RoutineAmount // 轮询通道
	b.routines[num] <- pushRoomMsgReq                                          // 将消息放入通道
}
