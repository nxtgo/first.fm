package lastfm

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/nxtgo/gce"
)

type Cache struct {
	User *gce.Cache[string, User]
	// todo: more cache
}

func NewCache() *Cache {
	return &Cache{
		User: gce.New[string, User](
			gce.WithDefaultTTL(30*time.Minute),
			gce.WithMaxEntries(10_000),
		),
	}
}

type CacheStats struct {
	Name  string
	Stats gce.Stats
}

func (c *Cache) Stats() []CacheStats {
	return []CacheStats{
		{"User", c.User.Stats()},
	}
}

func (c *Cache) Close() {
	c.User.Close()
}

func GenerateCacheKey(method string, args P) string {
	if len(args) == 0 {
		return method
	}

	keys := make([]string, 0, len(args))
	for k := range args {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString(method)
	for _, k := range keys {
		sb.WriteByte('|')
		sb.WriteString(k)
		sb.WriteByte(':')
		fmt.Fprint(&sb, args[k])
	}

	hash := sha256.Sum256([]byte(sb.String()))

	return fmt.Sprintf("%s|%x", method, hash[:16])
}
