package v1

import (
	"net/http"

	"github.com/bambu-group-03/apigw/config"
	"github.com/bambu-group-03/apigw/internal/entity"
	"github.com/bambu-group-03/apigw/pkg/logger"
	"github.com/gin-gonic/gin"
)

type gatewayRoutes struct {
	l   logger.Interface
	cfg *config.Config
}

func newGatewayRoutes(handler *gin.RouterGroup, l logger.Interface, cfg *config.Config) {
	r := &gatewayRoutes{l, cfg}

	h := handler.Group("/gateway")
	{
		h.GET("/history", r.history)
		h.POST("/route/*any", r.doRoute)
	}
}

type gatewayResponse struct {
	Gateway entity.Gateway `json:"gateway"`
}

// @Summary     Show history
// @Description Show all gateway history
// @ID          gateway-history
// @Tags  	    gateway
// @Accept      json
// @Produce     json
// @Success     200 {object} gatewayResponse
// @Failure     500 {object} response
// @Router      /gateway/history [get]
func (r *gatewayRoutes) history(c *gin.Context) {
	gateway := entity.Gateway{
		Source:      "auto",
		Destination: "es",
		Original:    "it's hardcoded, believe is not from memory",
		Translation: r.cfg.SERVICE.IDENTITY_SOCIALIZER_URL}

	c.JSON(http.StatusOK, gatewayResponse{gateway})
}

type doRouteRequest struct {
	Source      string `json:"source"       binding:"required"  example:"auto"`
	Destination string `json:"destination"  binding:"required"  example:"en"`
	Original    string `json:"original"     binding:"required"  example:"текст для перевода"`
}

// @Summary     Route
// @Description Route a text
// @ID          do-route
// @Tags  	    gateway
// @Accept      json
// @Produce     json
// @Param       request body doRouteRequest true "Set up gateway"
// @Success     200 {object} entity.Gateway
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /gateway/route [post]
func (r *gatewayRoutes) doRoute(c *gin.Context) {
	var request doRouteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		r.l.Error(err, "http - v1 - doRoute")
		errorResponse(c, http.StatusBadRequest, "invalid request body")

		return
	}
	gateway := entity.Gateway{
		Source:      request.Source,
		Destination: request.Destination,
		Original:    request.Original}
	// if err != nil {
	// 	r.l.Error(err, "http - v1 - doRoute")
	// 	errorResponse(c, http.StatusInternalServerError, "gateway service problems")

	// 	return
	// }
	c.JSON(http.StatusOK, gateway)
}
