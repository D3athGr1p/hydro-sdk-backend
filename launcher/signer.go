
package launcher

import (
	"crypto/ecdsa"
	"database/sql"

	"github.com/ethereum/go-ethereum/common"
	ethType "github.com/ethereum/go-ethereum/core/types"

	"github.com/D3athgr1p/hydro-sdk-backend/sdk/crypto"
	// "github.com/D3athgr1p/hydro-sdk-backend/sdk/signer"
	// "github.com/HydroProtocol/hydro-sdk-backend/sdk/types"
	"github.com/D3athgr1p/hydro-sdk-backend/utils"
	"sync"
)

type ISignService interface {
	Sign(launchLog *LaunchLog) string
	AfterSign() //what you want to do when signature has been used
}

type localSignService struct {
	privateKey *ecdsa.PrivateKey
	nonce      int64
	mutex      sync.Mutex
}

func (s *localSignService) AfterSign() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.nonce = s.nonce + 1
}

func (s *localSignService) Sign(launchLog *LaunchLog) string {
	
	transaction := ethType.NewTransaction(
		uint64(s.nonce),
		common.HexToAddress(launchLog.To),
		utils.DecimalToBigInt(launchLog.Value),
		uint64(launchLog.GasLimit),
		utils.DecimalToBigInt(launchLog.GasPrice.Decimal),
		utils.Hex2Bytes(launchLog.Data[2:]),
	)

	// signedTransaction, err := signer.SignTx(transaction, s.privateKey)
	signers := ethType.HomesteadSigner{}
	signedTransaction, err := ethType.SignTx(transaction, signers, s.privateKey)

	if err != nil {
		utils.Errorf("sign transaction error: %v", err)
		panic(err)
	}

	launchLog.Nonce = sql.NullInt64{
		Int64: s.nonce,
		Valid: true,
	}

	launchLog.Hash = sql.NullString{
		String: signedTransaction.Hash().String(),
		Valid:  true,
	}

	// return utils.Bytes2HexP(signer.EncodeRlp(signedTransaction))
	bty, _ := signedTransaction.MarshalBinary()
	return utils.Bytes2HexP(bty)
}

func NewDefaultSignService(privateKeyStr string, getNonce func(string) (int, error)) ISignService {
	utils.Infof(privateKeyStr)
	privateKey, err := crypto.NewPrivateKeyByHex(privateKeyStr)
	if err != nil {
		panic(err)
	}

	//nonce := LaunchLogDao.FindPendingLogWithMaxNonce() + 1
	chainNonce, err := getNonce(crypto.PubKey2Address(privateKey.PublicKey))

	if err != nil {
		panic(err)
	}

	//if int64(chainNonce) > nonce {
	//	nonce = int64(chainNonce)
	//}

	return &localSignService{
		privateKey: privateKey,
		mutex:      sync.Mutex{},
		nonce:      int64(chainNonce),
	}
}
