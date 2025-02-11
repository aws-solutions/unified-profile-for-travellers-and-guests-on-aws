package common

type IdentityProof struct {
	ProofId       string `json:"proofId"`                 //ID of the proof (ex: passport number, )
	ProofTypeId   string `json:"proofTypeId"`             //ID of the type of proof (ex: passport, driving license, etc.)
	ProofTypeName string `json:"proofTypeName,omitempty"` //Name of the type of proof (ex: passport, driving license, etc.)
	ProofData     string `json:"proofData"`               //Data of the proof (ex: passport number, driving license number, etc.)
}
