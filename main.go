package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)
func main(){
    router:=gin.Default()
	// one of the reuest 
	router.GET("/users",func(c *gin.Context) {
		c.String(200,"Hello magic stream")
	})
	//creating the server in a particular port 

	if err:=router.Run(":8080");err!= nil{
      fmt.Println("Failed to start server",err)
	}


    
}