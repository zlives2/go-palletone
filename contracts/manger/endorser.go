/*
	This file is part of go-palletone.
	go-palletone is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	go-palletone is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/
/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package manger

import (
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/contracts/rwset"
	"github.com/palletone/go-palletone/contracts/core"
	"github.com/palletone/go-palletone/core/vmContractPub/util"
	"github.com/palletone/go-palletone/core/vmContractPub/flogging"
	//ut "github.com/palletone/go-palletone/dag/modules"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	putils "github.com/palletone/go-palletone/core/vmContractPub/protos/utils"
)

type chaincodeError struct {
	status int32
	msg    string
}

func (ce chaincodeError) Error() string {
	return fmt.Sprintf("chaincode error (status: %d, message: %s)", ce.status, ce.msg)
}

var logger = flogging.MustGetLogger("ccmanger")

// Support contains functions that the endorser requires to execute its tasks
type Support interface {
	IsSysCCAndNotInvokableExternal(name string) bool
    // GetTxSimulator returns the transaction simulator ,they are made unique
	// by way of the supplied txid
	GetTxSimulator(chainid string, txid string) (rwset.TxSimulator, error)

	IsSysCC(name string) bool

	Execute(ctxt context.Context, cid, name, version, txid string, syscc bool, signedProp *pb.SignedProposal, prop *pb.Proposal, spec interface{}) (*pb.Response, *pb.ChaincodeEvent, error)
}

// Endorser provides the Endorser service ProcessProposal
type Endorser struct {
	//distributePrivateData privateDataDistributor
	s                     Support
}

// validateResult provides the result of endorseProposal verification
type validateResult struct {
	prop    *pb.Proposal
	hdrExt  *pb.ChaincodeHeaderExtension
	chainID string
	txid    string
	resp    *pb.ProposalResponse
}

// NewEndorserServer creates and returns a new Endorser server instance.
func NewEndorserServer(s Support) pb.EndorserServer {
	e := &Endorser{
		s: s,
	}
	return e
}

//call specified chaincode (system or user)
func (e *Endorser) callChaincode(ctxt context.Context, chainID string, version string, txid string, signedProp *pb.SignedProposal, prop *pb.Proposal, cis *pb.ChaincodeInvocationSpec, chaincodeName string, txsim rwset.TxSimulator) (*pb.Response, *pb.ChaincodeEvent, error) {
	logger.Debugf("[%s][%s] Entry chaincode: %s version: %s", chainID, shorttxid(txid), chaincodeName, version)
	defer logger.Debugf("[%s][%s] Exit", chainID, shorttxid(txid))
	var err error
	var res *pb.Response
	var ccevent *pb.ChaincodeEvent

	if txsim != nil {
		ctxt = context.WithValue(ctxt, core.TXSimulatorKey, txsim)
	}

	scc := e.s.IsSysCC(chaincodeName)
	res, ccevent, err = e.s.Execute(ctxt, chainID, chaincodeName, version, txid, scc, signedProp, prop, cis)
	if err != nil {
		return nil, nil, err
	}

	if res.Status >= shim.ERRORTHRESHOLD {
		return res, nil, nil
	}

	return res, ccevent, err
}

func (e *Endorser) simulateProposal(ctx context.Context, chainID string, txid string, signedProp *pb.SignedProposal, prop *pb.Proposal, cid *pb.ChaincodeID, txsim rwset.TxSimulator) (*pb.Response, []byte, *pb.ChaincodeEvent, error) {
	logger.Debugf("[%s][%s] Entry chaincode: %s", chainID, shorttxid(txid), cid)
	defer logger.Debugf("[%s][%s] Exit", chainID, shorttxid(txid))

	cis, err := putils.GetChaincodeInvocationSpec(prop)
	if err != nil {
		return nil, nil, nil, err
	}
	logger.Infof("spec=%v", cis)

	//var cdLedger resourcesconfig.ChaincodeDefinition
	//
	//if !e.s.IsSysCC(cid.Name) {
	//	cdLedger, err = e.s.GetChaincodeDefinition(ctx, chainID, txid, signedProp, prop, cid.Name, txsim)
	//	if err != nil {
	//		return nil, nil, nil, nil, errors.WithMessage(err, fmt.Sprintf("make sure the chaincode %s has been successfully instantiated and try again", cid.Name))
	//	}
	//	version = cdLedger.CCVersion()
	//
	//	err = e.s.CheckInstantiationPolicy(cid.Name, version, cdLedger)
	//	if err != nil {
	//		return nil, nil, nil, nil, err
	//	}
	//} else {
	//	version = util.GetSysCCVersion()
	//}

	//---3. execute the proposal and get simulation results
	//var simResult *ledger.TxSimulationResults
	var simResBytes []byte
	var res *pb.Response
	var ccevent *pb.ChaincodeEvent
	res, ccevent, err = e.callChaincode(ctx, chainID, cid.Version, txid, signedProp, prop, cis, cid.Name, txsim)
	if err != nil {
		logger.Errorf("[%s][%s] failed to invoke chaincode %s, error: %+v", chainID, shorttxid(txid), cid, err)
		return  nil, nil, nil, err
	}

	if txsim != nil {
		//if simResult, err = txsim.GetTxSimulationResults(); err != nil {
		//	return  nil, nil, nil, err
		//}
	}
	return res, simResBytes, ccevent, nil
}

//endorse the proposal
func (e *Endorser) endorseProposal(ctx context.Context, chainID string, txid string, signedProp *pb.SignedProposal, proposal *pb.Proposal, response *pb.Response, simRes []byte, event *pb.ChaincodeEvent, visibility []byte, ccid *pb.ChaincodeID, txsim rwset.TxSimulator) (*pb.ProposalResponse, error) {
	logger.Debugf("[%s][%s] Entry chaincode: %s", chainID, shorttxid(txid), ccid)
	defer logger.Debugf("[%s][%s] Exit", chainID, shorttxid(txid))

	return nil, nil
}

//preProcess checks the tx proposal headers, uniqueness and ACL
func (e *Endorser) validateProcess(signedProp *pb.SignedProposal) (*validateResult, error) {
	vr := &validateResult{}

	// extract the Proposal message from signedProp
	prop, err := putils.GetProposal(signedProp.ProposalBytes)
	if err != nil {
		return nil, err
	}

	// 1) look at the ProposalHeader
	hdr, err := putils.GetHeader(prop.Header)
	if err != nil {
		return nil, err
	}

	//TODO validate the header

	if err != nil {
		vr.resp = &pb.ProposalResponse{Response: &pb.Response{Status: 500, Message: err.Error()}}
		return vr, err
	}

	chdr, err := putils.UnmarshalChannelHeader(hdr.ChannelHeader)
	if err != nil {
		vr.resp = &pb.ProposalResponse{Response: &pb.Response{Status: 500, Message: err.Error()}}
		return vr, err
	}
	//shdr, err := putils.GetSignatureHeader(hdr.SignatureHeader)
	//if err != nil {
	//	vr.resp = &pb.ProposalResponse{Response: &pb.Response{Status: 500, Message: err.Error()}}
	//	return vr, err
	//}

	vr.prop, vr.chainID, vr.txid = prop, chdr.ChannelId, chdr.TxId

	return vr, nil
}

// ProcessProposal process the Proposal
//func (e *Endorser) ProcessProposal(ctx context.Context, signedProp *pb.SignedProposal) (*pb.ProposalResponse, error) {
func (e *Endorser) ProcessProposal(ctx context.Context, signedProp *pb.SignedProposal, prop *pb.Proposal, chainID string, cid *pb.ChaincodeID) (*pb.ProposalResponse, *pb.ContractInvokePayload, error) {
	var txsim rwset.TxSimulator

	addr := util.ExtractRemoteAddress(ctx)
	logger.Debug("Entering: Got request from", addr)
	defer logger.Debugf("Exit: request from", addr)

	//0 -- check and validate
	result, err := e.validateProcess(signedProp)
	if err != nil {
		logger.Debugf("validate signedProp err:%s", err)
		return nil, nil, err
	}

	txid := result.txid
	if chainID != "" {
		if txsim, err = e.s.GetTxSimulator(chainID, txid); err != nil {
			return &pb.ProposalResponse{Response: &pb.Response{Status: 500, Message: err.Error()}}, nil, err
		}
		//defer txsim.Done()
	}
	if  err != nil {
		return &pb.ProposalResponse{Response: &pb.Response{Status: 500, Message: err.Error()}}, nil, err
	}

	//1 -- simulate
	res, _, ccevent, err := e.simulateProposal(ctx, chainID, txid, signedProp, prop, cid, txsim)
	if err != nil {
		logger.Error(ccevent)
		return &pb.ProposalResponse{Response: &pb.Response{Status: 500, Message: err.Error()}}, nil, err
	}
	if res != nil {
		if res.Status >= shim.ERROR {
			logger.Errorf("[%s][%s] simulateProposal() resulted in chaincode[] response status %d for txid: %s",
				chainID, shorttxid(txid),  res.Status, txid)

			resp := &pb.ProposalResponse{
				Payload:  nil,
				Response: &pb.Response{Status: 500, Message: "Chaincode Error"}}
			return resp, nil, err
		}
	}else {
		logger.Error("simulateProposal response is nil")
		return &pb.ProposalResponse{
			Payload:  nil, Response: &pb.Response{Status: 500, Message: "Chaincode Error"}}, nil, nil
	}

	//2 -- endorse and get a marshalled ProposalResponse message
	pResp := &pb.ProposalResponse{Response: res}

	cis, err := putils.GetChaincodeInvocationSpec(prop)
	if err != nil {
	}

	unit, err := RwTxResult2DagInvokeUnit(txsim, txid, cis.ChaincodeSpec.ChaincodeId.Name, cis.ChaincodeSpec.Input.Args[0])
	if err != nil {
		logger.Errorf("chainID[%s] converRwTxResult2DagUnit failed", chainID)
		return nil, nil, errors.New("Conver RwSet to dag unit fail")
	}
	logger.Debug("unit:")
	logger.Debug(unit)
	// todo

	pResp.Response.Payload = res.Payload

	return pResp, unit, nil
}
