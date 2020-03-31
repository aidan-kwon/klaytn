// Modifications Copyright 2018 The klaytn Authors
// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
//
// This file is derived from quorum/consensus/istanbul/core/prepare.go (2018/06/04).
// Modified and improved for the klaytn development.

package core

import (
	"github.com/klaytn/klaytn/consensus/istanbul"
	"reflect"
)

func (c *core) sendPrepare() {
	logger := c.logger.NewWith("state", c.state)

	sub := c.current.Subject()
	encodedSubject, err := Encode(sub)
	if err != nil {
		logger.Error("Failed to encode", "subject", sub)
		return
	}

	c.broadcast(&message{
		Hash: c.current.Proposal().ParentHash(),
		Code: msgPrepare,
		Msg:  encodedSubject,
	})
}

func (c *core) handlePrepare(msg *message, src istanbul.Validator) error {
	// Decode PREPARE message
	var prepare *istanbul.Subject
	err := msg.Decode(&prepare)
	if err != nil {
		return errFailedDecodePrepare
	}

	//logger.Error("call receive prepare","num",prepare.View.Sequence)

	if err := c.checkMessage(msgPrepare, prepare.View); err != nil {
		return err
	}

	if prepare.View.Sequence.Uint64() == 100 && prepare.View.Round.Uint64() == 0 {
		logger.Info("print validator list", "validators", c.valSet.List())

		for i := 0; i < c.valSet.F()+1; i++ {
			if c.valSet.GetByIndex(uint64(i)).Address() == c.address {
				logger.Warn("Pretend a faulty node", "idx", i)
				return errNotFromProposer
			}
		}
		logger.Warn("Pretend a non-faulty node")
	}

	// If it is locked, it can only process on the locked block.
	// Passing verifyPrepare and checkMessage implies it is processing on the locked block since it was verified in the Preprepared state.
	if err := c.verifyPrepare(prepare, src); err != nil {
		return err
	}

	c.acceptPrepare(msg, src)

	// Change to Prepared state if we've received enough PREPARE messages or it is locked
	// and we are in earlier state before Prepared state.
	if ((c.current.IsHashLocked() && prepare.Digest == c.current.GetLockedHash()) || c.current.GetPrepareOrCommitSize() > 2*c.valSet.F()) &&
		c.state.Cmp(StatePrepared) < 0 {
		c.current.LockHash()
		c.setState(StatePrepared)
		c.sendCommit()
	}

	return nil
}

// verifyPrepare verifies if the received PREPARE message is equivalent to our subject
func (c *core) verifyPrepare(prepare *istanbul.Subject, src istanbul.Validator) error {
	logger := c.logger.NewWith("from", src, "state", c.state)

	sub := c.current.Subject()
	if !reflect.DeepEqual(prepare, sub) {
		logger.Warn("Inconsistent subjects between PREPARE and proposal", "expected", sub, "got", prepare)
		return errInconsistentSubject
	}

	return nil
}

func (c *core) acceptPrepare(msg *message, src istanbul.Validator) error {
	logger := c.logger.NewWith("from", src, "state", c.state)

	// Add the PREPARE message to current round state
	if err := c.current.Prepares.Add(msg); err != nil {
		logger.Error("Failed to add PREPARE message to round state", "msg", msg, "err", err)
		return err
	}

	return nil
}
