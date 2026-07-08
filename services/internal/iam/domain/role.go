package domain

type Permission string

const (
	CatalogRead      Permission = "catalog.read"
	CatalogCreate    Permission = "catalog.create"
	CatalogUpdate    Permission = "catalog.update"
	CatalogDelete    Permission = "catalog.delete"
	SalesCreate      Permission = "sales.create"
	SalesRead        Permission = "sales.read"
	SalesUpdate      Permission = "sales.update"
	SalesDelete      Permission = "sales.delete"
	SalesRefund      Permission = "sales.refund"
	InventoryRead    Permission = "inventory.read"
	InventoryWrite   Permission = "inventory.write"
	InventoryAdjust  Permission = "inventory.adjust"
	MerchantManage   Permission = "merchant.manage"
	UserCreate       Permission = "user.create"
	UserRead         Permission = "user.read"
	UserUpdate       Permission = "user.update"
	UserDelete       Permission = "user.delete"
	UserList         Permission = "user.list"
	RoleCreate       Permission = "role.create"
	RoleRead         Permission = "role.read"
	RoleUpdate       Permission = "role.update"
	RoleDelete       Permission = "role.delete"
	RoleAssign       Permission = "role.assign"
	PermissionCreate Permission = "permission.create"
	PermissionRead   Permission = "permission.read"
	PermissionUpdate Permission = "permission.update"
	PermissionDelete Permission = "permission.delete"
	PermissionAssign Permission = "permission.assign"
	SessionManage    Permission = "session.manage"
	APIKeyManage     Permission = "apikey.manage"
	AuditView        Permission = "audit.view"
)

type Role struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
	IsSystem    bool         `json:"is_system"`
}

var DefaultRoles = map[string]Role{
	"superadmin": {
		Name:        "superadmin",
		Description: "Platform super administrator — full access to all merchants",
		IsSystem:    true,
		Permissions: []Permission{
			CatalogRead, CatalogCreate, CatalogUpdate, CatalogDelete,
			SalesCreate, SalesRead, SalesUpdate, SalesDelete, SalesRefund,
			InventoryRead, InventoryWrite, InventoryAdjust,
			MerchantManage,
			UserCreate, UserRead, UserUpdate, UserDelete, UserList,
			RoleCreate, RoleRead, RoleUpdate, RoleDelete, RoleAssign,
			PermissionCreate, PermissionRead, PermissionUpdate, PermissionDelete, PermissionAssign,
			SessionManage, APIKeyManage, AuditView,
		},
	},
	"owner": {
		Name:        "owner",
		Description: "Full system access",
		IsSystem:    true,
		Permissions: []Permission{
			CatalogRead, CatalogCreate, CatalogUpdate, CatalogDelete,
			SalesCreate, SalesRead, SalesUpdate, SalesDelete, SalesRefund,
			InventoryRead, InventoryWrite, InventoryAdjust,
			MerchantManage,
			UserCreate, UserRead, UserUpdate, UserDelete, UserList,
			RoleCreate, RoleRead, RoleUpdate, RoleDelete, RoleAssign,
			PermissionCreate, PermissionRead, PermissionUpdate, PermissionDelete, PermissionAssign,
			SessionManage, APIKeyManage, AuditView,
		},
	},
	"admin": {
		Name:        "admin",
		Description: "Administrative access",
		IsSystem:    true,
		Permissions: []Permission{
			CatalogRead, CatalogCreate, CatalogUpdate, CatalogDelete,
			SalesCreate, SalesRead, SalesUpdate, SalesDelete, SalesRefund,
			InventoryRead, InventoryWrite, InventoryAdjust,
			MerchantManage,
			UserCreate, UserRead, UserUpdate, UserDelete, UserList,
			RoleCreate, RoleRead, RoleUpdate, RoleDelete, RoleAssign,
			PermissionCreate, PermissionRead, PermissionUpdate, PermissionDelete, PermissionAssign,
			SessionManage, APIKeyManage, AuditView,
		},
	},
	"supervisor": {
		Name:        "supervisor",
		Description: "Supervisory access",
		IsSystem:    true,
		Permissions: []Permission{
			CatalogRead, CatalogCreate, CatalogUpdate,
			SalesCreate, SalesRead, SalesUpdate, SalesRefund,
			InventoryRead, InventoryWrite, InventoryAdjust,
			UserRead, UserList,
		},
	},
	"cashier": {
		Name:        "cashier",
		Description: "Cashier access",
		IsSystem:    true,
		Permissions: []Permission{
			CatalogRead,
			SalesCreate, SalesRead,
			InventoryRead,
		},
	},
	"viewer": {
		Name:        "viewer",
		Description: "Read-only access",
		IsSystem:    true,
		Permissions: []Permission{
			CatalogRead,
			SalesRead,
			InventoryRead,
			UserRead, UserList,
		},
	},
}
