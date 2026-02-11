package models

import (
	"time"
)

// PeticionUniforme mapea a la tabla "Petición Uniforme"
type PeticionUniforme struct {
	IDPeticion        int        `gorm:"primaryKey;column:id_peticion" json:"id_peticion"`
	IdFuncionario     int        `gorm:"column:id_funcionario" json:"id_funcionario"`
	IdUniforme        int        `gorm:"column:id_uniforme" json:"id_uniforme"`
	IdTipoPeticion    int        `gorm:"column:id_tipo_peticion" json:"id_tipo_peticion"`
	IdTemporada       *int       `gorm:"column:id_temporada" json:"id_temporada"`
	IdEstado          int        `gorm:"column:id_estado" json:"id_estado"`
	IdDespacho        *int       `gorm:"column:id_despacho" json:"id_despacho"`
	Maternal          bool       `gorm:"column:maternal;default:false" json:"maternal"`
	Observacion       *string    `gorm:"column:observación" json:"observacion"`
	FechaRegistro     time.Time  `gorm:"column:fecha_registro" json:"fecha_registro"`
	FechaModificacion *time.Time `gorm:"column:fecha_modificacion" json:"fecha_modificacion"`

	// Relaciones
	Funcionario  Funcionario  `gorm:"foreignKey:IdFuncionario" json:"funcionario"`
	Uniforme     Uniforme     `gorm:"foreignKey:IdUniforme" json:"uniforme"`
	TipoPeticion TipoPeticion `gorm:"foreignKey:IdTipoPeticion" json:"tipo_peticion"`
	Temporada    *Temporada   `gorm:"foreignKey:IdTemporada" json:"temporada"`
	Estado       Estado       `gorm:"foreignKey:IdEstado" json:"estado"`
	Despacho     *Despacho    `gorm:"foreignKey:IdDespacho" json:"despacho"`
	Tallajes     []Tallaje    `gorm:"foreignKey:IDPeticion" json:"tallaje"`
}

func (PeticionUniforme) TableName() string {
	return "\"Petición Uniforme\""
}

// Tallaje mapea a la tabla "Tallaje"
type Tallaje struct {
	IDTallaje  int    `gorm:"primaryKey;column:id_tallaje" json:"id_tallaje"`
	IDPeticion int    `gorm:"column:id_peticion" json:"id_peticion"`
	IDPrenda   int    `gorm:"column:id_prenda" json:"id_prenda"`
	ValorTalla string `gorm:"column:valor_talla" json:"valor_talla"`
	Cantidad   int    `gorm:"column:cantidad" json:"cantidad"`

	// Relaciones
	Prenda Prenda `gorm:"foreignKey:IDPrenda" json:"prenda"`
}

func (Tallaje) TableName() string {
	return "\"Tallaje\""
}

// Despacho mapea a la tabla "Despacho"
type Despacho struct {
	IDDespacho           int        `gorm:"primaryKey;column:id_despacho" json:"id_despacho"`
	GuiaDeDespacho       string     `gorm:"column:guia_de_despacho;unique" json:"guia_de_despacho"`
	FechaDespacho        *time.Time `gorm:"column:fecha_despacho" json:"fecha_despacho"`
	NumeroVoucher        *string    `gorm:"column:numero_voucher" json:"numero_voucher"`
	Sucursal             string     `gorm:"column:sucursal" json:"sucursal"`
	ResponsableRecepcion *string    `gorm:"column:responsable_recepcion" json:"responsable_recepcion"`
	IdEstadoDespacho     int        `gorm:"column:id_estado_despacho" json:"id_estado_despacho"`
	IdFactura            *int       `gorm:"column:id_factura" json:"id_factura"`

	// Relaciones
	EstadoDespacho Estado             `gorm:"foreignKey:IdEstadoDespacho" json:"estado_despacho"`
	Factura        *Factura           `gorm:"foreignKey:IdFactura" json:"factura"`
	Peticiones     []PeticionUniforme `gorm:"foreignKey:IdDespacho" json:"peticiones"`
}

