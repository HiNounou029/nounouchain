// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package cert

import (
	"encoding/json"
	"github.com/HiNounou029/nounouchain/api/utils"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	. "github.com/HiNounou029/nounouchain/caclient"
)

type Cert struct {
	rootCertPath string
}

func New(path string) *Cert {
	return &Cert{
		path,
	}
}

func (n *Cert) handleCertVerify(w http.ResponseWriter, req *http.Request) error {
	reqBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	reqCertJson := make(map[string]interface{})
	json.Unmarshal(reqBytes, &reqCertJson)
	certByteString := reqCertJson["cert"].(string)

	respMap := make(map[string]interface{})
	respMap["result"] = "success"

	err = ValCert(n.rootCertPath, []byte(certByteString))
	if err != nil {
		respMap["result"] = err.Error()
	}

	bytesData, err := json.Marshal(respMap)
	if err != nil {
		return err
	}
	w.Write(bytesData)

	return nil
}

func (n *Cert) Mount(root *mux.Router, pathPrefix string) {
	sub := root.PathPrefix(pathPrefix).Subrouter()
	sub.Path("/cert").Methods("Post").HandlerFunc(utils.WrapHandlerFunc(n.handleCertVerify))
}
