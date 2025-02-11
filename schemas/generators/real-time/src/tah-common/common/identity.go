package common

type IdentityProof struct {
	ProofId       string `json:"proofId"`
	ProofTypeId   string `json:"proofTypeId"`
	ProofTypeName string `json:"proofTypeName"`
	ProofData     string `json:"proofData"`
}
