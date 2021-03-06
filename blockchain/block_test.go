// Copyright (c) 2018 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package blockchain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/facebookgo/clock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/blake2b"

	"github.com/iotexproject/iotex-core/blockchain/action"
	"github.com/iotexproject/iotex-core/config"
	"github.com/iotexproject/iotex-core/crypto"
	"github.com/iotexproject/iotex-core/pkg/hash"
	"github.com/iotexproject/iotex-core/pkg/version"
	"github.com/iotexproject/iotex-core/proto"
	"github.com/iotexproject/iotex-core/state"
	ta "github.com/iotexproject/iotex-core/test/testaddress"
	"github.com/iotexproject/iotex-core/testutil"
)

func TestBasicHash(t *testing.T) {
	require := require.New(t)

	// basic hash test
	input := []byte("hello")
	hash := sha256.Sum256(input)
	hash = sha256.Sum256(hash[:])
	hello, _ := hex.DecodeString("9595c9df90075148eb06860365df33584b75bff782a510c6cd4883a419833d50")
	require.Equal(hello, hash[:])
	t.Logf("sha256(sha256(\"hello\") = %x", hash)

	hash = blake2b.Sum256(input)
	hash = blake2b.Sum256(hash[:])
	hello, _ = hex.DecodeString("901c60ffffd77f743729f8fea0233c0b00223428b5192c2015f853562b45ce59")
	require.Equal(hello, hash[:])
	t.Logf("blake2b(blake2b(\"hello\") = %x", hash)
}

func TestMerkle(t *testing.T) {
	require := require.New(t)

	amount := uint64(50 << 22)
	// create testing transactions
	cbtsf0 := action.NewCoinBaseTransfer(big.NewInt(int64(amount)), ta.Addrinfo["producer"].RawAddress)
	require.NotNil(cbtsf0)
	cbtsf1 := action.NewCoinBaseTransfer(big.NewInt(int64(amount)), ta.Addrinfo["alfa"].RawAddress)
	require.NotNil(cbtsf1)
	cbtsf2 := action.NewCoinBaseTransfer(big.NewInt(int64(amount)), ta.Addrinfo["bravo"].RawAddress)
	require.NotNil(cbtsf2)
	cbtsf3 := action.NewCoinBaseTransfer(big.NewInt(int64(amount)), ta.Addrinfo["charlie"].RawAddress)
	require.NotNil(cbtsf3)
	cbtsf4 := action.NewCoinBaseTransfer(big.NewInt(int64(amount)), ta.Addrinfo["echo"].RawAddress)
	require.NotNil(cbtsf4)

	// verify tx hash
	hash0, _ := hex.DecodeString("c42f754fdf676a6ac4cdccba96f2dc1055c41c25effc72ac9477e120712e5634")
	actual := cbtsf0.Hash()
	require.Equal(hash0, actual[:])
	t.Logf("actual hash = %x", actual[:])

	hash1, _ := hex.DecodeString("2c4bcfb59297b3e472f7c15ff31a3ed080b749a952c18bb585ef517542c8381d")
	actual = cbtsf1.Hash()
	require.Equal(hash1, actual[:])
	t.Logf("actual hash = %x", actual[:])

	hash2, _ := hex.DecodeString("46e07d8753a07d66f9b76797a0e3257fd2b70b019722dfb3394ba51db2b21b62")
	actual = cbtsf2.Hash()
	require.Equal(hash2, actual[:])
	t.Logf("actual hash = %x", actual[:])

	hash3, _ := hex.DecodeString("d300718263371fb0218a2616f8822866547dade0f0b1dbe3d326950c4488f6de")
	actual = cbtsf3.Hash()
	require.Equal(hash3, actual[:])
	t.Logf("actual hash = %x", actual[:])

	hash4, _ := hex.DecodeString("75b315ef2baaa13af4579876d018db0f512e132d3c4b41b5ebe9d0b75e9cf054")
	actual = cbtsf4.Hash()
	require.Equal(hash4, actual[:])
	t.Logf("actual hash = %x", actual[:])

	// manually compute merkle root
	cat := append(hash0, hash1...)
	hash01 := blake2b.Sum256(cat)
	t.Logf("hash01 = %x", hash01)

	cat = append(hash2, hash3...)
	hash23 := blake2b.Sum256(cat)
	t.Logf("hash23 = %x", hash23)

	cat = append(hash4, hash4...)
	hash45 := blake2b.Sum256(cat)
	t.Logf("hash45 = %x", hash45)

	cat = append(hash01[:], hash23[:]...)
	hash03 := blake2b.Sum256(cat)
	t.Logf("hash03 = %x", hash03)

	cat = append(hash45[:], hash45[:]...)
	hash47 := blake2b.Sum256(cat)
	t.Logf("hash47 = %x", hash47)

	cat = append(hash03[:], hash47[:]...)
	hash07 := blake2b.Sum256(cat)
	t.Logf("hash07 = %x", hash07)

	// create block using above 5 tx and verify merkle
	block := NewBlock(
		0,
		0,
		hash.ZeroHash32B,
		clock.New(),
		[]*action.Transfer{cbtsf0, cbtsf1, cbtsf2, cbtsf3, cbtsf4},
		nil,
		nil,
	)
	hash := block.TxRoot()
	require.Equal(hash07[:], hash[:])

	t.Log("Merkle root match pass\n")
}

