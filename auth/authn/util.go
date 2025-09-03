package authn

import (
	"encoding/json"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// GetInt64 When you need to parse a certain field in Claims as int64, it will automatically parse it as int64 for you,
// regardless of whether the actual type is float64 or json.Number(if you pass in jwt.WithJSONNumber for Config)
func GetInt64(claims jwt.MapClaims, key string) (int64, error) {
	val, exists := claims[key]
	if !exists {
		return 0, fmt.Errorf("key %s not found", key)
	}

	num, ok := val.(json.Number)
	if ok {
		return num.Int64()
	}

	f, ok := val.(float64)
	if ok {
		return int64(f), nil
	}

	return 0, fmt.Errorf("value is not a number")
}
