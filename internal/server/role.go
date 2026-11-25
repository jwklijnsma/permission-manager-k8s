package server

import (
	"sighupio/permission-manager/internal/resources"

	"github.com/labstack/echo/v4"
	rbacv1 "k8s.io/api/rbac/v1"
)

func deleteRole(c echo.Context) error {
	ac := c.(*AppContext)

	type Request struct {
		RoleName  string `json:"roleName" validate:"required"`
		Namespace string `json:"namespace" validate:"required"`
	}

	r := new(Request)

	err := ac.validateAndBindRequest(r)

	if err != nil {
		return err
	}

	err = ac.ResourceManager.RoleDelete(r.Namespace, r.RoleName)

	if err != nil {
		return err
	}

	return ac.okResponse()
}

func deleteRolebinding(c echo.Context) error {
	ac := c.(*AppContext)

	type Request struct {
		RolebindingName string `json:"rolebindingName" validate:"required"`
		Namespace       string `json:"namespace" validate:"required"`
	}

	r := new(Request)

	err := ac.validateAndBindRequest(r)

	if err != nil {
		return err
	}

	err = ac.ResourceManager.RoleBindingDelete(r.Namespace, r.RolebindingName)

	if err != nil {
		return err
	}

	return ac.okResponse()
}

func createRoleBinding(c echo.Context) error {
	ac := c.(*AppContext)

	type Request struct {
		RolebindingName string           `json:"rolebindingName" validate:"required"`
		Namespace       string           `json:"namespace" validate:"required"`
		Username        string           `json:"generated_for_user" validate:"required"`
		Subjects        []rbacv1.Subject `json:"subjects" validate:"required"`
		RoleKind        string           `json:"roleKind" validate:"required"`
		RoleName        string           `json:"roleName" validate:"required"`
	}
	r := new(Request)

	err := ac.validateAndBindRequest(r)

	if err != nil {
		return err
	}

	if !isValidRoleKind(r.RoleKind) {
		return ac.errorResponse("roleKind must be ClusterRole or Role")
	}

	subjs, err := normalizeSubjects(r.Subjects, ac.Config.Cluster.Namespace, ac.Config.RBAC.AdditionalSubjectKinds)
	if err != nil {
		return ac.errorResponse(err.Error())
	}

	_, err = ac.ResourceManager.RoleBindingCreate(r.Namespace, r.Username, resources.RoleBindingRequirements{
		RoleKind:        r.RoleKind,
		RoleName:        r.RoleName,
		RolebindingName: r.RolebindingName,
		Subjects:        subjs,
	})

	if err != nil {
		return err
	}

	return ac.okResponse()
}
