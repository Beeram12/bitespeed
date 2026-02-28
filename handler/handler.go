package httpapi

import (
	"net/http"

	"github.com/Beeram12/bitespeed-identify/contact"
	"github.com/Beeram12/bitespeed-identify/db"
	"github.com/Beeram12/bitespeed-identify/service"
	"github.com/gin-gonic/gin"
)

type identifyRequest struct {
	Email       *string `json:"email"`
	PhoneNumber *string `json:"phoneNumber"`
}

type identifyResponse struct {
	Contact identifyContactResponse `json:"contact"`
}

type identifyContactResponse struct {
	PrimaryContactID    int64    `json:"primaryContatctId"`
	Emails              []string `json:"emails"`
	PhoneNumbers        []string `json:"phoneNumbers"`
	SecondaryContactIDs []int64  `json:"secondaryContactIds"`
}

func RegisterRoutes(r *gin.Engine, store *db.Store) {
	contactRepo := contact.NewRepository(store.DB)
	svc := service.NewIdentifyService(store.DB, contactRepo)

	r.POST("/identify", func(c *gin.Context) {
		var req identifyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
			return
		}

		if req.Email == nil && req.PhoneNumber == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "either email or phoneNumber is required"})
			return
		}

		result, err := svc.Identify(c.Request.Context(), service.IdentifyRequest{
			Email:       req.Email,
			PhoneNumber: req.PhoneNumber,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := identifyResponse{
			Contact: identifyContactResponse{
				PrimaryContactID:    result.PrimaryContactID,
				Emails:              result.Emails,
				PhoneNumbers:        result.PhoneNumbers,
				SecondaryContactIDs: result.SecondaryContactIDs,
			},
		}

		c.JSON(http.StatusOK, resp)
	})
}
