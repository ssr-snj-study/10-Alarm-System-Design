package alarm

import (
	"api/config"
	"api/model"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

func SendMsg(c echo.Context) error {
	b := new(model.Message)
	if err := c.Bind(b); err != nil {
		data := map[string]interface{}{
			"message": err.Error(),
		}

		return c.JSON(http.StatusInternalServerError, data)
	}
	var err error
	// 인증 먼저 시도

	// cache에서 값을 먼저 찾는다.
	token := checkCacheUser(b)
	if token == "" {
		token, err = checkUser(b)
		if err != nil {
			data := map[string]interface{}{
				"message": err.Error(),
			}
			return c.JSON(http.StatusInternalServerError, data)
		}
		if token == "" {
			data := map[string]interface{}{
				"message": errors.New("user_id is not available"),
			}
			return c.JSON(http.StatusInternalServerError, data)
		}
		inCacheDeviceToken(b.UserId, token)
	}
	// 계정별 전송률 제한
	if getSendAmount(token) > 300 {
		data := map[string]interface{}{
			"message": errors.New("message send over 300, in one minute, wait please"),
		}
		return c.JSON(http.StatusInternalServerError, data)
	} else {
		fmt.Println("insert message queue")
		producer := config.KafkaProducer()
		b.DeviceToken = token
		msg, err := json.Marshal(b)
		if err != nil {
			data := map[string]interface{}{
				"message": errors.New("message is not correct"),
			}
			return c.JSON(http.StatusInternalServerError, data)
		}
		producer.ProduceMsg(string(msg))
		//}
		response := map[string]interface{}{}

		return c.JSON(http.StatusOK, response)
	}
}