func (Despacho) TableName() string {
	return "\"Despacho\""
}

// Notificacion mapea a la tabla "Notificacion"
type Notificacion struct {
	IDNotificacion int       `gorm:"primaryKey;column:id_notificacion" json:"id_notificacion"`
	IDUsuario      int       `gorm:"column:id_usuario" json:"id_usuario"`
	Titulo         string    `gorm:"column:titulo" json:"titulo"`
	Cuerpo         string    `gorm:"column:cuerpo" json:"mensaje"` // API espera "mensaje", BD tiene "cuerpo"
	Leido          bool      `gorm:"column:leido" json:"leida"`
	Fecha          time.Time `gorm:"column:fecha" json:"fecha"`
}

func (Notificacion) TableName() string {
	return "\"Notificacion\""
}

// Funcionario mapea a la tabla "Funcionario"
type Funcionario struct {
	IDFuncionario   int     `gorm:"primaryKey;column:id_funcionario" json:"id_funcionario"`
	RutFuncionario  string  `gorm:"column:rut_funcionario" json:"rut_funcionario"`
	IDUsuario       int     `gorm:"column:id_usuario" json:"id_usuario"`
	Nombres         string  `gorm:"column:nombres" json:"nombres"`
	ApellidoPaterno string  `gorm:"column:apellido_paterno" json:"apellido_paterno"`
	ApellidoMaterno string  `gorm:"column:apellido_materno" json:"apellido_materno"`
	Email           string  `gorm:"column:email" json:"email"`
	Celular         *string `gorm:"column:celular" json:"celular"`
	IdSucursal      int     `gorm:"column:id_sucursal" json:"id_sucursal"`
	IdCargo         int     `gorm:"column:id_cargo" json:"id_cargo"`

	// Relaciones
	Sucursal Sucursal `gorm:"foreignKey:IdSucursal" json:"sucursal"`
	Cargo    Cargo    `gorm:"foreignKey:IdCargo" json:"cargo"`
}

// NombreCompleto retorna el nombre completo del funcionario
func (f Funcionario) NombreCompleto() string {
	return f.Nombres + " " + f.ApellidoPaterno + " " + f.ApellidoMaterno
}

func (Funcionario) TableName() string {
	return "\"Funcionario\""
}

// Estado mapea a la tabla "Estado"
type Estado struct {
	IDEstado     int    `gorm:"primaryKey;column:id_estado" json:"id_estado"`
	NombreEstado string `gorm:"column:nombre_estado" json:"nombre_estado"`
	TablaEstado  string `gorm:"column:tabla_estado" json:"tabla_estado"`
}

func (Estado) TableName() string {
	return "\"Estado\""
}

// Prenda mapea a la tabla "Prenda"
type Prenda struct {
	IDPrenda     int    `gorm:"primaryKey;column:id_prenda" json:"id_prenda"`
	NombrePrenda string `gorm:"column:nombre_prenda" json:"nombre_prenda"`
	IdTipoPrenda int    `gorm:"column:id_tipo_prenda" json:"id_tipo_prenda"`
	IdGenero     int    `gorm:"column:id_genero" json:"id_genero"`

	// Relaciones
	TipoPrenda TipoPrenda `gorm:"foreignKey:IdTipoPrenda" json:"tipo_prenda"`
	Genero     Genero     `gorm:"foreignKey:IdGenero" json:"genero"`
}

func (Prenda) TableName() string {
	return "\"Prenda\""
}

// TipoPeticion mapea a la tabla "Tipo Petición"
type TipoPeticion struct {
	IdTipoPeticion     int    `gorm:"primaryKey;column:id_tipo_peticion" json:"id_tipo_peticion"`
	NombreTipoPeticion string `gorm:"column:nombre_tipo_peticion" json:"nombre_tipo_peticion"`
}

func (TipoPeticion) TableName() string {
	return "\"Tipo Petición\""
}

