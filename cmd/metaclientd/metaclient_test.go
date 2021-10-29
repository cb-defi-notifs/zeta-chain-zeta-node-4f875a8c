// +build integration

// this is integration test; must be run when a chain is running:
// starport chain serve

package metaclientd

import (
	"github.com/rs/zerolog/log"
	. "gopkg.in/check.v1"
	"os"
	"path/filepath"
)

type MySuite struct {
	bridge *MetachainBridge
}

var _ = Suite(&MySuite{})

func (s *MySuite) SetUpTest(c *C) {
	SetupConfigForTest() // setup meta-prefix

	c.Logf("Settting up test...")
	homeDir, err := os.UserHomeDir()
	c.Logf("user home dir: %s", homeDir)
	chainHomeFoler := filepath.Join(homeDir, ".metacore")
	c.Logf("chain home dir: %s", chainHomeFoler)

	// alice is the default user created by Starport chain serve
	signerName := "alice"
	signerPass := "password"
	kb, _, err := GetKeyringKeybase(chainHomeFoler, signerName, signerPass)
	if err != nil {
		log.Fatal().Err(err).Msg("fail to get keyring keybase")
	}

	k := NewKeysWithKeybase(kb, signerName, signerPass)
	//log.Info().Msgf("keybase: %s", k.GetSignerInfo().GetAddress())

	chainIP := "127.0.0.1"
	bridge, err := NewMetachainBridge(k, chainIP, "alice")
	if err != nil {
		c.Fail()
	}
	s.bridge = bridge
}

func (s *MySuite) TestGetBlockHeight(c *C) {
	h, err := s.bridge.GetBlockHeight()
	c.Assert(err, IsNil)
	c.Logf("height %d", h)
}

func (s *MySuite) TestGetAccountNumberAndSeuqeuence(c *C) {
	an, as, err := s.bridge.GetAccountNumberAndSequenceNumber()
	c.Assert(err, IsNil)
	c.Logf("acc number %d acc sequence %d", an, as)
}