func TestConvertFromBlockPb(t *testing.T) {
	blk := Block{}
	blk.ConvertFromBlockPb(&iproto.BlockPb{
		Header: &iproto.BlockHeaderPb{
			Version: version.ProtocolVersion,
			Height:  123456789,
		},
		Actions: []*iproto.ActionPb{
			{Action: &iproto.ActionPb_Transfer{
				Transfer: &iproto.TransferPb{},
			},
				Version: version.ProtocolVersion,
				Nonce:   101,
			},
			{Action: &iproto.ActionPb_Transfer{
				Transfer: &iproto.TransferPb{},
			},
				Version: version.ProtocolVersion,
				Nonce:   102,
			},
			{Action: &iproto.ActionPb_Vote{
				Vote: &iproto.VotePb{},
			},
				Version: version.ProtocolVersion,
				Nonce:   103,
			},
			{Action: &iproto.ActionPb_Vote{
				Vote: &iproto.VotePb{},
			},
				Version: version.ProtocolVersion,
				Nonce:   104,
			},
		},
	})

	blk.Header.txRoot = blk.TxRoot()

	raw, err := blk.Serialize()
	require.Nil(t, err)

	var newblk Block
	err = newblk.Deserialize(raw)
	require.Nil(t, err)

	blockBytes := blk.ByteStream()
	require.True(t, len(blockBytes) > 0)

	require.Equal(t, uint64(123456789), newblk.Header.height)

	require.Equal(t, uint64(101), newblk.Transfers[0].Nonce)
	require.Equal(t, uint64(102), newblk.Transfers[1].Nonce)

	require.Equal(t, uint64(103), newblk.Votes[0].Nonce)
	require.Equal(t, uint64(104), newblk.Votes[1].Nonce)
}

func TestWrongRootHash(t *testing.T) {
	require := require.New(t)
	val := validator{nil}
	tsf1, err := action.NewTransfer(1, big.NewInt(20), ta.Addrinfo["producer"].RawAddress, ta.Addrinfo["alfa"].RawAddress, []byte{}, uint64(100000), big.NewInt(10))
	require.NoError(err)
	tsf1, err = tsf1.Sign(ta.Addrinfo["producer"])
	require.Nil(err)
	tsf2, err := action.NewTransfer(1, big.NewInt(30), ta.Addrinfo["producer"].RawAddress, ta.Addrinfo["bravo"].RawAddress, []byte{}, uint64(100000), big.NewInt(10))
	require.NoError(err)
	tsf2, err = tsf2.Sign(ta.Addrinfo["producer"])
	require.Nil(err)
	hash := tsf1.Hash()
	blk := NewBlock(1, 1, hash, clock.New(), []*action.Transfer{tsf1, tsf2}, nil, nil)
	blk.Header.Pubkey = ta.Addrinfo["producer"].PublicKey
	blkHash := blk.HashBlock()
	blk.Header.blockSig = crypto.EC283.Sign(ta.Addrinfo["producer"].PrivateKey, blkHash[:])
	require.Nil(val.Validate(blk, 0, hash))
	blk.Transfers[0], blk.Transfers[1] = blk.Transfers[1], blk.Transfers[0]
	require.NotNil(val.Validate(blk, 0, hash))
}

