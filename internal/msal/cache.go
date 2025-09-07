package msal

import (
	"context"
	"os"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
)

// MSAL cache accessor
type CacheAccessor struct {
	file string
}

// constructor for CacheAccessor
func NewCacheAccessor(file string) *CacheAccessor {
	return &CacheAccessor{file: file}
}

// MSAL cache accessor interface method
func (c *CacheAccessor) Export(ctx context.Context, marshaller cache.Marshaler, hints cache.ExportHints) error {
	data, err := marshaller.Marshal()
	if err != nil {
		return err
	}
	return os.WriteFile(c.file, data, 0600)
}

// MSAL cache accessor interface method
func (c *CacheAccessor) Replace(ctx context.Context, unmarshaler cache.Unmarshaler, hints cache.ReplaceHints) error {
	data, err := os.ReadFile(c.file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return unmarshaler.Unmarshal(data)
}
