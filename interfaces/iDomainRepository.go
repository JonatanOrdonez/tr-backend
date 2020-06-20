package interfaces

import "github.com/JonatanOrdonez/tr-backend/models"

// IDomainRepository...
type IDomainRepository interface {
	FindByID(ID int64) (*models.Domain, error)
	GetAll() ([]*models.Domain, error)
	Save(domain *models.Domain) (int64, error)
	Update(domain *models.Domain) (int64, error)
	Delete(ID int) error
	FindByUrl(url string) (*models.Domain, error)
}
