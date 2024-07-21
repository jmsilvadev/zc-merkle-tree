package leveldb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHappyPath(t *testing.T) {
	db, err := New("/tmp/test.db")
	require.NoError(t, err)

	key := "test"
	t.Run("Put", func(t *testing.T) {
		err := db.Put(key, []byte(key))
		require.NoError(t, err)
	})

	t.Run("Get", func(t *testing.T) {
		r, err := db.Get(key)
		require.NoError(t, err)
		require.Equal(t, []byte(key), r)
	})

	t.Run("Delete", func(t *testing.T) {
		err := db.Delete(key)
		require.NoError(t, err)
	})

	t.Run("DeleteByPrefix", func(t *testing.T) {
		err := db.DeleteByPrefix(key)
		require.NoError(t, err)
	})

	t.Run("GetByPrefix", func(t *testing.T) {
		prefix := "prefix_"
		keys := []string{"prefix_key1", "prefix_key2", "prefix_key3"}
		values := []string{"value1", "value2", "value3"}

		for i, k := range keys {
			err := db.Put(k, []byte(values[i]))
			require.NoError(t, err)
		}

		result, err := db.GetByPrefix(prefix)
		require.NoError(t, err)
		require.Len(t, result, len(keys))

		for _, k := range keys {
			require.Contains(t, result, k)
		}
	})

	err = db.Close()
	require.NoError(t, err)

}
