// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package caclient

import (
	calib "github.com/HiNounou029/nounouchain/caclient/lib"
	. "github.com/HiNounou029/nounouchain/caclient/lib/tcert"
)

func ValCert(rootPaht string, certInfo []byte) error {
	peerCert, err := GetCertificate(certInfo)
	if err != nil {
		return  err
	}

	// Verification certificate
	caCfg := &calib.CAConfig{}
	caCfg.CA.Chainfile = rootPaht

	// Verify that the certificate is expired
	err = calib.ValidateDates(peerCert)
	if err != nil {
		return err
	}

	var CAChain calib.CA
	CAChain.Config = caCfg
	err = CAChain.VerifyCertificate(peerCert)
	if err != nil{
		return err
	}

	return  nil
}