// Uniforme mapea a la tabla "Uniforme"
type Uniforme struct {
	IdUniforme     int     `gorm:"primaryKey;column:id_uniforme" json:"id_uniforme"`
	NombreUniforme string  `gorm:"column:nombre_uniforme" json:"nombre_uniforme"`
	Descripcion    *string `gorm:"column:descripcion" json:"descripcion"`
}

func (Uniforme) TableName() string {
	return "\"Uniforme\""
}

// Temporada mapea a la tabla "Temporada"
type Temporada struct {
	IdTemporada     int        `gorm:"primaryKey;column:id_temporada" json:"id_temporada"`
	NombreTemporada string     `gorm:"column:nombre_temporada" json:"nombre_temporada"`
	FechaInicio     *time.Time `gorm:"column:fecha_inicio" json:"fecha_inicio"`
	FechaFin        *time.Time `gorm:"column:fecha_fin" json:"fecha_fin"`
}

func (Temporada) TableName() string {
	return "\"Temporada\""
}

// Sucursal mapea a la tabla "Sucursal"
type Sucursal struct {
	IdSucursal     int     `gorm:"primaryKey;column:id_sucursal" json:"id_sucursal"`
	NombreSucursal string  `gorm:"column:nombre_sucursal" json:"nombre_sucursal"`
	Direccion      *string `gorm:"column:direccion" json:"direccion"`
}

func (Sucursal) TableName() string {
	return "\"Sucursal\""
}

// Cargo mapea a la tabla "Cargo"
type Cargo struct {
	IdCargo     int    `gorm:"primaryKey;column:id_cargo" json:"id_cargo"`
	NombreCargo string `gorm:"column:nombre_cargo" json:"nombre_cargo"`
}

func (Cargo) TableName() string {
	return "\"Cargo\""
}

// Factura mapea a la tabla "Factura"
type Factura struct {
	IdFactura     int       `gorm:"primaryKey;column:id_factura" json:"id_factura"`
	NumeroFactura string    `gorm:"column:numero_factura;unique" json:"numero_factura"`
	FechaEmision  time.Time `gorm:"column:fecha_emision" json:"fecha_emision"`
	MontoTotal    float64   `gorm:"column:monto_total" json:"monto_total"`
	IdEstado      int       `gorm:"column:id_estado" json:"id_estado"`
	IdEmpresa     int       `gorm:"column:id_empresa" json:"id_empresa"`

	// Relaciones
	Estado  Estado  `gorm:"foreignKey:IdEstado" json:"estado"`
	Empresa Empresa `gorm:"foreignKey:IdEmpresa" json:"empresa"`
}

func (Factura) TableName() string {
	return "\"Factura\""
}

// Empresa mapea a la tabla "Empresa"
type Empresa struct {
	IdEmpresa     int     `gorm:"primaryKey;column:id_empresa" json:"id_empresa"`
	NombreEmpresa string  `gorm:"column:nombre_empresa" json:"nombre_empresa"`
	RutEmpresa    *string `gorm:"column:rut_empresa" json:"rut_empresa"`
}

func (Empresa) TableName() string {
	return "\"Empresa\""
}

// TipoPrenda mapea a la tabla "Tipo Prenda"
type TipoPrenda struct {
	IdTipoPrenda     int    `gorm:"primaryKey;column:id_tipo_prenda" json:"id_tipo_prenda"`
	NombreTipoPrenda string `gorm:"column:nombre_tipo_prenda" json:"nombre_tipo_prenda"`
}

func (TipoPrenda) TableName() string {
	return "\"Tipo Prenda\""
}

// Genero mapea a la tabla "Genero"
type Genero struct {
	IdGenero     int    `gorm:"primaryKey;column:id_genero" json:"id_genero"`
	NombreGenero string `gorm:"column:nombre_genero" json:"nombre_genero"`
}

func (Genero) TableName() string {
	return "\"Genero\""
}
