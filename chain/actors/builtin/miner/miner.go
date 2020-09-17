package miner

import (
	"github.com/filecoin-project/go-state-types/dline"
	"github.com/libp2p/go-libp2p-core/peer"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-bitfield"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/cbor"
	v0builtin "github.com/filecoin-project/specs-actors/actors/builtin"
	v0miner "github.com/filecoin-project/specs-actors/actors/builtin/miner"

	"github.com/filecoin-project/lotus/chain/actors/adt"
	"github.com/filecoin-project/lotus/chain/types"
)

var Address = v0builtin.InitActorAddr

func Load(store adt.Store, act *types.Actor) (st State, err error) {
	switch act.Code {
	case v0builtin.StorageMinerActorCodeID:
		out := v0State{store: store}
		err := store.Get(store.Context(), act.Head, &out)
		if err != nil {
			return nil, err
		}
		return &out, nil
	}
	return nil, xerrors.Errorf("unknown actor code %s", act.Code)
}

type State interface {
	cbor.Marshaler

	AvailableBalance(abi.TokenAmount) (abi.TokenAmount, error)
	VestedFunds(abi.ChainEpoch) (abi.TokenAmount, error)

	GetSector(abi.SectorNumber) (*SectorOnChainInfo, error)
	FindSector(abi.SectorNumber) (*SectorLocation, error)
	GetSectorExpiration(abi.SectorNumber) (*SectorExpiration, error)
	GetPrecommittedSector(abi.SectorNumber) (*SectorPreCommitOnChainInfo, error)
	LoadSectorsFromSet(filter *bitfield.BitField, filterOut bool) (adt.Array, error)
	LoadPreCommittedSectors() (adt.Map, error)
	IsAllocated(abi.SectorNumber) (bool, error)

	LoadDeadline(idx uint64) (Deadline, error)
	ForEachDeadline(cb func(idx uint64, dl Deadline) error) error
	NumDeadlines() (uint64, error)
	Info() (MinerInfo, error)

	DeadlineInfo(epoch abi.ChainEpoch) *dline.Info
	WpostProvingPeriod() abi.ChainEpoch
}

type Deadline interface {
	LoadPartition(idx uint64) (Partition, error)
	ForEachPartition(cb func(idx uint64, part Partition) error) error
	PostSubmissions() (bitfield.BitField, error)
}

type Partition interface {
	AllSectors() (bitfield.BitField, error)
	FaultySectors() (bitfield.BitField, error)
	RecoveringSectors() (bitfield.BitField, error)
	LiveSectors() (bitfield.BitField, error)
	ActiveSectors() (bitfield.BitField, error)
}

type SectorOnChainInfo = v0miner.SectorOnChainInfo
type SectorPreCommitInfo = v0miner.SectorPreCommitInfo
type SectorPreCommitOnChainInfo = v0miner.SectorPreCommitOnChainInfo
type PoStPartition = v0miner.PoStPartition
type RecoveryDeclaration = v0miner.RecoveryDeclaration
type FaultDeclaration = v0miner.FaultDeclaration

// Params
type DeclareFaultsParams = v0miner.DeclareFaultsParams
type DeclareFaultsRecoveredParams = v0miner.DeclareFaultsRecoveredParams
type SubmitWindowedPoStParams = v0miner.SubmitWindowedPoStParams
type ProveCommitSectorParams = v0miner.ProveCommitSectorParams

type MinerInfo struct {
	Owner                      address.Address   // Must be an ID-address.
	Worker                     address.Address   // Must be an ID-address.
	NewWorker                  address.Address   // Must be an ID-address.
	ControlAddresses           []address.Address // Must be an ID-addresses.
	WorkerChangeEpoch          abi.ChainEpoch
	PeerId                     *peer.ID
	Multiaddrs                 []abi.Multiaddrs
	SealProofType              abi.RegisteredSealProof
	SectorSize                 abi.SectorSize
	WindowPoStPartitionSectors uint64
}

type ChainSectorInfo struct {
	Info SectorOnChainInfo
	ID   abi.SectorNumber
}

type SectorExpiration struct {
	OnTime abi.ChainEpoch

	// non-zero if sector is faulty, epoch at which it will be permanently
	// removed if it doesn't recover
	Early abi.ChainEpoch
}

type SectorLocation struct {
	Deadline  uint64
	Partition uint64
}
