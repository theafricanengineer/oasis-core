package checkpoint

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/golang/snappy"
	"github.com/stretchr/testify/require"

	"github.com/oasislabs/oasis-core/go/common"
	"github.com/oasislabs/oasis-core/go/common/cbor"
	"github.com/oasislabs/oasis-core/go/common/crypto/hash"
	"github.com/oasislabs/oasis-core/go/storage/mkvs"
	db "github.com/oasislabs/oasis-core/go/storage/mkvs/db/api"
	badgerDb "github.com/oasislabs/oasis-core/go/storage/mkvs/db/badger"
	"github.com/oasislabs/oasis-core/go/storage/mkvs/node"
)

var testNs = common.NewTestNamespaceFromSeed([]byte("oasis mkvs checkpoint test ns"), 0)

func TestFileCheckpointCreator(t *testing.T) {
	require := require.New(t)

	// Generate some data.
	dir, err := ioutil.TempDir("", "mkvs.checkpoint")
	require.NoError(err, "TempDir")
	defer os.RemoveAll(dir)

	ndb, err := badgerDb.New(&db.Config{
		DB:           filepath.Join(dir, "db"),
		Namespace:    testNs,
		MaxCacheSize: 16 * 1024 * 1024,
	})
	require.NoError(err, "New")

	ctx := context.Background()
	tree := mkvs.New(nil, ndb)
	for i := 0; i < 1000; i++ {
		err = tree.Insert(ctx, []byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
		require.NoError(err, "Insert")
	}

	_, rootHash, err := tree.Commit(ctx, testNs, 0)
	require.NoError(err, "Commit")
	root := node.Root{
		Namespace: testNs,
		Version:   0,
		Hash:      rootHash,
	}

	// Create a file-based checkpoint creator.
	fc, err := NewFileCreator(filepath.Join(dir, "checkpoints"), ndb)
	require.NoError(err, "NewFileCreator")

	// There should be no checkpoints before one is created.
	cps, err := fc.GetCheckpoints(ctx, &GetCheckpointsRequest{})
	require.NoError(err, "GetCheckpoints")
	require.Len(cps, 0)

	_, err = fc.GetCheckpoint(ctx, &GetCheckpointRequest{Root: root})
	require.Error(err, "GetCheckpoint should fail with non-existent checkpoint")

	// Create a checkpoint and check that it has been created correctly.
	cp, err := fc.CreateCheckpoint(ctx, root, 16*1024)
	require.NoError(err, "CreateCheckpoint")
	require.EqualValues(1, cp.Version, "version should be correct")
	require.EqualValues(root, cp.Root, "checkpoint root should be correct")
	require.Len(cp.Chunks, 3, "there should be the correct number of chunks")

	var expectedChunks []hash.Hash
	for _, hh := range []string{
		"620f318c245858b351602a8b21e708663b03cd2befd982210ffbaa3c56bf9358",
		"37d38a95492038df4afee65b5b91bf8499f522dea346c0e553e03d3333bff394",
		"df608e9821dc8d248a0f0d0ff4cb998d51f329767dfc8a7520c15e38c47be1e9",
	} {
		var h hash.Hash
		_ = h.UnmarshalHex(hh)
		expectedChunks = append(expectedChunks, h)
	}
	require.EqualValues(expectedChunks, cp.Chunks, "chunk hashes should be correct")

	// There should now be one checkpoint.
	cps, err = fc.GetCheckpoints(ctx, &GetCheckpointsRequest{Version: 1})
	require.NoError(err, "GetCheckpoints")
	require.Len(cps, 1, "there should be one checkpoint")
	require.Equal(cp, cps[0], "checkpoint returned by GetCheckpoint should be correct")

	gcp, err := fc.GetCheckpoint(ctx, &GetCheckpointRequest{Version: 1, Root: root})
	require.NoError(err, "GetCheckpoint")
	require.Equal(cp, gcp)

	// Try re-creating the same checkpoint again and make sure we get the same metadata.
	existingCp, err := fc.CreateCheckpoint(ctx, root, 16*1024)
	require.NoError(err, "CreateCheckpoint on an existing root should work")
	require.Equal(cp, existingCp, "created checkpoint should be correct")

	// We should be able to retrieve chunks.
	_, err = cp.GetChunkMetadata(999)
	require.Error(err, "GetChunkMetadata should fail for unknown chunk")
	chunk0, err := cp.GetChunkMetadata(0)
	require.NoError(err, "GetChunkMetadata")

	var buf bytes.Buffer
	err = fc.GetCheckpointChunk(ctx, chunk0, &buf)
	require.NoError(err, "GetChunk should work")

	// Fetching a non-existent chunk should fail.
	invalidChunk := *chunk0
	invalidChunk.Index = 999
	err = fc.GetCheckpointChunk(ctx, &invalidChunk, &buf)
	require.Error(err, "GetChunk on a non-existent chunk should fail")

	// Create a fresh node database to restore into.
	ndb2, err := badgerDb.New(&db.Config{
		DB:           filepath.Join(dir, "db2"),
		Namespace:    testNs,
		MaxCacheSize: 16 * 1024 * 1024,
	})
	require.NoError(err, "New")

	// Try to restore some chunks.
	rs, err := NewRestorer(ndb2)
	require.NoError(err, "NewRestorer")

	_, err = rs.RestoreChunk(ctx, 0, &buf)
	require.Error(err, "RestoreChunk should fail when no restore is in progress")
	require.True(errors.Is(err, ErrNoRestoreInProgress))

	// Generate a bogus manifest which does not verify by corrupting chunk at index 1.
	bogusCp, err := fc.GetCheckpoint(ctx, &GetCheckpointRequest{Version: 1, Root: root})
	require.NoError(err, "GetCheckpoint")
	require.Equal(cp, bogusCp)

	buf.Reset()
	sw := snappy.NewBufferedWriter(&buf)
	enc := cbor.NewEncoder(sw)
	_ = enc.Encode([]byte("this chunk is bogus"))
	sw.Close()

	bogusChunk := make([]byte, buf.Len())
	copy(bogusChunk, buf.Bytes())
	// Make sure that the chunk integrity is correct.
	bogusCp.Chunks[1].FromBytes(bogusChunk)

	err = rs.StartRestore(ctx, bogusCp)
	require.NoError(err, "StartRestore")
	for i := 0; i < len(bogusCp.Chunks); i++ {
		var cm *ChunkMetadata
		cm, err = cp.GetChunkMetadata(uint64(i))
		require.NoError(err, "GetChunkMetadata")

		buf.Reset()
		if i == 1 {
			// Substitute the bogus chunk.
			_, _ = buf.Write(bogusChunk)
		} else {
			err = fc.GetCheckpointChunk(ctx, cm, &buf)
			require.NoError(err, "GetChunk")
		}
		var done bool
		done, err = rs.RestoreChunk(ctx, uint64(i), &buf)
		require.False(done, "RestoreChunk should not signal completed restoration")
		if i == 1 {
			require.Error(err, "RestoreChunk should fail with bogus chunk")
			require.True(errors.Is(err, ErrChunkProofVerificationFailed))
			// Restorer should be reset.
			break
		} else {
			require.NoError(err, "RestoreChunk")
		}
	}

	// Try to correctly restore.
	err = rs.StartRestore(ctx, cp)
	require.NoError(err, "StartRestore")
	err = rs.StartRestore(ctx, cp)
	require.Error(err, "StartRestore should fail when a restore is already in progress")
	require.True(errors.Is(err, ErrRestoreAlreadyInProgress))
	rcp := rs.GetCurrentCheckpoint()
	require.EqualValues(rcp, cp, "GetCurrentCheckpoint should return the checkpoint being restored")
	require.NotSame(rcp, cp, "GetCurrentCheckpoint should return a copy")
	for i := 0; i < len(cp.Chunks); i++ {
		var cm *ChunkMetadata
		cm, err = cp.GetChunkMetadata(uint64(i))
		require.NoError(err, "GetChunkMetadata")

		// Try with a corrupted chunk first.
		buf.Reset()
		_, _ = buf.Write([]byte("corrupted chunk"))
		_, err = rs.RestoreChunk(ctx, uint64(i), &buf)
		require.Error(err, "RestoreChunk should fail with corrupted chunk")

		buf.Reset()
		err = fc.GetCheckpointChunk(ctx, cm, &buf)
		require.NoError(err, "GetChunk")

		var done bool
		done, err = rs.RestoreChunk(ctx, uint64(i), &buf)
		require.NoError(err, "RestoreChunk")

		if i == len(cp.Chunks)-1 {
			require.True(done, "RestoreChunk should signal completed restoration when done")
		} else {
			require.False(done, "RestoreChunk should not signal completed restoration early")

			_, err = rs.RestoreChunk(ctx, uint64(i), &buf)
			require.Error(err, "RestoreChunk should fail if the same chunk has already been restored")
			require.True(errors.Is(err, ErrChunkAlreadyRestored))
		}
	}
	err = ndb2.Finalize(ctx, root.Version, []hash.Hash{root.Hash})
	require.NoError(err, "Finalize")

	// Verify that everything has been restored.
	tree = mkvs.NewWithRoot(nil, ndb2, root)
	for i := 0; i < 1000; i++ {
		var value []byte
		value, err = tree.Get(ctx, []byte(strconv.Itoa(i)))
		require.NoError(err, "Get")
		require.Equal([]byte(strconv.Itoa(i)), value)
	}

	// Deleting a checkpoint should work.
	err = fc.DeleteCheckpoint(ctx, &DeleteCheckpointRequest{Version: 1, Root: root})
	require.NoError(err, "DeleteCheckpoint")

	// There should now be no checkpoints.
	cps, err = fc.GetCheckpoints(ctx, &GetCheckpointsRequest{Version: 1})
	require.NoError(err, "GetCheckpoints")
	require.Len(cps, 0, "there should be no checkpoints")

	_, err = fc.GetCheckpoint(ctx, &GetCheckpointRequest{Version: 1, Root: root})
	require.Error(err, "GetCheckpoint should fail with non-existent checkpoint")

	// Deleting a non-existent checkpoint should fail.
	err = fc.DeleteCheckpoint(ctx, &DeleteCheckpointRequest{Version: 1, Root: root})
	require.Error(err, "DeleteCheckpoint on a non-existent checkpoint should fail")

	// Fetching a non-existent chunk should fail.
	err = fc.GetCheckpointChunk(ctx, chunk0, &buf)
	require.Error(err, "GetChunk on a non-existent chunk should fail")

	// Create a checkpoint with unknown root.
	invalidRoot := root
	invalidRoot.Hash.FromBytes([]byte("mkvs checkpoint test invalid root"))
	_, err = fc.CreateCheckpoint(ctx, invalidRoot, 16*1024)
	require.Error(err, "CreateCheckpoint should fail for invalid root")
}
