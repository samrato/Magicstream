package controllers
import (
	"github.com/gin-gonic/gin"
)

func GetMovie()gin.HandlerFunc{
     return func (c *gin.Context){
    c.JSON(200,gin.H{"message":"Hello juma"})
	 }
}