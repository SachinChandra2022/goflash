package repository

import (
	"errors"
	"github.com/sachinchandra/goflash/internal/database"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var requestScript = redis.NewScript(`
	local current_stock = tonumber(redis.call("get", KEYS[1]))
	if current_stock <= 0 then
		return -1
	end
	redis.call("decr", KEYS[1])
	return current_stock - 1
`)

func InitializeRedisSync() {
	err := database.Rdb.Set(database.Ctx, "product:1:quantity", 100, 0).Err()
	if err != nil {
		panic(err)
	}
}

func PurchaseProductRedis(productID int) error {
	key := "product:1:quantity" 

	val, err := database.Rdb.Get(database.Ctx, key).Result()
    if err == redis.Nil {
        fmt.Println("DEBUG: Redis Key does not exist!")
    } else if err != nil {
        fmt.Println("DEBUG: Redis Error:", err)
    } else {
        fmt.Printf("DEBUG: Current Redis Stock: %s\n", val)
    }

	result, err := requestScript.Run(database.Ctx, database.Rdb, []string{key}).Int()
	
	if err != nil {
		return err
	}

	if result == -1 {
		return errors.New("sold out")
	}
	return nil
}