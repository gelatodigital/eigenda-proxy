package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Layr-Labs/eigenda-proxy/metrics"
	"github.com/Layr-Labs/eigenda-proxy/server"
	"github.com/Layr-Labs/eigenda/api/clients"
	"github.com/ethereum/go-ethereum/log"

	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"

	"github.com/stretchr/testify/require"
)

const (
	privateKey = "SIGNER_PRIVATE_KEY"
	transport  = "http"
	svcName    = "eigenda_proxy"
	host       = "127.0.0.1"
	holeskyDA  = "disperser-holesky.eigenda.xyz:443"
)

type TestSuite struct {
	Ctx    context.Context
	Log    log.Logger
	Server *server.Server
}

func CreateTestSuite(t *testing.T, useMemory bool) (TestSuite, func()) {
	ctx := context.Background()

	// load signer key from environment
	pk := os.Getenv(privateKey)
	if pk == "" && !useMemory {
		t.Fatal("SIGNER_PRIVATE_KEY environment variable not set")
	}

	log := oplog.NewLogger(os.Stdout, oplog.CLIConfig{
		Level:  log.LevelDebug,
		Format: oplog.FormatLogFmt,
		Color:  true,
	}).New("role", svcName)

	eigendaCfg := server.Config{
		ClientConfig: clients.EigenDAClientConfig{
			RPC:                      holeskyDA,
			StatusQueryTimeout:       time.Minute * 45,
			StatusQueryRetryInterval: time.Second * 1,
			DisableTLS:               false,
			SignerPrivateKeyHex:      pk,
		},
		CacheDir:               "../test/resources/SRSTables/",
		G1Path:                 "../operator-setup/resources/g1_abbr.point",
		MaxBlobLength:          "90kib",
		G2PowerOfTauPath:       "../operator-setup/resources/kzg/g2_abbr.point.powerOf2",
		PutBlobEncodingVersion: 0x00,
		MemstoreEnabled:        useMemory,
		MemstoreBlobExpiration: 14 * 24 * time.Hour,
	}

	store, err := server.LoadStore(
		server.CLIConfig{
			EigenDAConfig: eigendaCfg,
			MetricsCfg:    opmetrics.CLIConfig{},
		},
		ctx,
		log,
	)
	require.NoError(t, err)
	server := server.NewServer(host, 0, store, log, metrics.NoopMetrics)

	t.Log("Starting proxy server...")
	err = server.Start()
	require.NoError(t, err)

	kill := func() {
		if err := server.Stop(); err != nil {
			panic(err)
		}
	}

	return TestSuite{
		Ctx:    ctx,
		Log:    log,
		Server: server,
	}, kill
}

func (ts *TestSuite) Address() string {
	// read port from listener
	port := ts.Server.Port()

	return fmt.Sprintf("%s://%s:%d", transport, host, port)
}