func TestSignBlock(t *testing.T) {
	require := require.New(t)
	val := validator{nil}
	tsf1, err := action.NewTransfer(1, big.NewInt(20), ta.Addrinfo["producer"].RawAddress, ta.Addrinfo["alfa"].RawAddress, []byte{}, uint64(100000), big.NewInt(10))
	require.NoError(err)
	tsf1, err = tsf1.Sign(ta.Addrinfo["producer"])
	require.Nil(err)
	tsf2, err := action.NewTransfer(1, big.NewInt(30), ta.Addrinfo["producer"].RawAddress, ta.Addrinfo["bravo"].RawAddress, []byte{}, uint64(100000), big.NewInt(10))
	require.NoError(err)
	tsf2, err = tsf2.Sign(ta.Addrinfo["producer"])
	require.Nil(err)
	hash := tsf1.Hash()
	blk := NewBlock(1, 3, hash, clock.New(), []*action.Transfer{tsf1, tsf2}, nil, nil)
	err = blk.SignBlock(ta.Addrinfo["producer"])
	require.Nil(err)
	require.Nil(val.Validate(blk, 2, hash))
}

func TestWrongNonce(t *testing.T) {
	cfg := &config.Default
	testutil.CleanupPath(t, cfg.Chain.TrieDBPath)
	defer testutil.CleanupPath(t, cfg.Chain.TrieDBPath)
	testutil.CleanupPath(t, cfg.Chain.ChainDBPath)
	defer testutil.CleanupPath(t, cfg.Chain.ChainDBPath)
	require := require.New(t)
	sf, err := state.NewFactory(cfg, state.DefaultTrieOption())
	require.NoError(err)
	require.NoError(sf.Start(context.Background()))
	_, err = sf.LoadOrCreateState(ta.Addrinfo["producer"].RawAddress, Gen.TotalSupply)
	assert.NoError(t, err)
	val := validator{sf}
	require.Nil(sf.CommitStateChanges(0, nil, nil, nil))

	// correct nonce
	coinbaseTsf := action.NewCoinBaseTransfer(big.NewInt(int64(Gen.BlockReward)), ta.Addrinfo["producer"].RawAddress)
	tsf1, err := action.NewTransfer(1, big.NewInt(20), ta.Addrinfo["producer"].RawAddress, ta.Addrinfo["alfa"].RawAddress, []byte{}, uint64(100000), big.NewInt(10))
	require.NoError(err)
	tsf1, err = tsf1.Sign(ta.Addrinfo["producer"])
	require.NoError(err)
	hash := tsf1.Hash()
	blk := NewBlock(1, 3, hash, clock.New(), []*action.Transfer{coinbaseTsf, tsf1}, nil, nil)
	err = blk.SignBlock(ta.Addrinfo["producer"])
	require.NoError(err)
	require.Nil(val.Validate(blk, 2, hash))
	require.NoError(sf.CommitStateChanges(1, []*action.Transfer{tsf1}, nil, nil))

	// low nonce
	tsf2, err := action.NewTransfer(1, big.NewInt(30), ta.Addrinfo["producer"].RawAddress, ta.Addrinfo["bravo"].RawAddress, []byte{}, uint64(100000), big.NewInt(10))
	require.NoError(err)
	tsf2, err = tsf2.Sign(ta.Addrinfo["producer"])
	require.NoError(err)
	hash = tsf1.Hash()
	blk = NewBlock(1, 3, hash, clock.New(), []*action.Transfer{coinbaseTsf, tsf1, tsf2}, nil, nil)
	err = blk.SignBlock(ta.Addrinfo["producer"])
	require.NoError(err)
	err = val.Validate(blk, 2, hash)
	require.Equal(ErrActionNonce, errors.Cause(err))

	vote, err := action.NewVote(1, ta.Addrinfo["producer"].RawAddress, ta.Addrinfo["bravo"].RawAddress, uint64(100000), big.NewInt(10))
	require.NoError(err)
	vote, err = vote.Sign(ta.Addrinfo["producer"])
	require.NoError(err)
	hash = tsf1.Hash()
	blk = NewBlock(1, 3, hash, clock.New(), []*action.Transfer{coinbaseTsf}, []*action.Vote{vote}, nil)
	err = blk.SignBlock(ta.Addrinfo["producer"])
	require.NoError(err)
	err = val.Validate(blk, 2, hash)
	require.Error(err)
	require.Equal(ErrActionNonce, errors.Cause(err))

	// duplicate nonce
	tsf3, err := action.NewTransfer(2, big.NewInt(30), ta.Addrinfo["producer"].RawAddress, ta.Addrinfo["bravo"].RawAddress, []byte{}, uint64(100000), big.NewInt(10))
	require.NoError(err)
	tsf3, err = tsf3.Sign(ta.Addrinfo["producer"])
	require.NoError(err)
	tsf4, err := action.NewTransfer(2, big.NewInt(30), ta.Addrinfo["producer"].RawAddress, ta.Addrinfo["bravo"].RawAddress, []byte{}, uint64(100000), big.NewInt(10))
	require.NoError(err)
	tsf4, err = tsf4.Sign(ta.Addrinfo["producer"])
	require.NoError(err)
	hash = tsf1.Hash()
	blk = NewBlock(1, 3, hash, clock.New(), []*action.Transfer{coinbaseTsf, tsf3, tsf4}, nil, nil)
	err = blk.SignBlock(ta.Addrinfo["producer"])
	require.NoError(err)
	err = val.Validate(blk, 2, hash)
	require.Error(err)
	require.Equal(ErrActionNonce, errors.Cause(err))

	vote2, err := action.NewVote(2, ta.Addrinfo["producer"].RawAddress, ta.Addrinfo["bravo"].RawAddress, uint64(100000), big.NewInt(10))
	require.NoError(err)
	vote2, err = vote2.Sign(ta.Addrinfo["producer"])
	require.NoError(err)
	vote3, err := action.NewVote(2, ta.Addrinfo["producer"].RawAddress, ta.Addrinfo["charlie"].RawAddress, uint64(100000), big.NewInt(10))
	require.NoError(err)
	vote3, err = vote3.Sign(ta.Addrinfo["producer"])
	require.NoError(err)
	hash = tsf1.Hash()
	blk = NewBlock(1, 3, hash, clock.New(), []*action.Transfer{coinbaseTsf}, []*action.Vote{vote2, vote3}, nil)
	err = blk.SignBlock(ta.Addrinfo["producer"])
	require.NoError(err)
	err = val.Validate(blk, 2, hash)
	require.Error(err)
	require.Equal(ErrActionNonce, errors.Cause(err))

	// non consecutive nonce
	tsf5, err := action.NewTransfer(2, big.NewInt(30), ta.Addrinfo["producer"].RawAddress, ta.Addrinfo["bravo"].RawAddress, []byte{}, uint64(100000), big.NewInt(10))
	require.NoError(err)
	tsf5, err = tsf5.Sign(ta.Addrinfo["producer"])
	require.NoError(err)
	tsf6, err := action.NewTransfer(4, big.NewInt(30), ta.Addrinfo["producer"].RawAddress, ta.Addrinfo["bravo"].RawAddress, []byte{}, uint64(100000), big.NewInt(10))
	require.NoError(err)
	tsf6, err = tsf6.Sign(ta.Addrinfo["producer"])
	require.NoError(err)
	hash = tsf1.Hash()
	blk = NewBlock(1, 3, hash, clock.New(), []*action.Transfer{coinbaseTsf, tsf5, tsf6}, nil, nil)
	err = blk.SignBlock(ta.Addrinfo["producer"])
	require.NoError(err)
	err = val.Validate(blk, 2, hash)
	require.Error(err)
	require.Equal(ErrActionNonce, errors.Cause(err))

	vote4, err := action.NewVote(2, ta.Addrinfo["producer"].RawAddress, ta.Addrinfo["bravo"].RawAddress, uint64(100000), big.NewInt(10))
	require.NoError(err)
	vote4, err = vote4.Sign(ta.Addrinfo["producer"])
	require.NoError(err)
	vote5, err := action.NewVote(4, ta.Addrinfo["producer"].RawAddress, ta.Addrinfo["charlie"].RawAddress, uint64(100000), big.NewInt(10))
	require.NoError(err)
	vote5, err = vote5.Sign(ta.Addrinfo["producer"])
	require.NoError(err)
	hash = tsf1.Hash()
	blk = NewBlock(1, 3, hash, clock.New(), []*action.Transfer{coinbaseTsf}, []*action.Vote{vote4, vote5}, nil)
	err = blk.SignBlock(ta.Addrinfo["producer"])
	require.NoError(err)
	err = val.Validate(blk, 2, hash)
	require.Error(err)
	require.Equal(ErrActionNonce, errors.Cause(err))
}

