package routes

import(
	controller "golang-speakbackend/controllers"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.RouterGroup){
	// incomingRoutes.POST("/signup", controller.SignUp())
	// incomingRoutes.POST("/login", controller.Login())
	incomingRoutes.GET("/user/:user_id", controller.GetUser())
	incomingRoutes.GET("/users", controller.GetUsers())
	incomingRoutes.PUT("/user/:user_id", controller.UpdateUser())
	incomingRoutes.DELETE("/user/:user_id", controller.DeleteUser())
	incomingRoutes.POST("/user/linkToTherapist/:user_id", controller.LinkToTherapist())
	incomingRoutes.GET("/patients/:therapist_id", controller.GetPatients())
	incomingRoutes.POST("/user/uploadprofile/:user_id", controller.UploadProfile())
	// incomingRoutes.POST("/user/refreshtoken", controller.RefreshToken())
}