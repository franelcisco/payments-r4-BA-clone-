package models

type R4ConsultaRequest struct {
	IdCliente        string `json:"IdCliente"`
	Monto            string `json:"Monto"`
	TelefonoComercio string `json:"TelefonoComercio"`
}

type R4NotificaRequest struct {
	IdComercio       string `json:"IdComercio"`
	TelefonoComercio string `json:"TelefonoComercio"`
	TelefonoEmisor   string `json:"TelefonoEmisor"`
	Concepto         string `json:"Concepto"`
	BancoEmisor      string `json:"BancoEmisor"`
	Monto            string `json:"Monto"`
	FechaHora        string `json:"FechaHora"`
	Referencia       string `json:"Referencia"`
	CodigoRed        string `json:"CodigoRed"`
}