func TestWrongCoinbaseTsf(t *testing.T) {
	cfg := &config.Default
	testutil.CleanupPath(t, cfg.Chain.TrieDBPath)
	defer testutil.CleanupPath(t, cfg.Chain.TrieDBPath)
	testutil.CleanupPath(t, cfg.Chain.ChainDBPath)
	defer testutil.CleanupPath(t, cfg.Chain.ChainDBPath)
	require := require.New(t)
	sf, err := state.NewFactory(cfg, state.DefaultTrieOption())
	require.NoError(err)
	require.NoError(sf.Start(context.Background()))
	_, err = sf.LoadOrCreateState(ta.Addrinfo["producer"].RawAddress, Gen.TotalSupply)
	assert.NoError(t, err)
	val := validator{sf}
	require.Nil(sf.CommitStateChanges(0, nil, nil, nil))

	// no coinbase tsf
	coinbaseTsf := action.NewCoinBaseTransfer(big.NewInt(int64(Gen.BlockReward)), ta.Addrinfo["producer"].RawAddress)
	tsf1, err := action.NewTransfer(1, big.NewInt(20), ta.Addrinfo["producer"].RawAddress, ta.Addrinfo["alfa"].RawAddress, []byte{}, uint64(100000), big.NewInt(10))
	require.NoError(err)
	tsf1, err = tsf1.Sign(ta.Addrinfo["producer"])
	require.NoError(err)
	hash := tsf1.Hash()
	blk := NewBlock(1, 3, hash, clock.New(), []*action.Transfer{tsf1}, nil, nil)
	err = blk.SignBlock(ta.Addrinfo["producer"])
	require.NoError(err)
	err = val.Validate(blk, 2, hash)
	require.Error(err)
	require.True(
		strings.Contains(err.Error(), "wrong number of coinbase transfers"),
	)

	// extra coinbase transfer
	blk = NewBlock(1, 3, hash, clock.New(), []*action.Transfer{coinbaseTsf, coinbaseTsf, tsf1}, nil, nil)
	err = blk.SignBlock(ta.Addrinfo["producer"])
	require.NoError(err)
	err = val.Validate(blk, 2, hash)
	require.Error(err)
	require.True(
		strings.Contains(err.Error(), "wrong number of coinbase transfers"),
	)

	// no transfer
	blk = NewBlock(1, 3, hash, clock.New(), []*action.Transfer{}, nil, nil)
	err = blk.SignBlock(ta.Addrinfo["producer"])
	require.NoError(err)
	err = val.Validate(blk, 2, hash)
	require.Error(err)
	require.True(
		strings.Contains(err.Error(), "wrong number of coinbase transfers"),
	)
}

