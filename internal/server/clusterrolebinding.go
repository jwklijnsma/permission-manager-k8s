package server

import (
	"github.com/labstack/echo/v4"
	rbacv1 "k8s.io/api/rbac/v1"
)

func createClusterRolebinding(c echo.Context) error {
	ac := c.(*AppContext)

	type Request struct {
		ClusterRolebindingName string `json:"clusterRolebindingName" validate:"required"`
		// GeneratedForUser is the field sent by the web-client (same name as in create-rolebinding)
		GeneratedForUser string `json:"generated_for_user"`
		// Username is kept for backwards compatibility with older API clients
		Username string           `json:"user"`
		Subjects []rbacv1.Subject `json:"subjects" validate:"required"`
		RoleName string           `json:"roleName" validate:"required"`
	}
	r := new(Request)

	err := ac.validateAndBindRequest(r)

	if err != nil {
		return err
	}

	subjs, err := normalizeSubjects(r.Subjects, ac.Config.Cluster.Namespace, ac.Config.RBAC.AdditionalSubjectKinds)
	if err != nil {
		return ac.errorResponse(err.Error())
	}

	username := r.GeneratedForUser
	if username == "" {
		username = r.Username
	}

	_, err = ac.ResourceManager.ClusterRoleBindingCreate(r.ClusterRolebindingName, username, r.RoleName, subjs)

	if err != nil {
		return err
	}

	return ac.okResponse()
}

func deleteClusterRolebinding(c echo.Context) error {
	ac := c.(*AppContext)

	type Request struct {
		RolebindingName string `json:"rolebindingName" validate:"required"`
	}

	r := new(Request)

	err := ac.validateAndBindRequest(r)

	if err != nil {
		return err
	}

	err = ac.ResourceManager.ClusterRoleBindingDelete(r.RolebindingName)

	if err != nil {
		return err
	}

	return ac.okResponse()
}
