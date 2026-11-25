package server

import (
	"crypto/subtle"
	"io/fs"
	"log"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"sighupio/permission-manager/internal/config"
	"sighupio/permission-manager/internal/resources"
	"sighupio/permission-manager/static"

	"k8s.io/client-go/kubernetes"
)

// New returns the permission-manager HTTP server, using the in-cluster (or local kubeconfig) client.
func New(cfg config.Config) *echo.Echo {
	return NewWithClient(cfg, resources.NewKubeClient())
}

// NewWithClient returns the permission-manager HTTP server using the given Kubernetes client.
func NewWithClient(cfg config.Config, kubeclient kubernetes.Interface) *echo.Echo {
	e := echo.New()

	e.Validator = &CustomValidator{validator: validator.New()}

	addMiddlewareStack(e, cfg, kubeclient)

	addRoutes(e)

	return e
}

func addMiddlewareStack(e *echo.Echo, cfg config.Config, kubeclient kubernetes.Interface) {
	// enable cors in local development
	e.Use(middleware.CORS())

	if cfg.Auth.BasicAuthEnabled {
		e.Use(basicAuthMiddleware(cfg.Auth))
	}

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))

	fsys, err := fs.Sub(static.WebClient, "build")
	if err != nil {
		log.Fatal(err)
	}

	e.Group("/*", middleware.StaticWithConfig(middleware.StaticConfig{
		Root:       ".",
		Filesystem: http.FS(fsys),
		HTML5:      true,
	}))

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			customContext := &AppContext{
				Context:         c,
				ResourceManager: resources.NewManager(kubeclient, c.Request().Context()),
				Config:          cfg,
			}
			return next(customContext)
		}
	})

}

// basicAuthMiddleware protects the UI and the API with HTTP basic auth.
// Credentials are compared in constant time.
func basicAuthMiddleware(auth config.AuthConfig) echo.MiddlewareFunc {
	return middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		userOK := subtle.ConstantTimeCompare([]byte(username), []byte(auth.Username)) == 1
		passOK := subtle.ConstantTimeCompare([]byte(password), []byte(auth.Password)) == 1
		return userOK && passOK, nil
	})
}

func addRoutes(e *echo.Echo) {
	api := e.Group("/api")

	api.GET("/list-users", listUsers)
	api.GET("/list-namespace", ListNamespaces)
	api.GET("/rbac", listRbac)

	api.POST("/create-cluster-role", createClusterRole)
	api.POST("/create-user", createUser)
	api.POST("/create-rolebinding", createRoleBinding)
	api.POST("/create-cluster-rolebinding", createClusterRolebinding)

	/* should use DELETE method, using POST due to a weird bug that looks now resolved */
	api.POST("/delete-cluster-role", deleteClusterRole)
	api.POST("/delete-cluster-rolebinding", deleteClusterRolebinding)
	api.POST("/delete-rolebinding", deleteRolebinding)
	api.POST("/delete-role", deleteRole)
	api.POST("/delete-user", deleteUser)

	api.POST("/create-kubeconfig", createKubeconfig)
	api.POST("/check-legacy-user", checkLegacyUser)
}
