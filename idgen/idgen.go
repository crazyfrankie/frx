package idgen

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	counterKeyExpirationTime = 10 * time.Minute
	maxCounterPosition       = 255
)

type IDGenerator interface {
	GenID(ctx context.Context) (int64, error)
	GenMultiIDs(ctx context.Context, counts int) ([]int64, error) // suggest batch size <= 200
}

func New(client redis.Cmdable) (IDGenerator, error) {
	// Initialization code.
	return &idGenImpl{
		cli: client,
	}, nil
}

type idGenImpl struct {
	cli       redis.Cmdable
	namespace string
}

func (i *idGenImpl) GenID(ctx context.Context) (int64, error) {
	ids, err := i.GenMultiIDs(ctx, 1)
	if err != nil {
		return 0, err
	}

	return ids[0], nil
}

func (i *idGenImpl) GenMultiIDs(ctx context.Context, counts int) ([]int64, error) {
	const maxTimeAddrTimes = 8

	leftNum := int64(counts)
	lastMs := int64(0)
	ids := make([]int64, 0, counts)
	svrID := int64(0) // A server id is all 0.

	for idx := int64(0); leftNum > 0 && idx < maxTimeAddrTimes; idx++ {
		ms := maxInt64(i.GetIDTimeMs(), lastMs)
		if ms <= lastMs {
			ms++
		}
		lastMs = ms

		redisKey := genIDKey(i.namespace, svrID, ms)

		counterPosition, err := i.IncrBy(ctx, redisKey, leftNum)
		if err != nil {
			return nil, err
		}

		var start, end int64
		start = counterPosition - leftNum

		if start == 0 {
			i.Expire(ctx, redisKey)
		}

		if start > maxCounterPosition {
			continue
		} else if counterPosition < leftNum {
			return nil, fmt.Errorf("recycling of counting space occurs, ms=%v", ms)
		}

		if counterPosition > maxCounterPosition {
			end = maxCounterPosition + 1
			leftNum = counterPosition - maxCounterPosition - 1
		} else {
			end = counterPosition
			leftNum = 0
		}

		seconds := ms / 1000
		millis := ms % 1000

		if seconds&0xFFFFFFFF != seconds {
			return nil, fmt.Errorf("seconds more than 32 bits, seconds=%v", seconds)
		}

		if svrID&0x3FFF != svrID {
			return nil, fmt.Errorf("server id more than 14 bits, serverID=%v", svrID)
		}

		for i := start; i < end; i++ {
			// fmt.Printf("sec=%v, ms=%v, counter=%v\n", seconds, millis, i)
			id := (seconds)<<32 + (millis)<<22 + i<<14 + svrID
			ids = append(ids, id)
		}
	}

	if len(ids) < counts || leftNum != 0 {
		return nil, fmt.Errorf("IDs num not enough, ns=%v, expect=%v, gotten=%v, lastMs=%v", i.namespace, counts, len(ids), lastMs)
	}

	return ids, nil
}

func (i *idGenImpl) IncrBy(ctx context.Context, key string, num int64) (cntPos int64, err error) {
	return i.cli.IncrBy(ctx, key, num).Result()
}

func (i *idGenImpl) GetIDTimeMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func (i *idGenImpl) Expire(ctx context.Context, key string) {
	// Temporarily ignore errors
	_, _ = i.cli.Expire(ctx, key, counterKeyExpirationTime).Result()
}

func genIDKey(space string, svrID int64, ms int64) string {
	// Once the format of this key is determined, it cannot be changed
	return fmt.Sprintf("id_generator:%v:%v:%v", space, svrID, ms)
}

func maxInt64(a, b int64) int64 {
	if a <= b {
		return b
	} else {
		return a
	}
}
