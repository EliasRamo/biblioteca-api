
package handler

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"biblioteca-api/internal/middleware"
)

// SetupRoutes registra todas las rutas de la API sobre el router de Gin.
func SetupRoutes(
	router *gin.Engine,
	authHandler *AuthHandler,
	bookHandler *BookHandler,
	loanHandler *LoanHandler,
	laptopHandler *LaptopHandler,
	jwtSecret string,
) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	api := router.Group("/api")

	// Rutas públicas de autenticación.
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Rutas para usuarios autenticados.
	protected := api.Group("")
	protected.Use(middleware.AuthRequired(jwtSecret))
	{
		protected.GET("/auth/me", authHandler.Me)

		// Libros.
		protected.GET("/books", bookHandler.List)
		protected.GET("/books/:id", bookHandler.Get)

		// Préstamos de libros.
		protected.POST("/loans", loanHandler.Borrow)
		protected.GET("/loans/me", loanHandler.MyLoans)
		protected.PUT("/loans/:id/return", loanHandler.Return)

		// Laptops disponibles.
		protected.GET("/laptops", laptopHandler.ListAll)
	}

	// Rutas exclusivas para estudiantes.
	student := api.Group("")
	student.Use(middleware.AuthRequired(jwtSecret), middleware.StudentRequired())
	{
		student.POST("/laptop-loans", laptopHandler.Borrow)
		student.GET("/laptop-loans/me", laptopHandler.MyLoans)
		student.PUT("/laptop-loans/:id/return", laptopHandler.Return)
	}

	// Rutas exclusivas para administradores.
	admin := api.Group("")
	admin.Use(middleware.AuthRequired(jwtSecret), middleware.AdminRequired())
	{
		// Administración de libros.
		admin.POST("/books", bookHandler.Create)
		admin.PUT("/books/:id", bookHandler.Update)
		admin.DELETE("/books/:id", bookHandler.Delete)
		admin.GET("/loans", loanHandler.AllLoans)

		// Administración de laptops.
		admin.POST("/laptops", laptopHandler.Create)
		admin.GET("/laptops/all", laptopHandler.ListAll)
		admin.GET("/laptop-loans", laptopHandler.AllLoans)
	}
}
