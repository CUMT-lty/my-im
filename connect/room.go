/**
 * Created by lock
 * Date: 2019-08-09
 * Time: 15:18
 */
package connect

import (
	"github.com/lty/my-go-chat/proto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sync"
)

const NoRoom = -1

type Room struct {
	Id          int          // 房间 id
	OnlineCount int          // 房间内的在线用户数量
	rLock       sync.RWMutex // 互斥锁
	drop        bool         // 房间是否存在 TODO: 已经弃用的房间怎么管理
	next        *Channel     // TODO: 房间内的会话链表？
}

func NewRoom(roomId int) *Room {
	room := new(Room)
	room.Id = roomId
	room.drop = false // 房间状态：在使用中
	room.next = nil   // 房间内的会话链表
	room.OnlineCount = 0
	return room
}

// 将一个会话加入该房间的会话链表
func (r *Room) Put(ch *Channel) (err error) {
	//doubly linked list
	r.rLock.Lock()         // 加锁
	defer r.rLock.Unlock() // 退出时释放锁
	if !r.drop {           // 如果房间还在使用
		// 头插法将一个会话连接插入双向链表中
		if r.next != nil { // 如果房间中已有会话
			r.next.Prev = ch
		}
		ch.Next = r.next
		ch.Prev = nil
		r.next = ch
		r.OnlineCount++ // 该房间中的在线用户数 +1
	} else {
		err = errors.New("room drop")
	}
	return
}

// 投送消息
func (r *Room) Push(msg *proto.Msg) {
	r.rLock.RLock()                             // 加锁
	for ch := r.next; ch != nil; ch = ch.Next { // 遍历房间会话链表
		if err := ch.Push(msg); err != nil { // 向每个会话投送消息
			logrus.Infof("push msg to channel in room err:%s", err.Error())
		}
	}
	r.rLock.RUnlock() // TODO: 这个方法和 Unlock 区别在哪？
	return
}

// 删除房间内的某个会话连接
func (r *Room) DeleteChannel(ch *Channel) bool {
	r.rLock.RLock()
	if ch.Next != nil { // 如果要删除的不是链表尾结点
		ch.Next.Prev = ch.Prev
	}
	if ch.Prev != nil { // 如果要删除的不是链表头结点
		ch.Prev.Next = ch.Next
	} else {
		r.next = ch.Next
	}
	r.OnlineCount-- // 房间内的在线用户数 -1
	r.drop = false
	if r.OnlineCount <= 0 { // 如果房间内的在线用户数 <=0，丢弃该房间
		r.drop = true
	}
	r.rLock.RUnlock() // TODO: 看一下这个释放锁方法
	return r.drop
}
