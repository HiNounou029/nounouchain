// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package gen

//go:generate rm -rf ./compiled/
//go:generate solc --optimize-runs 200 --overwrite --bin-runtime --abi -o ./compiled authority.sol executor.sol extension.sol measure.sol params.sol prototype.sol
//go:generate go-bindata -nometadata -ignore=_ -pkg gen -o bindata.go compiled/
