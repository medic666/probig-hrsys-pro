package database

import (
	"log"

	"golang.org/x/crypto/bcrypt"

	"probig/internal/models"
)

func Seed() {
	seedRBAC()
	seedAdminUser()
}

func seedRBAC() {
	modules := []string{"personnel", "organization", "attendance", "salary", "file", "audit", "settings"}
	actions := []string{"read", "write", "delete"}

	for _, mod := range modules {
		for _, act := range actions {
			perm := models.Permission{Module: mod, Action: act}
			DB.Where("module = ? AND action = ?", mod, act).FirstOrCreate(&perm)
		}
	}

	adminRole := models.Role{Name: "admin", Description: "系统管理员"}
	DB.Where("name = ?", "admin").FirstOrCreate(&adminRole)

	var allPerms []models.Permission
	DB.Find(&allPerms)
	DB.Model(&adminRole).Association("Permissions").Replace(allPerms)
}

func seedAdminUser() {
	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	adminUser := models.User{
		Username: "admin",
		Password: string(hash),
		RealName: "系统管理员",
		Status:   "active",
	}
	if err := DB.Create(&adminUser).Error; err != nil {
		log.Fatalf("failed to create admin user: %v", err)
	}

	var adminRole models.Role
	DB.Where("name = ?", "admin").First(&adminRole)
	DB.Model(&adminUser).Association("Roles").Append(&adminRole)

	log.Println("Seeded admin user: admin / admin123")
}
