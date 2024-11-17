type TransactionStatus string

const (
	TransactionPending   TransactionStatus = "pending"
	TransactionCompleted TransactionStatus = "completed"
	TransactionCancelled TransactionStatus = "cancelled"
	TransactionRefunded  TransactionStatus = "refunded"
)

type AffiliateTransaction struct {
	ID              uint              `gorm:"primaryKey"`
	LinkID          uint              `gorm:"not null"`
	UserID          uint              `gorm:"not null"`
	PartnerID       uint              `gorm:"not null"`
	OrderID         string            `gorm:"not null"` // Partner'ın sipariş ID'si
	Amount          float64           `gorm:"not null"` // Sipariş tutarı
	Commission      float64           `gorm:"not null"` // Kazanılan komisyon
	Status          TransactionStatus `gorm:"not null;default:'pending'"`
	TransactionDate time.Time         `gorm:"not null"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`

	Link    AffiliateLink    `gorm:"foreignkey:LinkID"`
	User    User             `gorm:"foreignkey:UserID"`
	Partner AffiliatePartner `gorm:"foreignkey:PartnerID"`
}