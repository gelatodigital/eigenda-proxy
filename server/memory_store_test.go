package server

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/Layr-Labs/eigenda-proxy/verify"
	"github.com/Layr-Labs/eigenda/encoding/kzg"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"
)

const (
	testPreimage = "Four score and seven years ago"
)

func TestGetSet(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kzgConfig := &kzg.KzgConfig{
		G1Path:          "../operator-setup/resources/g1_abbr.point",
		G2PowerOf2Path:  "../operator-setup/resources/g2_abbr.point.powerOf2",
		CacheDir:        "../test/resources/SRSTables/",
		SRSOrder:        3000,
		SRSNumberToLoad: 3000,
		NumWorker:       uint64(runtime.GOMAXPROCS(0)),
	}
	verifier, err := verify.NewVerifier(kzgConfig)
	assert.NoError(t, err)

	ms, err := NewMemStore(
		ctx,
		verifier,
		log.New(),
		1024*1024*2,
		time.Hour*1000,
	)

	assert.NoError(t, err)

	expected := []byte(testPreimage)
	key, err := ms.Put(ctx, expected)
	assert.NoError(t, err)

	actual, err := ms.Get(ctx, key, BinaryDomain)
	assert.NoError(t, err)
	assert.Equal(t, actual, expected)
}

func TestExpiration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kzgConfig := &kzg.KzgConfig{
		G1Path:          "../operator-setup/resources/g1_abbr.point",
		G2PowerOf2Path:  "../operator-setup/resources/g2_abbr.point.powerOf2",
		CacheDir:        "../test/resources/SRSTables/",
		SRSOrder:        3000,
		SRSNumberToLoad: 3000,
		NumWorker:       uint64(runtime.GOMAXPROCS(0)),
	}
	verifier, err := verify.NewVerifier(kzgConfig)
	assert.NoError(t, err)

	ms, err := NewMemStore(
		ctx,
		verifier,
		log.New(),
		1024*1024*2,
		time.Millisecond*10,
	)

	assert.NoError(t, err)

	preimage := []byte(testPreimage)
	key, err := ms.Put(ctx, preimage)
	assert.NoError(t, err)

	// sleep 1 second and verify that older blob entries are removed
	time.Sleep(time.Second * 1)

	_, err = ms.Get(ctx, key, BinaryDomain)
	assert.Error(t, err)

}
