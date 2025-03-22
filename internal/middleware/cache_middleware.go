package middleware

import "go.uber.org/fx"

type CacheMiddleware struct {
	fx.In

	Storage StorageMiddleware
	cache   CacheStrategy
}

func NewCacheMiddleware(storage StorageMiddleware, strategy CacheStrategy) *CacheMiddleware {
	return &CacheMiddleware{
		Storage: storage,
		cache:   strategy,
	}
}

func (cm *CacheMiddleware) Set(key string, value string) error {
	if err := cm.Storage.Set(key, value); err != nil {
		return err
	}
	cm.cache.Put(key, value)
	return nil
}

func (cm *CacheMiddleware) Get(key string) (string, error) {
	// Try cache first
	if value, exists := cm.cache.Get(key); exists {
		return value, nil
	}

	// If not in cache, get from storage
	value, err := cm.Storage.Get(key)
	if err != nil {
		return "", err
	}

	// Add to cache
	cm.cache.Put(key, value)
	return value, nil
}

func (cm *CacheMiddleware) Delete(key string) error {
	if err := cm.Storage.Delete(key); err != nil {
		return err
	}
	cm.cache.Remove(key)
	return nil
}
