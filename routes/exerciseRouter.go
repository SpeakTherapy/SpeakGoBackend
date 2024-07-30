package routes

import(
	controller "golang-speakbackend/controllers"

	"github.com/gin-gonic/gin"
)

func ExerciseRoutes(incomingRoutes *gin.RouterGroup){
	incomingRoutes.POST("/exercise", controller.CreateExercise())
	incomingRoutes.GET("/exercise/:exercise_id", controller.GetExercise())
	incomingRoutes.GET("/exercises", controller.GetExercises())
	incomingRoutes.PUT("/exercise/:exercise_id", controller.UpdateExercise())
	incomingRoutes.DELETE("/exercise/:exercise_id", controller.DeleteExercise())
}