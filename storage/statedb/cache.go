// Copyright 2020 The klaytn Authors
// This file is part of the klaytn library.
//
// The klaytn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The klaytn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the klaytn library. If not, see <http://www.gnu.org/licenses/>.

package statedb

import (
	"errors"
	"strings"

	"github.com/go-redis/redis/v7"
)

type TrieNodeCacheType string

// TrieNodeCacheConfig contains configuration values of all TrieNodeCache.
type TrieNodeCacheConfig struct {
	CacheType                 TrieNodeCacheType
	LocalCacheSizeMB          int      // Memory allowance (MB) to use for caching trie nodes in fast cache
	FastCacheFileDir          string   // Directory where the persistent fastcache data is stored
	RedisEndpoints            []string // Endpoints of redis cache
	RedisClusterEnable        bool     // Enable cluster-enabled mode of redis cache
	RedisPublishBlockEnable   bool     // Enable publishing every inserted block to the redis server
	RedisSubscribeBlockEnable bool     // Enable subscribing blocks from the redis server
}

//go:generate mockgen -destination=storage/statedb/mocks/trie_node_cache_mock.go github.com/klaytn/klaytn/storage/statedb TrieNodeCache
// TrieNodeCache interface the cache of stateDB
type TrieNodeCache interface {
	Set(k, v []byte)
	Get(k []byte) []byte
	Has(k []byte) ([]byte, bool)
	UpdateStats() interface{}
	SaveToFile(filePath string, concurrency int) error
	Close() error
}

type BlockPubSub interface {
	PublishBlock(msg string) error
	SubscribeBlockCh() <-chan *redis.Message
	UnsubscribeBlock() error
}

const (
	// Available trie node cache types
	CacheTypeLocal  TrieNodeCacheType = "LocalCache"
	CacheTypeRedis                    = "RemoteCache"
	CacheTypeHybrid                   = "HybridCache"
)

var (
	errNotSupportedCacheType = errors.New("not supported stateDB TrieNodeCache type")
)

func (cacheType TrieNodeCacheType) ToValid() TrieNodeCacheType {
	validTrieNodeCacheTypes := []TrieNodeCacheType{CacheTypeLocal, CacheTypeRedis, CacheTypeHybrid}
	for _, validType := range validTrieNodeCacheTypes {
		if strings.ToLower(string(cacheType)) == strings.ToLower(string(validType)) {
			return validType
		}
	}
	logger.Warn("Invalid trie node cache type", "inputType", cacheType, "validTypes", validTrieNodeCacheTypes)
	return ""
}

// NewTrieNodeCache creates one type of any supported trie node caches.
// NOTE: It returns (nil, nil) when the cache type is CacheTypeLocal and its size is set to zero.
func NewTrieNodeCache(config TrieNodeCacheConfig) (TrieNodeCache, error) {
	switch config.CacheType {
	case CacheTypeLocal:
		return NewFastCache(config), nil
	case CacheTypeRedis:
		return NewRedisCache(config)
	case CacheTypeHybrid:
		logger.Info("Set hybrid trie node cache using both of localCache (fastCache) and redisCache")
		return NewHybridCache(config)
	default:
	}
	logger.Error("Invalid trie node cache type", "cacheType", config.CacheType)
	return nil, errNotSupportedCacheType
}

func GetEmptyTrieNodeCacheConfig() TrieNodeCacheConfig {
	return TrieNodeCacheConfig{
		CacheType:          CacheTypeLocal,
		LocalCacheSizeMB:   0,
		FastCacheFileDir:   "",
		RedisEndpoints:     nil,
		RedisClusterEnable: false,
	}
}