func TestWrongAddress(t *testing.T) {
	val := validator{}
	invalidRecipient := "io1qyqsyqcyq5narhapakcsrhksfajfcpl24us3xp38zwvsep"
	tsf, err := action.NewTransfer(1, big.NewInt(1), ta.Addrinfo["producer"].RawAddress, invalidRecipient, []byte{}, uint64(100000), big.NewInt(10))
	require.NoError(t, err)
	blk1 := NewBlock(1, 3, hash.ZeroHash32B, clock.New(), []*action.Transfer{tsf}, nil, nil)
	err = val.verifyActions(blk1)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to validate transfer recipient's address"))

	invalidVotee := "ioaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	vote, err := action.NewVote(1, ta.Addrinfo["producer"].RawAddress, invalidVotee, uint64(100000), big.NewInt(10))
	require.NoError(t, err)
	blk2 := NewBlock(1, 3, hash.ZeroHash32B, clock.New(), nil, []*action.Vote{vote}, nil)
	err = val.verifyActions(blk2)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to validate votee's address"))

	invalidContract := "123"
	execution, err := action.NewExecution(ta.Addrinfo["producer"].RawAddress, invalidContract, 1, big.NewInt(1), uint64(100000), big.NewInt(10), []byte{})
	require.NoError(t, err)
	blk3 := NewBlock(1, 3, hash.ZeroHash32B, clock.New(), nil, nil, []*action.Execution{execution})
	err = val.verifyActions(blk3)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to validate contract's address"))
}
